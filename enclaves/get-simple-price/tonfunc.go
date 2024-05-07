package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"io/ioutil"
)

const (
	oracleResponsePath = `mount/response.json`
)

type OracleResponse struct {
	Signature string `json:"signature"`
	Payload   string `json:"payload"`
	Hash      string `json:"hash"`
}

func buildOracleResponse(usdPrice uint64, lastUpdatedAt uint64, key ed25519.PrivateKey) (OracleResponse, error) {
	var payload *cell.Cell
	payload = cell.BeginCell().
		MustStoreUInt(usdPrice, 64).
		MustStoreUInt(lastUpdatedAt, 64).
		EndCell()
	hash := payload.Hash()
	if hash == nil {
		return OracleResponse{}, errors.New("Failed to compute payload hash")
	}

	var signature []byte
	defer func() {
		if r := recover(); r != nil {
			signature = nil
		}
	}()
	signature = payload.Sign(key)
	if signature == nil {
		return OracleResponse{}, errors.New("Failed to sign payload hash")
	}

	return OracleResponse{
		base64.StdEncoding.EncodeToString(signature),
		base64.StdEncoding.EncodeToString(payload.ToBOC()),
		base64.StdEncoding.EncodeToString(hash),
	}, nil
}

func saveOracleResponse(response OracleResponse) error {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(oracleResponsePath, data, 0600)
	if err != nil {
		return err
	}

	return nil
}
