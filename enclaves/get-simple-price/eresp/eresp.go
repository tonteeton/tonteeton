package eresp

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	EnclaveResponsePath = `mount/response.json`
)

type EnclavePrice struct {
	USD           uint64
	LastUpdatedAt uint64
}

type EnclaveResponse struct {
	Signature string `json:"signature"`
	Payload   string `json:"payload"`
	Hash      string `json:"hash"`
}

func NewEnclaveResponse(price EnclavePrice, key ed25519.PrivateKey) (EnclaveResponse, error) {
	var payload *cell.Cell
	payload = cell.BeginCell().
		MustStoreUInt(price.USD, 64).
		MustStoreUInt(price.LastUpdatedAt, 64).
		EndCell()
	hash := payload.Hash()
	if hash == nil {
		return EnclaveResponse{}, errors.New("Failed to compute payload hash")
	}

	var signature []byte
	defer func() {
		if r := recover(); r != nil {
			signature = nil
		}
	}()
	signature = payload.Sign(key)
	if signature == nil {
		return EnclaveResponse{}, errors.New("Failed to sign payload hash")
	}

	return EnclaveResponse{
		base64.StdEncoding.EncodeToString(signature),
		base64.StdEncoding.EncodeToString(payload.ToBOC()),
		base64.StdEncoding.EncodeToString(hash),
	}, nil
}

func (response EnclaveResponse) Save() error {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}
	responseDir := filepath.Dir(EnclaveResponsePath)
	err = os.MkdirAll(responseDir, 0700)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(EnclaveResponsePath, data, 0600)
	if err != nil {
		return err
	}

	return nil
}
