// Package eattest provides functionality for attesting SGX enclaves.
package eattest

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/edgelesssys/ego/attestation"
	"github.com/edgelesssys/ego/attestation/tcbstatus"
	"github.com/edgelesssys/ego/enclave"
)

// Attestation represents SGX attestation methods.
type Attestation struct {
	GetSelfReport      func() (attestation.Report, error)
	GetRemoteReport    func(reportData []byte) ([]byte, error)
	VerifyRemoteReport func(reportBytes []byte) (attestation.Report, error)
}

// NewAttestation creates a new instance of Attestation with default methods.
func NewAttestation() Attestation {
	return Attestation{
		GetSelfReport: enclave.GetSelfReport,
		GetRemoteReport: enclave.GetRemoteReport,
		VerifyRemoteReport: enclave.VerifyRemoteReport,
	}
}

// GetReport requests a remote attestation report for the provided user data.
func (attest Attestation) GetReport(userData []byte) ([]byte, error) {
	report, err := attest.GetRemoteReport(userData)
	if err != nil {
		return nil, err
	}
	return report, nil
}

// Verify verifies the authenticity of the remote attestation report.
// It compares the remote report with the local SGX report and ensures they match.
func (attest Attestation) Verify(reportData []byte) ([]byte, error) {
	remoteReport, err := attest.VerifyRemoteReport(reportData)
	if err != nil {
		return nil, err
	}

	localReportData, err := attest.GetRemoteReport([]byte("local"))
	if err != nil {
		return nil, err
	}
	localReport, err := attest.VerifyRemoteReport(localReportData)
	if err != nil {
		return nil, err
	}

	return verify(remoteReport, localReport)
}

// verify checks if the remote and local reports are valid and from the same enclave.
func verify(remoteReport attestation.Report, localReport attestation.Report) ([]byte, error) {
	err := verifyReportFields(remoteReport)
	if err != nil {
		return nil, err
	}
	err = verifyReportFields(localReport)
	if err != nil {
		return nil, fmt.Errorf("local report: %s", err)
	}

	if remoteReport.Data == nil {
		return nil, errors.New("report data is empty")
	}

	if !reportsFromSameEnclave(remoteReport, localReport) {
		return nil, errors.New("unexpected report")
	}

	return remoteReport.Data, nil
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

// reportsFromSameEnclave checks if two attestation reports are from the same enclave
// and have identical settings.
func reportsFromSameEnclave(r1 attestation.Report, r2 attestation.Report) bool {
	return bytes.Equal(r1.UniqueID, r2.UniqueID) &&
		bytes.Equal(r1.SignerID, r2.SignerID) &&
		bytes.Equal(r1.ProductID, r2.ProductID) &&
		r1.SecurityVersion == r2.SecurityVersion &&
		r1.Debug == r2.Debug
}
