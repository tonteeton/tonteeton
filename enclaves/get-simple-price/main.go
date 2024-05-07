package main

import (
	"encoding/base64"
	"fmt"
)

func main() {
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
