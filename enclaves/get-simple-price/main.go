package main

import (
	"enclave/coinconv"
	"enclave/coingecko"
	"enclave/econf"
	"enclave/ekeys"
	"enclave/eresp"
	"fmt"
)

func main() {
	cfg, err := econf.LoadConfig()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	gecko := coingecko.NewGecko(
		cfg.CoinGecko.DemoKey,
		cfg.CoinGecko.ProKey,
	)
	fmt.Printf("%+v\n", gecko)
	geckoPrice, err := gecko.GetTONPrice()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%+v\n", geckoPrice)

	var price eresp.EnclavePrice
	price = coinconv.ConvertPrice(geckoPrice.TON, cfg.Tickers.TON)
	if err := coinconv.ValidatePrice(price); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%+v", price)

	privateKey, err := ekeys.GetPrivateKey(cfg.Keys)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	result, err := eresp.NewEnclaveResponse(price, privateKey)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	err = result.Save()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
