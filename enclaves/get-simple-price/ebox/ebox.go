// Package ebox provides functions for managing NaCl Box encryption keys.
// It leverages the ekeys package to retrieve public and private key pairs.
package ebox

import (
	"crypto/rand"
	"enclave/econf"
	"enclave/ekeys"
	"errors"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io"
)

const (
	NonceSize      = 24
	PublicKeySize  = 32
	PrivateKeySize = 32
)

// BoxKey provides methods to access encryption keys.
type BoxKey struct {
	publicKey  [PublicKeySize]byte
	privateKey [PrivateKeySize]byte
}

func (boxkey BoxKey) GetPublicKey() []byte {
	return boxkey.publicKey[:]
}

func (boxkey BoxKey) GetPrivateKey() []byte {
	return boxkey.privateKey[:]
}

// Encrypt encrypts the given message using the recipient's public key.
func (boxkey BoxKey) Encrypt(msg []byte, recipientPublicKey []byte) ([]byte, error) {
	if len(recipientPublicKey) != PublicKeySize {
		return nil, errors.New("invalid public key size")
	}
	var recipientKey [32]byte
	copy(recipientKey[:], recipientPublicKey)

	var nonce [NonceSize]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, errors.New("nonce genereation error")
	}

	encrypted := box.Seal(nonce[:], msg, &nonce, &recipientKey, &boxkey.privateKey)
	return encrypted, nil
}

// Decrypt decrypts the given encrypted message using the sender's public key.
func (boxkey BoxKey) Decrypt(encryptedMsg []byte, senderPublicKey []byte) ([]byte, error) {
	if len(senderPublicKey) != PublicKeySize {
		return nil, errors.New("invalid public key size")
	}

	if len(encryptedMsg) <= NonceSize {
		return nil, errors.New("invalid message size")
	}

	var senderKey [32]byte
	copy(senderKey[:], senderPublicKey)

	var decryptNonce [NonceSize]byte
	copy(decryptNonce[:], encryptedMsg[:NonceSize])

	decrypted, ok := box.Open(nil, encryptedMsg[NonceSize:], &decryptNonce, &senderKey, &boxkey.privateKey)
	if !ok {
		return nil, errors.New("decryption error")
	}

	return decrypted, nil
}

// GetBoxKey retrieves the NaCl Box key using the provided KeysConfig.
func GetBoxKey(config econf.KeysConfig) (BoxKey, error) {
	keys := ekeys.KeyManager{Config: config}
	keyData, err := keys.GetPrivateKey(generateBoxKey)
	if err != nil {
		return BoxKey{}, err
	}

	if len(keyData) != PublicKeySize+PrivateKeySize {
		return BoxKey{}, fmt.Errorf("invalid keys size: %d", len(keyData))
	}

	var key BoxKey
	copy(key.publicKey[:], keyData[:PublicKeySize])
	copy(key.privateKey[:], keyData[PublicKeySize:])
	return key, nil
}

var GenerateKey = box.GenerateKey

// generateBoxKey implements the ekeys.KeyGenerator function type.
func generateBoxKey() (public []byte, private []byte, err error) {
	publicKey, privateKey, err := GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	if publicKey == nil || privateKey == nil {
		return nil, nil, errors.New("generated empty key pair")
	}

	// Concatenate the public and private keys for storage.
	// Both public and private keys are loaded together from the encrypted file.
	// And public key is also stored in unencrypted form.
	keyPair := append(publicKey[:], privateKey[:]...)

	return publicKey[:], keyPair, nil
}
