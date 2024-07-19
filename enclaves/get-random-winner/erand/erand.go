// Package erand provides functions for generating random numbers.
package erand

import (
	"crypto/rand"
	"errors"
	"math"
	"math/big"
)

// RandUInt64 generates a pseudo-random uint64 value in the range [0, maxValue).
func RandUInt64(maxValue uint64) (uint64, error) {
	if maxValue == 0 {
		return math.MaxUint64, errors.New("maxValue must be greater than 0")
	}
	maxBig := big.NewInt(0).SetUint64(maxValue)
	value, err := rand.Int(rand.Reader, maxBig)
	if err != nil {
		return math.MaxUint64, err
	}
	return value.Uint64(), nil
}
