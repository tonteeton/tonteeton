package eattest

import (
	"fmt"
	"github.com/edgelesssys/ego/attestation"
	"github.com/edgelesssys/ego/attestation/tcbstatus"
	"strings"
	"testing"
)


func TestAttestationGenerated(t *testing.T) {
	attest := NewAttestation()
	testCases := []struct {
		Name     string
		RunFunc  func() (error)
		Expected string // empty string if no error expected
	}{
		{
			Name:    "GetReport data too large",
			Expected: "too large",
			RunFunc: func () error {
				var reportData [65]byte
				_, err := attest.GetReport(reportData[:])
				return err
			},
		},
		{
			Name:    "GetReport require SGX platform",
			Expected: "OE_UNSUPPORTED",
			RunFunc: func () error {
				var reportData [64]byte
				_, err := attest.GetReport(reportData[:])
				return err
			},
		},
		{
			Name:    "Verify require SGX platform",
			Expected: "OE_UNSUPPORTED",
			RunFunc: func () error {
				_, err := attest.Verify([]byte("test"))
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			verifyError(t, tc.RunFunc(), tc.Expected)
		})
	}
}

func verifyError(t *testing.T, err error, expectedSubstring string) {
	if expectedSubstring == "" {
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	} else {
		if err == nil || !strings.Contains(err.Error(), expectedSubstring) {
			t.Fatalf("Expected error containing %q, got: %v", expectedSubstring, err)
		}
	}
}


func TestReportsComparisonAttestation(t *testing.T) {

	cases := []struct {
		r1       attestation.Report
		r2       attestation.Report
		expected bool
	}{
		{
			attestation.Report{},
			attestation.Report{},
			true,
		},
		{
			attestation.Report{
				UniqueID:        []byte("mr1"),
				SignerID:        []byte("sign1"),
				ProductID:       []byte("p1"),
				SecurityVersion: 1,
				Debug:           true,
			},
			attestation.Report{
				UniqueID:        []byte("mr1"),
				SignerID:        []byte("sign1"),
				ProductID:       []byte("p1"),
				SecurityVersion: 1,
				Debug:           true,
			},
			true,
		},
		{
			attestation.Report{Debug: true},
			attestation.Report{},
			false,
		},
		{
			attestation.Report{UniqueID: []byte("test1")},
			attestation.Report{UniqueID: []byte("test2")},
			false,
		},
		{
			attestation.Report{
				UniqueID:        []byte("mr1"),
				SignerID:        []byte("sign1"),
				ProductID:       []byte("p1"),
				SecurityVersion: 1,
				Debug:           true,
			},
			attestation.Report{
				UniqueID:        []byte("mr2"),
				SignerID:        []byte("sign1"),
				ProductID:       []byte("p1"),
				SecurityVersion: 1,
				Debug:           true,
			},
			false,
		},
	}
	for i, tcase := range cases {
		t.Run(fmt.Sprintf("Compare reports %d", i), func(t *testing.T) {
			if reportsFromSameEnclave(tcase.r1, tcase.r2) != tcase.expected {
				t.Errorf("Unexpected result for case %d", i)
			}
		})
	}
}

func TestReportsVerification(t *testing.T) {
	validationCases := []struct {
		r1       attestation.Report
		r2       attestation.Report
		expected string
	}{
		{
			attestation.Report{},
			attestation.Report{},
			string("invalid report"),
		},
		{
			attestation.Report{TCBStatus: tcbstatus.Revoked},
			attestation.Report{},
			string("reporter SGX platform is not verified"),
		},
		{
			attestation.Report{
				UniqueID: []byte("mr1"),
				SignerID: []byte("s1"),
				Data:     []byte("test"),
			},
			attestation.Report{
				UniqueID: []byte("mr2"),
			},
			string("local report: invalid report"),
		},
		{
			attestation.Report{
				UniqueID: []byte("mr1"),
				SignerID: []byte("s1"),
				Data:     []byte("test"),
			},
			attestation.Report{
				UniqueID: []byte("mr2"),
				SignerID: []byte("s1"),
			},
			string("unexpected report"),
		},
		{
			attestation.Report{
				UniqueID: []byte("mr1"),
				SignerID: []byte("s1"),
				Data:     []byte("test"),
			},
			attestation.Report{
				UniqueID: []byte("mr1"),
				SignerID: []byte("s1"),
			},
			string(""),
		},
	}
	for i, tcase := range validationCases {
		t.Run(fmt.Sprintf("Verify report %d", i), func(t *testing.T) {
			_, err := verify(tcase.r1, tcase.r2)
			if tcase.expected == "" && err != nil {
				t.Errorf("Error: %v", err)
			} else if tcase.expected != "" && err == nil {
				t.Errorf("Expected error not raised")
			} else if tcase.expected != "" && err != nil && err.Error() != tcase.expected {
				t.Errorf("Unexpected error raised: %v\nError expected: %v", err, tcase.expected)
			}

		})
	}
}
