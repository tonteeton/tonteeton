// Package ereport provides functionality for generating and handling
// signed reports containing public keys and encrypted data.
package ereport

import (
	"enclave/eattest"
	"enclave/ebox"
	"enclave/econf"
	"enclave/esign"
	"errors"
	"google.golang.org/protobuf/proto"
	"os"
)

const (
	PublicSignatureSize     = 32
	PublicEncryptionKeySize = 32
)

// Config contains configuration parameters for generating reports.
type Config struct {
	Reports        econf.ReportsConfig
	SignatureKeys  econf.KeysConfig
	EncryptionKeys econf.KeysConfig
}

// Generate SGX-signed report with public signature and encryption keys.
func ExportPublicKeys(cfg Config, attest eattest.Attestation) error {
	signature, err := esign.GetSignatureKey(cfg.SignatureKeys)
	if err != nil {
		return err
	}
	box, err := ebox.GetBoxKey(cfg.EncryptionKeys)
	if err != nil {
		return err
	}

	report := &PublicKeysReport{
		PublicSignature:     signature.GetPublicKey(),
		PublicEncryptionKey: box.GetPublicKey(),
	}

	attestationData, err := report.Encode(attest)
	if err != nil {
		return err
	}

	err = os.WriteFile(cfg.Reports.PublicKeysPath, attestationData, 0600)

	return err
}

// ExportPrivateSignature exports a private signature encrypted for the recipient.
func ExportPrivateSignature(cfg Config, attest eattest.Attestation) error {
	// Load recipent public encryption key
	recipientData, err := os.ReadFile(cfg.Reports.SignatureRequestPath)
	if err != nil {
		return err
	}
	recipientKeys := &PublicKeysReport{}
	err = recipientKeys.Decode(recipientData, attest)
	if err != nil {
		return err
	}

	// Load the signature
	signature, err := esign.GetSignatureKey(cfg.SignatureKeys)
	if err != nil {
		return err
	}

	payload, err := encodePrivateSignatureKey(cfg, attest, signature, recipientKeys)
	if err != nil {
		return err
	}

	err = os.WriteFile(cfg.Reports.SignatureExportPath, payload, 0600)

	return err
}

func encodePrivateSignatureKey(
	cfg Config,
	attest eattest.Attestation,
	signature esign.SignatureKey,
	recipientKeys *PublicKeysReport,
) ([]byte, error) {
	// Encrypt signature private key using recipient public key
	box, err := ebox.GetBoxKey(cfg.EncryptionKeys)
	if err != nil {
		return nil, err
	}
	encryptedSignature, err := box.Encrypt(
		signature.GetPrivateKey(),
		recipientKeys.PublicEncryptionKey,
	)
	if err != nil {
		return nil, err
	}

	// Store signature of the payload (encrypted signature key)
	// in the attestation report
	report, err := attest.GetReport(signature.Sign(encryptedSignature))
	if err != nil {
		return nil, err
	}

	// Marshals the payload
	payload, err := proto.Marshal(&PrivateKeysReport{
		AttestationReport:     report,
		PublicEncryptionKey:   box.GetPublicKey(),
		EncryptedSignatureKey: encryptedSignature,
	})
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// ImportPrivateSignature imports encrypted private signature from an attestation report.
func ImportPrivateSignature(cfg Config, attest eattest.Attestation) error {
	encodedData, err := os.ReadFile(cfg.Reports.SignatureImportPath)
	if err != nil {
		return err
	}
	var rpt PrivateKeysReport
	err = proto.Unmarshal(encodedData, &rpt)
	if err != nil {
		return err
	}

	signaturePrivateKey, err := decodePrivateSignatureKey(cfg, attest, &rpt)
	if err != nil {
		return err
	}

	err = esign.SaveSignatureKey(cfg.SignatureKeys, signaturePrivateKey)

	if err != nil {
		return err
	}

	return nil
}

func decodePrivateSignatureKey(cfg Config, attest eattest.Attestation, rpt *PrivateKeysReport) ([]byte, error) {
	// Get signature of the payload (encrypted signature key)
	// from the attestation report
	sig, err := attest.Verify(rpt.AttestationReport)
	if err != nil {
		return nil, err
	}

	box, err := ebox.GetBoxKey(cfg.EncryptionKeys)
	if err != nil {
		return nil, err
	}

	signaturePrivateKey, err := box.Decrypt(rpt.EncryptedSignatureKey, rpt.PublicEncryptionKey)
	if err != nil {
		return nil, err
	}

	// Verify signature of the payload (encrypted signature key)
	if !(esign.SignatureKey{PrivateKey: signaturePrivateKey}).Verify(
		rpt.EncryptedSignatureKey,
		sig,
	) {
		return nil, errors.New("wrong signature")
	}

	return signaturePrivateKey, nil
}

// PublicKeysReport represents a report containing public signature and encryption keys.
type PublicKeysReport struct {
	PublicSignature     []byte
	PublicEncryptionKey []byte
}

// Encode encodes the public signature and encryption key into an attestation report using the provided attestation service.
func (rpt *PublicKeysReport) Encode(attest eattest.Attestation) ([]byte, error) {
	if len(rpt.PublicSignature) != PublicSignatureSize || len(rpt.PublicEncryptionKey) != PublicEncryptionKeySize {
		return nil, errors.New("unexpected keys length")
	}

	userData := append(
		rpt.PublicSignature[:],
		rpt.PublicEncryptionKey[:]...,
	)

	attestationData, err := attest.GetReport(userData)
	if err != nil {
		return nil, err
	}

	return attestationData, nil
}

// Decode decodes the public signature and encryption key from the provided attestation data.
// The attestation data is verified using the attestation service.
func (rpt *PublicKeysReport) Decode(attestationData []byte, attest eattest.Attestation) error {
	userData, err := attest.Verify(attestationData)
	if err != nil {
		return err
	}

	if len(userData) != PublicSignatureSize+PublicEncryptionKeySize {
		return errors.New("unexpected user data size")
	}

	rpt.PublicSignature = userData[:PublicSignatureSize]
	rpt.PublicEncryptionKey = userData[PublicSignatureSize:]

	return nil
}
