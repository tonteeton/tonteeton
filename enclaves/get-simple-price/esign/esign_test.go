package esign

import (
	"crypto/ed25519"
	"enclave/econf"
	"os"
	"testing"
)

func setupTest(t *testing.T, config econf.KeysConfig) func() {

	os.Chdir(t.TempDir())

	if _, err := os.Stat(config.PrivateKeyPath); err == nil {
		err = os.Remove(config.PrivateKeyPath)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	}

	return func() {

	}
}

func TestSignatureKey(t *testing.T) {
	config := econf.KeysConfig{
		PublicKeyPath:  "key.pub",
		PrivateKeyPath: "key.priv.enc",
		SealedDatePath: "created.enc",
		Version:        "test",
	}

	t.Run("Keys loaded correctly", func(t *testing.T) {
		defer setupTest(t, config)()
		_, err := GetSignatureKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	})

	t.Run("Key loaded and reused", func(t *testing.T) {
		defer setupTest(t, config)()

		key1, err := GetSignatureKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if _, err = os.Stat(config.PrivateKeyPath); err != nil {
			t.Fatalf("Key was not saved")
		}

		key2, err := GetSignatureKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if !key1.privateKey.Equal(key2.privateKey) {
			t.Fatalf("Key was not reused")
		}

		err = os.Remove(config.PrivateKeyPath)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		key3, err := GetSignatureKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if key1.privateKey.Equal(key3.privateKey) {
			t.Fatalf("Private keys always same, %x %x", key1.privateKey, key3.privateKey)
		}
	})

	t.Run("Public key returned", func(t *testing.T) {
		defer setupTest(t, config)()
		key, err := GetSignatureKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		publicKey := key.GetPublicKey()
		if len(publicKey) != ed25519.PublicKeySize {
			t.Fatalf("Key of unexpected length")
		}
	})
}
