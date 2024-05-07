package enclave

import (
	"encoding/base64"
	"testing"
)

func TestOracleResponse(t *testing.T) {
	t.Helper()

	secretKey, err := base64.StdEncoding.DecodeString(
		"yMJNiUZf3kMeEkQ+0r57+Ou8DEfOKmNC/BCN9c2TfPc5PICixeaQ8vlV/79OARLthRMyTOXEVDU16/1JY3BP1Q==",
	)
	if err != nil {
		t.Errorf("Error decoding secret key: %v", err)
	}

	t.Run("ValidInputs", func(t *testing.T) {
		got, err := buildOracleResponse(345, 1715092161, secretKey)
		if err != nil {
			t.Errorf("Error building Oracle response: %v", err)
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
		_, err := buildOracleResponse(345, 1715092161, []byte("invalidsecretkey"))
		if err != nil {
			t.Error("Error expected for invalid secret key")
		}
	})

}
