package main

import (
	"enclave/coinconv"
	"enclave/coingecko"
	"enclave/eattest"
	"enclave/econf"
	"enclave/ereport"
	"enclave/eresp"
	"enclave/esign"
	"errors"
	"flag"
	"fmt"
	"os"
)

func getPrice(cfg *econf.Config) error {
	gecko := coingecko.NewGecko(
		cfg.CoinGecko.DemoKey,
		cfg.CoinGecko.ProKey,
	)

	geckoPrice, err := gecko.GetTONPrice()
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", geckoPrice)

	var price eresp.EnclavePrice
	price = coinconv.ConvertPrice(geckoPrice.TON, cfg.Tickers.TON)
	if err := coinconv.ValidatePrice(price); err != nil {
		return err
	}
	fmt.Printf("%+v\n", price)

	signature, err := esign.GetSignatureKey(cfg.SignatureKeys)
	if err != nil {
		return err
	}

	result, err := eresp.NewEnclaveResponse(price, signature.GetPrivateKey())
	if err != nil {
		return err
	}
	err = result.Save()
	if err != nil {
		return err
	}
	return nil
}

func executeReportFunc(fn func(ereport.Config, eattest.Attestation) error, cfg *econf.Config) error {
	reportCfg := ereport.Config{
		Reports:        cfg.Reports,
		SignatureKeys:  cfg.SignatureKeys,
		EncryptionKeys: cfg.EncryptionKeys,
	}
	attest := eattest.NewAttestation()
	return fn(reportCfg, attest)
}

func main() {
	cfg, err := econf.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	flag.Usage = func() {
		fmt.Println("Usage: [command]")
		fmt.Println("Commands:")
		fmt.Println("  get-price        Get the TON price")
		fmt.Println("  report-key       Generate SGX-signed report with public keys")
		fmt.Println("  import-key       Import encrypted signature Private key")
		fmt.Println("  export-key       Export encrypted signature Private key")
	}

	cmds := map[string]func(cfg *econf.Config) error{
		"get-price":  getPrice,
		"report-key": func(cfg *econf.Config) error { return executeReportFunc(ereport.ExportPublicKeys, cfg) },
		"import-key": func(cfg *econf.Config) error { return executeReportFunc(ereport.ImportPrivateSignature, cfg) },
		"export-key": func(cfg *econf.Config) error { return executeReportFunc(ereport.ExportPrivateSignature, cfg) },
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	cmd, ok := cmds[os.Args[1]]
	if ok {
		err = cmd(cfg)
	} else {
		flag.Usage()
		err = errors.New("Unknown command.")
	}

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
