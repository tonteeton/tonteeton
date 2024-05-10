package coinconv

import (
	"enclave/coingecko"
	"enclave/eresp"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

func TestConvertPrice(t *testing.T) {
	cases := []struct {
		input    coingecko.SimplePrice
		expected eresp.EnclavePrice
	}{
		{
			coingecko.SimplePrice{},
			eresp.EnclavePrice{},
		},
		{
			coingecko.SimplePrice{USD: 1},
			eresp.EnclavePrice{USD: 1_00},
		},
		{
			coingecko.SimplePrice{USD: 1},
			eresp.EnclavePrice{USD: 1_00},
		},
		{
			coingecko.SimplePrice{USD: 5.792609218137362},
			eresp.EnclavePrice{USD: 5_79},
		},
		{
			coingecko.SimplePrice{USD: 0.2697380},
			eresp.EnclavePrice{USD: 27},
		},
		{
			coingecko.SimplePrice{USD: 0.00999},
			eresp.EnclavePrice{USD: 1},
		},
		{
			coingecko.SimplePrice{BTC: 9.2687479721799e-05},
			eresp.EnclavePrice{BTC: 9_269},
		},
		{
			coingecko.SimplePrice{
				LastUpdatedAt: 1715266741,
				USD:           100.09218137362,
				USD24HVol:     331937525.21919525,
				USD24HChange:  -5.178720470976301,
				BTC:           0.92687479721799e-05,
			},
			eresp.EnclavePrice{
				LastUpdatedAt: 1715266741,
				USD:           100_09,
				USD24HChange:  -518,
				USD24HVol:     331_937_525_22,
				BTC:           927,
			},
		},
	}

	for _, tcase := range cases {
		t.Run(fmt.Sprintf("%+v", tcase.input), func(t *testing.T) {
			got := ConvertPrice(tcase.input, 0)
			if got != tcase.expected {
				t.Errorf("Unexpected: %+v,\n expected: %+v", got, tcase.expected)
			}
		})
	}

	t.Run("Ticker is set", func(t *testing.T) {
		got := ConvertPrice(coingecko.SimplePrice{}, 1234)
		if got.Ticker != 1234 {
			t.Errorf("Ticker is not set")
		}

	})
}

func TestConvertFloatValueToInt(t *testing.T) {
	cases := []struct {
		value     float64
		precision int
		expected  int64
	}{
		{6.661117286704982, 2, 666},
		{6.669117286704982, 2, 667},
		{6.669117286704982, 8, 666911729},
		{-7.1599020652620675, 2, -716},
		{0.000110204824752621, 8, 11020},
		{110204824752621, 8, math.MaxInt64},
		{0, 2, 0},
	}
	for _, tcase := range cases {
		t.Run(fmt.Sprintf("%+v", tcase), func(t *testing.T) {
			got := convertFloatValueToInt(tcase.value, tcase.precision)
			if got != tcase.expected {
				t.Errorf("Unexpected: %+v,\n expected: %+v", got, tcase.expected)
			}
		})
	}
}

func TestPriceIsValid(t *testing.T) {
	cases := []eresp.EnclavePrice{
		eresp.EnclavePrice{
			LastUpdatedAt: uint64(time.Now().Unix()),
			Ticker:        1,
			USD:           100_09,
			USD24HChange:  -518,
			USD24HVol:     331_937_525_22,
			BTC:           927,
		},
	}

	for _, tcase := range cases {
		t.Run(fmt.Sprintf("%+v", tcase), func(t *testing.T) {
			err := ValidatePrice(tcase)
			if err != nil {
				t.Errorf("Error: %v", err)
			}
		})
	}
}

func TestPriceIsInvalid(t *testing.T) {
	now := uint64(time.Now().Unix())

	cases := []struct {
		input       eresp.EnclavePrice
		expectedErr string
	}{
		{
			eresp.EnclavePrice{},
			"",
		},
		{
			eresp.EnclavePrice{
				LastUpdatedAt: uint64(
					time.Now().Add(-24 * time.Hour).Unix(),
				),
				Ticker:       1,
				USD:          100_09,
				USD24HChange: -518,
				USD24HVol:    331_937_525_22,
				BTC:          927,
			},
			"LastUpdatedAt",
		},
		{
			eresp.EnclavePrice{
				LastUpdatedAt: now,
				Ticker:        1,
				USD:           0,
				USD24HChange:  -518,
				USD24HVol:     331_937_525_22,
				BTC:           927,
			},
			"USD value",
		},
		{
			eresp.EnclavePrice{
				LastUpdatedAt: uint64(time.Now().Unix()),
				Ticker:        1,
				USD:           1,
				USD24HChange:  -1e10,
				USD24HVol:     331_937_525_22,
				BTC:           927,
			},
			"USD24HChange",
		},
		{
			eresp.EnclavePrice{
				LastUpdatedAt: uint64(time.Now().Unix()),
				Ticker:        1,
				USD:           1,
				USD24HChange:  0,
				USD24HVol:     331_937_525_22,
				BTC:           927,
			},
			"USD24HChange",
		},
		{
			eresp.EnclavePrice{
				LastUpdatedAt: now,
				Ticker:        1,
				USD:           1,
				USD24HChange:  1,
				USD24HVol:     1e19,
				BTC:           927,
			},
			"USD24HVol",
		},

		{
			eresp.EnclavePrice{
				LastUpdatedAt: now,
				Ticker:        1,
				USD:           1,
				USD24HChange:  -1000,
				USD24HVol:     1,
				BTC:           0,
			},
			"BTC value",
		},
		{

			eresp.EnclavePrice{
				LastUpdatedAt: uint64(time.Now().Unix()),
				USD:           100_09,
				USD24HChange:  -518,
				USD24HVol:     331_937_525_22,
				BTC:           927,
			},
			"Ticker",
		},
	}

	for _, tcase := range cases {
		t.Run(fmt.Sprintf("%+v", tcase.input), func(t *testing.T) {
			err := ValidatePrice(tcase.input)
			if err == nil {
				t.Errorf("Expected error not raised: %+v", tcase.expectedErr)
			} else if !strings.Contains(err.Error(), tcase.expectedErr) {
				t.Errorf(
					"Unexpected error: %+v,\n expected error: %+v",
					err,
					tcase.expectedErr,
				)
			}
		})
	}
}
