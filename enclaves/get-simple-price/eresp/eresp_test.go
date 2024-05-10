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
		LastUpdatedAt: 1715092161,
		Ticker:        0x72716023,
		USD:           345,
		USD24HVol:     81968225604,
		USD24HChange:  1566,
		BTC:           10967,
	}

	t.Run("ValidInputs", func(t *testing.T) {
		got, err := NewEnclaveResponse(price, secretKey)
		if err != nil {
			t.Errorf("Error building Enclave response: %v", err)
		}

		expectedPayload := "te6cckEBAQEAMgAAYAAAAABmOjrBAAAAAHJxYCMAAAAAAAABWQAAABMVr91EAAAAAAAABh4AAAAAAAAq11siUa4="
		if got.Payload != expectedPayload {
			t.Errorf("Unexpected payload. Got: %s, Expected: %s", got.Payload, expectedPayload)
		}

		expectedHash := "KWraQp7R+lYAaGw9VqJnMeKcar9q+mKtudCST/4h3GY="
		if got.Hash != expectedHash {
			t.Errorf("Unexpected hash. Got: %s, Expected: %s", got.Hash, expectedHash)
		}

		expectedSignature := "Id3NO8Tbq4ZFcZ1mp4gr78g7+SgmHuCdTSSBXmzXYy7u3W/UPisnTsE7CuDUATiaOFnE208w1fyb8+s6BM/0BA=="
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
