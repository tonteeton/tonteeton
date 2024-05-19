// Package esign provides functions for managing Ed25519 signature keys.
// It leverages the ekeys package to retrieve public and private key pairs.
package esign

import (
	"crypto/ed25519"
	"enclave/econf"
	"enclave/ekeys"
	"fmt"
)

// SignatureKey provides methods to access Ed25519 signature keys.
type SignatureKey struct {
	privateKey ed25519.PrivateKey
}

func (key SignatureKey) GetPublicKey() ed25519.PublicKey {
	return key.privateKey.Public().(ed25519.PublicKey)
}

func (key SignatureKey) GetPrivateKey() ed25519.PrivateKey {
	return key.privateKey
}

// GetSignatureKey retrieves the Ed25519 signature key using the provided KeysConfig.
func GetSignatureKey(config econf.KeysConfig) (SignatureKey, error) {
	keys := ekeys.KeyManager{Config: config}
	keyData, err := keys.GetPrivateKey(generateSignatureKey)
	if err != nil {
		return SignatureKey{}, err
	}

	if len(keyData) != ed25519.PrivateKeySize {
		return SignatureKey{}, fmt.Errorf("invalid keys size: %d", len(keyData))
	}

	return SignatureKey{ed25519.PrivateKey(keyData)}, nil
}

// generateSignatureKey implements the ekeys.KeyGenerator function type.
func generateSignatureKey() (public []byte, private []byte, err error) {
	return ed25519.GenerateKey(nil)
}
