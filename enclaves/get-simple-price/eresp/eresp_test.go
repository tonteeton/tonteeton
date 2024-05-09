package eresp

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

func TestEnclaveResponse(t *testing.T) {
	secretKey, err := base64.StdEncoding.DecodeString(
		"yMJNiUZf3kMeEkQ+0r57+Ou8DEfOKmNC/BCN9c2TfPc5PICixeaQ8vlV/79OARLthRMyTOXEVDU16/1JY3BP1Q==",
	)
	if err != nil {
		t.Errorf("Error decoding secret key: %v", err)
	}
	price := EnclavePrice{
		USD:           345,
		LastUpdatedAt: 1715092161,
	}

	t.Run("ValidInputs", func(t *testing.T) {
		got, err := NewEnclaveResponse(price, secretKey)
		if err != nil {
			t.Errorf("Error building Enclave response: %v", err)
		}

		expectedPayload := "te6cckEBAQEAEgAAIAAAAAAAAAFZAAAAAGY6OsEspGH5"
		if got.Payload != expectedPayload {
			t.Errorf("Unexpected payload. Got: %s, Expected: %s", got.Payload, expectedPayload)
		}

		expectedHash := "5MMpsekXDDQw6yiY4HOu6mEV2k/YiRz5GbB6kLSgVIA="
		if got.Hash != expectedHash {
			t.Errorf("Unexpected hash. Got: %s, Expected: %s", got.Hash, expectedHash)
		}

		expectedSignature := "QxBExxLU/NQMysbB6t3sevQBdbXgl2zo//V9yWkrkXWOQQXEVTnK45cYC/O6X17NKt2FtlWzjEchxQdOPnAFDA=="
		if got.Signature != expectedSignature {
			t.Errorf("Unexpected signature. Got: %s, Expected: %s", got.Signature, expectedSignature)
		}
	})

	t.Run("InvalidSecretKey", func(t *testing.T) {
		_, err := NewEnclaveResponse(price, []byte("invalidsecretkey"))
		if err != nil {
			t.Error("Error expected for invalid secret key")
		}
	})

}

func TestSaveEnclaveResponseToJson(t *testing.T) {
	os.Chdir(t.TempDir())

	response := EnclaveResponse{
		Signature: "signature",
		Payload:   "payload",
		Hash:      "hash",
	}

	err := response.Save()
	if err != nil {
		t.Errorf("Error saving EnclaveResponse to JSON: %v", err)
	}

	data, _ := ioutil.ReadFile(EnclaveResponsePath)
	var savedResponse EnclaveResponse
	err = json.Unmarshal(data, &savedResponse)
	if err != nil {
		t.Errorf("Error unmarshaling saved JSON data: %v", err)
	}
}

