package ebox

import (
	"bytes"
	"enclave/econf"
	"errors"
	"io"
	"os"
	"testing"
)

func setupTest(t *testing.T, config econf.KeysConfig) func() {

	os.Chdir(t.TempDir())

	OriginalGenerateKey := GenerateKey

	return func() {
		GenerateKey = OriginalGenerateKey
	}
}

func TestBoxKey(t *testing.T) {
	config := econf.KeysConfig{
		PublicKeyPath:  "key.pub",
		PrivateKeyPath: "key.priv.enc",
		SealedDatePath: "created.enc",
		Version:        "test",
	}

	var expectedPub [32]byte
	for i := range expectedPub {
		expectedPub[i] = 0xaa
	}

	var expectedPriv [32]byte
	for i := range expectedPriv {
		expectedPriv[i] = 0xbb
	}

	t.Run("Keys loaded correctly", func(t *testing.T) {
		defer setupTest(t, config)()

		GenerateKey = func(_ io.Reader) (*[32]byte, *[32]byte, error) {
			return &expectedPub, &expectedPriv, nil
		}

		key, err := GetBoxKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if key.publicKey != expectedPub {
			t.Fatalf("Expected: %#v\nGot: %#v", expectedPub, key.publicKey)
		}
		if key.privateKey != expectedPriv {
			t.Fatalf("Expected: %#v\nGot: %#v", expectedPub, key.publicKey)
		}
	})

	t.Run("Key loaded and reused", func(t *testing.T) {
		defer setupTest(t, config)()
		key1, err := GetBoxKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if _, err := os.Stat(config.PrivateKeyPath); errors.Is(err, os.ErrNotExist) {
			t.Fatalf("Private key was not saved")
		}

		key2, err := GetBoxKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if key1.publicKey != key2.publicKey {
			t.Fatalf("Public keys not reused, %x %x", key1.publicKey, key2.publicKey)
		}

		if key1.privateKey != key2.privateKey {
			t.Fatalf("Private keys not reused, %x %x", key1.privateKey, key2.privateKey)
		}

		err = os.Remove(config.PrivateKeyPath)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		key3, err := GetBoxKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		if key1.publicKey == key3.publicKey {
			t.Fatalf("Public keys always same, %x %x", key1.publicKey, key3.publicKey)
		}
		if key1.privateKey != key2.privateKey {
			t.Fatalf("Private keys always same, %x %x", key1.privateKey, key3.privateKey)
		}

	})

	t.Run("Keys generated", func(t *testing.T) {
		defer setupTest(t, config)()
		publicKey, privateKey, err := generateBoxKey()
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if len(publicKey) != 32 || len(privateKey) != 64 {
			t.Fatalf("Unexpected length of keys")
		}
	})

	t.Run("Public key returned", func(t *testing.T) {
		defer setupTest(t, config)()
		GenerateKey = func(_ io.Reader) (*[32]byte, *[32]byte, error) {
			return &expectedPub, &expectedPriv, nil
		}
		key, err := GetBoxKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		publicKey := key.GetPublicKey()
		if len(publicKey) != PublicKeySize {
			t.Fatalf("Key of unexpected length")
		}
		if !bytes.Equal(publicKey, expectedPub[:]) {
			t.Fatalf("Unexpected key data %v %v", publicKey, expectedPub[:])
		}
	})

	t.Run("Private key returned", func(t *testing.T) {
		defer setupTest(t, config)()
		GenerateKey = func(_ io.Reader) (*[32]byte, *[32]byte, error) {
			return &expectedPub, &expectedPriv, nil
		}
		key, err := GetBoxKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		privateKey := key.GetPrivateKey()
		if len(privateKey) != PrivateKeySize {
			t.Fatalf("Key of unexpected length")
		}
		if !bytes.Equal(privateKey, expectedPriv[:]) {
			t.Fatalf("Unexpected key data")
		}
	})

	t.Run("Data encrypted and decrypted", func(t *testing.T) {
		defer setupTest(t, config)()
		msg := []byte("test text")

		sender, err := GetBoxKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		err = os.Remove(config.PrivateKeyPath)
		recipient, err := GetBoxKey(config)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		encrypted, err := sender.Encrypt(msg, recipient.GetPublicKey())
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if bytes.Equal(encrypted, msg) {
			t.Fatalf("Data was not encrypted")
		}

		decrypted, err := recipient.Decrypt(encrypted, sender.GetPublicKey())
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if !bytes.Equal(decrypted, msg) {
			t.Fatalf("Data was not decrypted")
		}

		decrypted, err = sender.Decrypt(encrypted, sender.GetPublicKey())
		if err == nil {
			t.Fatalf("Data can be decryped with a wrong key")
		}
	})
}
