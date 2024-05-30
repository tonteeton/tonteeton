package ereport

import (
	"bytes"
	"enclave/eattest"
	"enclave/ebox"
	"enclave/econf"
	"github.com/edgelesssys/ego/attestation"
	"os"
	"testing"
)

func mockAttestation(report attestation.Report, data []byte) (eattest.Attestation, *[]byte) {
	var lastReportData []byte
	return eattest.Attestation{
		GetSelfReport: func() (attestation.Report, error) {
			return report, nil
		},
		GetRemoteReport: func(reportData []byte) ([]byte, error) {
			lastReportData = reportData
			return report.Data, nil
		},
		VerifyRemoteReport: func(reportBytes []byte) (attestation.Report, error) {
			return report, nil
		},
	}, &lastReportData
}

type MockedValues struct {
	config              Config
	attest              eattest.Attestation
	mockedSignature     []byte
	mockedEncryptionKey []byte
	tearDown            func()
}

func setupTest(t *testing.T) MockedValues {
	os.Chdir(t.TempDir())

	var mockedSignature [32]byte
	for i := range mockedSignature {
		mockedSignature[i] = 0xaa
	}

	var mockedEncryptionKey [32]byte
	for i := range mockedEncryptionKey {
		mockedEncryptionKey[i] = 0xbb
	}

	config := Config{
		Reports: econf.ReportsConfig{
			PublicKeysPath:       "report_keys.pub",
			SignatureRequestPath: "report_signature_request.pub",
			SignatureImportPath:  "report_signature.enc",
			SignatureExportPath:  "report_signature.enc",
		},
		SignatureKeys: econf.KeysConfig{
			PublicKeyPath:  "signature_key.pub",
			PrivateKeyPath: "signature_key.priv.enc",
			SealedDatePath: "signature_created.enc",
			Version:        "test",
		},
		EncryptionKeys: econf.KeysConfig{
			PublicKeyPath:  "box_key.pub",
			PrivateKeyPath: "box_key.priv.enc",
			SealedDatePath: "box_created.enc",
			Version:        "test",
		},
	}

	attest, _ := mockAttestation(
		attestation.Report{
			UniqueID:        []byte("mr1"),
			SignerID:        []byte("sign1"),
			ProductID:       []byte("p1"),
			SecurityVersion: 1,
			Debug:           false,
			Data:            append(mockedSignature[:], mockedEncryptionKey[:]...),
		},
		append(mockedSignature[:], mockedEncryptionKey[:]...),
	)

	return MockedValues{
		config,
		attest,
		mockedSignature[:],
		mockedEncryptionKey[:],
		func() {

		},
	}
}

func TestEreport(t *testing.T) {

	t.Run("Keys loaded correctly", func(t *testing.T) {
		mocks := setupTest(t)
		defer mocks.tearDown()

		err := ExportPublicKeys(mocks.config, mocks.attest)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	})

	t.Run("Export private signature key", func(t *testing.T) {
		mocks := setupTest(t)
		defer mocks.tearDown()

		err := os.WriteFile(mocks.config.Reports.SignatureRequestPath, []byte("..."), 0600)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		err = ExportPrivateSignature(mocks.config, mocks.attest)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
	})

}

func TestImportPrivateSignature(t *testing.T) {

	t.Run("Signature exported and imported", func(t *testing.T) {
		mocks := setupTest(t)
		defer mocks.tearDown()

		box, err := ebox.GetBoxKey(mocks.config.EncryptionKeys)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		recipientAttest, lastReportData := mockAttestation(
			attestation.Report{
				UniqueID:        []byte("mr1"),
				SignerID:        []byte("sign1"),
				ProductID:       []byte("p1"),
				SecurityVersion: 1,
				Debug:           false,
				Data:            append(mocks.mockedSignature, box.GetPublicKey()...),
			},
			[]byte("mocked"),
		)

		err = os.WriteFile(mocks.config.Reports.SignatureRequestPath, []byte("mocked"), 0600)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		err = ExportPrivateSignature(mocks.config, recipientAttest)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		signatureAttest, _ := mockAttestation(
			attestation.Report{
				UniqueID:        []byte("mr1"),
				SignerID:        []byte("sign1"),
				ProductID:       []byte("p1"),
				SecurityVersion: 1,
				Debug:           false,
				Data:            *lastReportData,
			},
			[]byte("mocked"),
		)

		err = ImportPrivateSignature(mocks.config, signatureAttest)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

	})

}

func TestPublicKeysReport(t *testing.T) {

	t.Run("Attestation encoded", func(t *testing.T) {
		mocks := setupTest(t)
		defer mocks.tearDown()

		report := &PublicKeysReport{
			PublicSignature:     mocks.mockedSignature[:],
			PublicEncryptionKey: mocks.mockedEncryptionKey[:],
		}
		attestationData, err := report.Encode(mocks.attest)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if !bytes.Equal(attestationData, append(mocks.mockedSignature[:], mocks.mockedEncryptionKey[:]...)) {
			t.Fatalf("Unexpected data: %v", attestationData)
		}
	})

	t.Run("Attestation decoded", func(t *testing.T) {
		mocks := setupTest(t)
		defer mocks.tearDown()

		report := &PublicKeysReport{}
		err := report.Decode([]byte("..."), mocks.attest)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if !bytes.Equal(report.PublicSignature, mocks.mockedSignature[:]) {
			t.Fatalf("Unexpected signature: %#v", report.PublicSignature)
		}
		if !bytes.Equal(report.PublicEncryptionKey, mocks.mockedEncryptionKey[:]) {
			t.Fatalf("Unexpected signature: %#v", report.PublicEncryptionKey)
		}
	})
}
