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

// DataSealer is a function type for sealing data with
// a signer and product id of the enclave (enclave.SealWithProductKey),
// and with a key derived from a measurement of the enclave (enclave.SealWithUniqueKey).
type DataSealer func(plaintext []byte, additionalData []byte) ([]byte, error)

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

	err = WriteEncryptedFile(keys.Config.SealedDatePath, creationInfo, keys.getAdditionalData())
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

	err = WriteEncryptedFile(keys.Config.PrivateKeyPath, privateKey, keys.getAdditionalData())
	if err != nil {
		return errors.New("failed to write private key")
	}

	return nil
}

func (keys KeyManager) loadKeys() ([]byte, error) {
	var dateData []byte
	var creationDate time.Time
	dateData, err := ReadEncryptedFile(keys.Config.SealedDatePath, keys.getAdditionalData())
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

	privateKeyData, err := ReadEncryptedFile(keys.Config.PrivateKeyPath, keys.getAdditionalData())
	if err != nil {
		return nil, errors.New("failed to read private key")
	}

	return privateKeyData, nil
}

// getAdditionalData returns the additional data used for sealing and unsealing keys.
func (keys KeyManager) getAdditionalData() []byte {
	return []byte(keys.Config.Version)
}

// ReadEncryptedFile reads and decrypts data from the specified path using the provided unsealers.
// If no sealer is provided, the default sealer ecrypto.Unseal is used.
func ReadEncryptedFile(path string, additionalData []byte, unsealers ...DataSealer) ([]byte, error) {
	var data []byte

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(unsealers) == 0 {
		unsealers = []DataSealer{ecrypto.Unseal}
	}

	for _, unseal := range unsealers {
		data, err = unseal(data, additionalData)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// WriteEncryptedFile encrypts and writes data to the specified path using the provided sealers.
// If no sealer is provided, the default sealer ecrypto.SealWithUniqueKey is used.
func WriteEncryptedFile(path string, data []byte, additionalData []byte, sealers ...DataSealer) error {
	var err error
	if len(sealers) == 0 {
		sealers = []DataSealer{ecrypto.SealWithUniqueKey}
	}
	for _, seal := range sealers {
		data, err = seal(data, additionalData)
		if err != nil {
			return err
		}
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	return nil
}
