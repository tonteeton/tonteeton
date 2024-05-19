package main

import (
	"enclave/coinconv"
	"enclave/coingecko"
	"enclave/econf"
	"enclave/eresp"
	"enclave/esign"
	"fmt"
	"os"
)

func run() error {
	cfg, err := econf.LoadConfig()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	gecko := coingecko.NewGecko(
		cfg.CoinGecko.DemoKey,
		cfg.CoinGecko.ProKey,
	)

	geckoPrice, err := gecko.GetTONPrice()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	fmt.Printf("%+v\n", geckoPrice)

	var price eresp.EnclavePrice
	price = coinconv.ConvertPrice(geckoPrice.TON, cfg.Tickers.TON)
	if err := coinconv.ValidatePrice(price); err != nil {
		fmt.Println("Error:", err)
		return err
	}
	fmt.Printf("%+v\n", price)

	signature, err := esign.GetSignatureKey(cfg.SignatureKeys)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	result, err := eresp.NewEnclaveResponse(price, signature.GetPrivateKey())
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	err = result.Save()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	return nil
}

func main() {
	exitCode := 0
	err := run()
	if err != nil {
		exitCode = 1
	}
	os.Exit(exitCode)
}
