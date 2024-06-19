// Package coinconv provides functions for converting price data.
package coinconv

import (
	"enclave/coingecko"
	"enclave/priceresp"
	"errors"
	"math"
	"time"
)

// ConvertPrice converts price data from Coingecko format to enclave response format.
func ConvertPrice(from coingecko.SimplePrice, ticker uint64) priceresp.Price {
	return priceresp.Price{
		LastUpdatedAt: from.LastUpdatedAt,
		Ticker:        ticker,
		USD:           uint64(convertFloatValueToInt(from.USD, 2)),
		USD24HVol:     uint64(convertFloatValueToInt(from.USD24HVol, 2)),
		USD24HChange:  convertFloatValueToInt(from.USD24HChange, 2),
		BTC:           uint64(convertFloatValueToInt(from.BTC, 8)),
	}
}

// convertFloatValueToInt converts a float value to an int64,
// rounding to the nearest integer with the specified precision.
func convertFloatValueToInt(value float64, precision int) int64 {
	scaledValue := value * math.Pow10(precision)
	roundedValue := math.Round(scaledValue)
	if roundedValue > float64(math.MaxInt64) {
		return math.MaxInt64
	}
	return int64(roundedValue)
}

// ValidatePrice validates priceresp.Price struct fields.
func ValidatePrice(price priceresp.Price) error {
	currentTime := time.Now()
	lastUpdatedAt := time.Unix(int64(price.LastUpdatedAt), 0)

	if price.USD < 1 || price.USD >= math.MaxInt64 {
		return errors.New("USD value is out of valid range")
	}

	if price.USD24HVol < 1 || price.USD24HVol >= math.MaxInt64 {
		return errors.New("USD24HVol value is out of valid range")
	}

	if price.USD24HChange < -1000*1e2 ||
		price.USD24HChange == 0 ||
		price.USD24HChange > 1000*1e2 {
		return errors.New("USD24HChange value is out of valid range")
	}

	if price.BTC < 1 || price.BTC >= math.MaxInt64 {
		return errors.New("BTC value is out of valid range")
	}

	if lastUpdatedAt.Before(currentTime.Add(-30*time.Minute)) ||
		lastUpdatedAt.After(currentTime.Add(30*time.Minute)) {
		return errors.New("LastUpdatedAt is not within the valid time range")
	}

	if price.Ticker == 0 {
		return errors.New("Ticker is not specified")
	}

	return nil
}
