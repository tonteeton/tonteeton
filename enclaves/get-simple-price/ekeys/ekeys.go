// Package ekeys provides functions for managing Ed25519 keys.
package ekeys

import (
	"crypto/ed25519"
	"enclave/econf"
	"encoding/base64"
	"errors"
	"github.com/edgelesssys/ego/ecrypto"
	"os"
	"time"
)

type KeyManager struct {
	config econf.KeysConfig
}

// GetPrivateKey retrieves the private from location specified by KeysConfig.
// If the private key file doesn't exist, it generates new keys and saves them.
// If the private key file exists, it loads and returns the existing private key.
func GetPrivateKey(config econf.KeysConfig) (ed25519.PrivateKey, error) {
	keys := KeyManager{config}
	if _, err := os.Stat(config.PrivateKeyPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := keys.createNewKeys(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return keys.loadKeys()
}

func (keys KeyManager) createNewKeys() error {
	publicKey, privateKey, err := generateKey()
	if err != nil {
		return errors.New("Failed to generate keys")
	}
	creationInfo := []byte(time.Now().Format(time.RFC3339))

	err = keys.writeEncryptedFile(keys.config.SealedDatePath, creationInfo)
	if err != nil {
		return errors.New("Failed to write creation info")
	}

	err = os.WriteFile(
		keys.config.PublicKeyPath,
		[]byte(base64.StdEncoding.EncodeToString(publicKey)),
		0600,
	)
	if err != nil {
		return errors.New("Failed to write public key")
	}

	err = keys.writeEncryptedFile(keys.config.PrivateKeyPath, privateKey)
	if err != nil {
		return errors.New("Failed to write private key")
	}

	return nil
}

func (keys KeyManager) loadKeys() (ed25519.PrivateKey, error) {
	var dateData []byte
	var creationDate time.Time
	dateData, err := keys.readEncryptedFile(keys.config.SealedDatePath)
	if err != nil {
		return nil, errors.New("Failed to read creation info")
	}
	creationDate, err = time.Parse(time.RFC3339, string(dateData))
	if err != nil {
		return nil, errors.New("Failed to parse creation date")
	}
	if time.Now().Before(creationDate) {
		return nil, errors.New("Unexpected keys creation date")
	}

	var privateKeyData []byte
	privateKeyData, err = keys.readEncryptedFile(keys.config.PrivateKeyPath)
	if err != nil {
		return nil, errors.New("Failed to read private key")
	}

	return ed25519.PrivateKey(privateKeyData), nil
}

func (keys KeyManager) readEncryptedFile(path string) ([]byte, error) {
	additionalData := []byte(keys.config.Version)

	sealedData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data []byte
	data, err = ecrypto.Unseal(sealedData, additionalData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (keys KeyManager) writeEncryptedFile(path string, data []byte) error {
	additionalData := []byte(keys.config.Version)
	encryptedData, err := ecrypto.SealWithUniqueKey(data, additionalData)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, encryptedData, 0600)
	if err != nil {
		return err
	}
	return err
}

func generateKey() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(nil)
}
