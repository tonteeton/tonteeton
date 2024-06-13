package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/edgelesssys/ego/attestation"
	"github.com/edgelesssys/ego/attestation/tcbstatus"
	"github.com/edgelesssys/ego/eclient"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func getContractMethod(api ton.APIClientWrapped, contractAddress *address.Address, methodName string) (*ton.ExecutionResult, error) {
	b, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		return nil, err
	}
	return api.WaitForBlock(b.SeqNo).RunGetMethod(
		context.Background(),
		b,
		contractAddress,
		methodName,
	)
}

func decodeReport(encodedData string) ([]byte, error) {
	decodedBase64, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64: %w", err)
	}

	reader, err := gzip.NewReader(bytes.NewReader(decodedBase64))
	if err != nil {
		return nil, fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer reader.Close()

	decodedData, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading decompressed data: %w", err)
	}

	return decodedData, nil
}

func verifyReport(reportBytes []byte, expectedMeasurement string, expectedPublicKey []byte) error {
	report, err := eclient.VerifyRemoteReport(reportBytes)
	if err != nil {
		return fmt.Errorf("error verifying remote report: %w", err)
	}

	if err := verifyReportFields(report); err != nil {
		return err
	}

	measurement := hex.EncodeToString(report.UniqueID)
	if measurement != expectedMeasurement {
		return fmt.Errorf("unexpected enclave measurement: %s", measurement)
	}

	enclavePublicKey := report.Data[:32]
	if !bytes.Equal(enclavePublicKey, expectedPublicKey) {
		return fmt.Errorf(
			"unexpected enclave public key: %s",
			base64.StdEncoding.EncodeToString(enclavePublicKey),
		)
	}

	return nil
}

func verifyReportFields(report attestation.Report) error {
	if report.TCBStatus != tcbstatus.UpToDate {
		return errors.New("reporter SGX platform is not verified")
	}
	if report.UniqueID == nil || report.SignerID == nil {
		return errors.New("invalid report")
	}
	if report.Debug {
		return errors.New("report has debug enabled")
	}
	return nil
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalln("Usage: <contractAddress> <expectedMeasurement>")
	}

	attestationAddress := os.Args[1]
	expectedMeasurement := os.Args[2]

	parsedAddress := address.MustParseAddr(attestationAddress)
	var configURL string
	if parsedAddress.IsTestnetOnly() {
		configURL = "https://ton-blockchain.github.io/testnet-global.config.json"
	} else {
		configURL = "https://ton.org/global.config.json"
	}

	client := liteclient.NewConnectionPool()
	err := client.AddConnectionsFromConfigUrl(context.Background(), configURL)
	if err != nil {
		log.Fatalf("error adding connections from config URL: %v", err)
	}

	api := ton.NewAPIClient(client).WithRetry()

	attestationData, err := getContractMethod(api, parsedAddress, "enclaveAttestation")
	if err != nil {
		log.Fatalf("Error running contract Get method enclaveAttestation: %v", err)
	}

	publicKeyData, err := getContractMethod(api, parsedAddress, "enclavePublicKey")
	if err != nil {
		log.Fatalf("Error running contract Get method enclavePublicKey: %v", err)
	}

	expectedPublicKey := publicKeyData.MustInt(0).Bytes()

	root := attestationData.AsTuple()[0].(*cell.Slice)
	payload, err := root.LoadStringSnake()
	if err != nil {
		log.Fatalf("error loading string: %v", err)
	}

	reportBytes, err := decodeReport(payload)
	if err != nil {
		log.Fatalf("error decoding report: %v", err)
	}

	err = os.WriteFile(
		"report.bin",
		reportBytes,
		0644,
	)
	if err != nil {
		log.Fatalf("error writing report file: %v", err)
	}

	if err = verifyReport(reportBytes, expectedMeasurement, expectedPublicKey); err != nil {
		log.Fatalf("error verifying report: %v", err)
	}

	fmt.Println("✓ Contract attestation report is verified")
}
