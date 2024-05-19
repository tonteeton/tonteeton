// Package ekeys provides functions for managing keys.
package ekeys

import (
	"enclave/econf"
	"encoding/base64"
	"errors"
	"github.com/edgelesssys/ego/ecrypto"
	"os"
	"time"
)

// KeyManager is responsible for generation, encryption, storage, and retrieval
// of keys according to the configuration.
type KeyManager struct {
	Config econf.KeysConfig
}

// KeyGenerator is a function type that generates a pair of public and private keys.
type KeyGenerator func() (publicKey []byte, privateKey []byte, err error)

// GetPrivateKey retrieves the key from location specified by KeysConfig.
// If the private key file doesn't exist, it generates new keys and saves them.
// If the private key file exists, it loads and returns the existing private key.
func (keys KeyManager) GetPrivateKey(generateKey KeyGenerator) ([]byte, error) {
	if _, err := os.Stat(keys.Config.PrivateKeyPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := keys.createNewKeys(generateKey); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return keys.loadKeys()
}

func (keys KeyManager) createNewKeys(generateKey KeyGenerator) error {
	publicKey, privateKey, err := generateKey()
	if err != nil {
		return errors.New("failed to generate keys")
	}
	creationInfo := []byte(time.Now().Format(time.RFC3339))

	err = keys.writeEncryptedFile(keys.Config.SealedDatePath, creationInfo)
	if err != nil {
		return errors.New("failed to write creation info")
	}

	err = os.WriteFile(
		keys.Config.PublicKeyPath,
		[]byte(base64.StdEncoding.EncodeToString(publicKey)),
		0600,
	)
	if err != nil {
		return errors.New("failed to write public key")
	}

	err = keys.writeEncryptedFile(keys.Config.PrivateKeyPath, privateKey)
	if err != nil {
		return errors.New("failed to write private key")
	}

	return nil
}

func (keys KeyManager) loadKeys() ([]byte, error) {
	var dateData []byte
	var creationDate time.Time
	dateData, err := keys.readEncryptedFile(keys.Config.SealedDatePath)
	if err != nil {
		return nil, errors.New("failed to read creation info")
	}

	creationDate, err = time.Parse(time.RFC3339, string(dateData))
	if err != nil {
		return nil, errors.New("failed to parse creation date")
	}

	if time.Now().Before(creationDate) {
		return nil, errors.New("unexpected keys creation date")
	}

	privateKeyData, err := keys.readEncryptedFile(keys.Config.PrivateKeyPath)
	if err != nil {
		return nil, errors.New("failed to read private key")
	}

	return privateKeyData, nil
}

func (keys KeyManager) readEncryptedFile(path string) ([]byte, error) {
	additionalData := []byte(keys.Config.Version)

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
	additionalData := []byte(keys.Config.Version)
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
