package ekeys

import (
	"crypto/ed25519"
	"enclave/econf"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config := econf.KeysConfig{
		PublicKeyPath:  "key.pub",
		PrivateKeyPath: "key.priv.enc",
		SealedDatePath: "created.enc",
		Version:        "test",
	}
	keys := KeyManager{config}
	os.Chdir(t.TempDir())

	t.Run("Write and read encrypted file", func(t *testing.T) {
		err := keys.writeEncryptedFile("test.enc", []byte("testdata"))
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		var data []byte
		data, err = ioutil.ReadFile("test.enc")
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if strings.Contains(string(data), "testdata") {
			t.Fatalf("Data was not encrypted: '%s'", data)
		}

		var decrypted []byte
		decrypted, err = keys.readEncryptedFile("test.enc")
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if string(decrypted) != "testdata" {
			t.Fatalf("Unexpected decrypted data: '%s'", decrypted)
		}

	})

	t.Run("Create and load keys", func(t *testing.T) {
		_, err := keys.loadKeys()
		if err == nil || string(err.Error()) != "Failed to read creation info" {
			t.Fatalf("Error on reading creation info expeted, got: %v", err)
		}

		err = keys.createNewKeys()
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		_, err = keys.loadKeys()
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	})

	t.Run("Key loaded and reused", func(t *testing.T) {
		var firstPrivKey, secondPrivKey ed25519.PrivateKey
		var err error
		if _, err = os.Stat(keys.config.PrivateKeyPath); errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Private key path is not empty")
		}

		firstPrivKey, err = GetPrivateKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if _, err = os.Stat(keys.config.PrivateKeyPath); err != nil {
			t.Fatalf("Key was not saved")
		}

		secondPrivKey, err = GetPrivateKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if !firstPrivKey.Equal(secondPrivKey) {
			t.Fatalf("Key was not reused")
		}
	})

	t.Run("Invalid Key not used", func(t *testing.T) {
		_, err := GetPrivateKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		os.WriteFile(keys.config.PrivateKeyPath, []byte("modified"), 0644)

		_, err = GetPrivateKey(config)
		if err == nil || !strings.Contains(err.Error(), "private key") {
			t.Fatalf("Expected 'unexpected keys creation date' error, got: %v", err)
		}
	})
}
