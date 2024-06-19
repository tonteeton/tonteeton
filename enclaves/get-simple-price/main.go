package main

import (
	"enclave/appconf"
	"enclave/coinconv"
	"enclave/coingecko"
	"enclave/priceresp"
	"errors"
	"flag"
	"fmt"
	"github.com/tonteeton/golib/eattest"
	"github.com/tonteeton/golib/ereport"
	"github.com/tonteeton/golib/eresp"
	"os"
)

func getPrice(cfg *appconf.Config) error {
	gecko := coingecko.NewGecko(
		cfg.CoinGecko.DemoKey,
		cfg.CoinGecko.ProKey,
	)

	geckoPrice, err := gecko.GetTONPrice()
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", geckoPrice)

	var price priceresp.Price
	price = coinconv.ConvertPrice(geckoPrice.TON, cfg.Tickers.TON)
	if err := coinconv.ValidatePrice(price); err != nil {
		return err
	}
	fmt.Printf("%+v\n", price)

	responseCfg := eresp.Config{
		Response:      cfg.Response,
		SignatureKeys: cfg.SignatureKeys,
	}
	return eresp.SaveResponse(responseCfg, price.ToCell())
}

func executeReportFunc(fn func(ereport.Config, eattest.Attestation) error, cfg *appconf.Config) error {
	reportCfg := ereport.Config{
		Reports:        cfg.Reports,
		SignatureKeys:  cfg.SignatureKeys,
		EncryptionKeys: cfg.EncryptionKeys,
	}
	attest := eattest.NewAttestation()
	return fn(reportCfg, attest)
}

func main() {
	cfg, err := appconf.LoadConfig()
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

	cmds := map[string]func(cfg *appconf.Config) error{
		"get-price":  getPrice,
		"report-key": func(cfg *appconf.Config) error { return executeReportFunc(ereport.ExportPublicKeys, cfg) },
		"import-key": func(cfg *appconf.Config) error { return executeReportFunc(ereport.ImportPrivateSignature, cfg) },
		"export-key": func(cfg *appconf.Config) error { return executeReportFunc(ereport.ExportPrivateSignature, cfg) },
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
