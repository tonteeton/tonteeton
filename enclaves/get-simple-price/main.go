package main

import (
	"enclave/coinconv"
	"enclave/coingecko"
	"enclave/eresp"
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	gecko := coingecko.NewGecko(
		os.Getenv("COINGECKO_API_KEY"),
		os.Getenv("COINGECKO_PRO_API_KEY"),
	)
	fmt.Printf("%+v\n", gecko)
	geckoPrice, err := gecko.GetTONPrice()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%+v\n", geckoPrice)

	var price eresp.EnclavePrice
	price = coinconv.ConvertPrice(geckoPrice.TON)
	if err := coinconv.ValidatePrice(price); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%+v", price)

	secretKey, _ := base64.StdEncoding.DecodeString(
		"yMJNiUZf3kMeEkQ+0r57+Ou8DEfOKmNC/BCN9c2TfPc5PICixeaQ8vlV/79OARLthRMyTOXEVDU16/1JY3BP1Q==",
	)
	result, err := eresp.NewEnclaveResponse(price, secretKey)
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
