package main

import (
	"enclave/coingecko"
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
	prices, err := gecko.GetTONPrice()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%+v\n", prices)
	secretKey, _ := base64.StdEncoding.DecodeString(
		"yMJNiUZf3kMeEkQ+0r57+Ou8DEfOKmNC/BCN9c2TfPc5PICixeaQ8vlV/79OARLthRMyTOXEVDU16/1JY3BP1Q==",
	)
	response, err := buildOracleResponse(31415926, 1715116956, secretKey)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = saveOracleResponse(response)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
