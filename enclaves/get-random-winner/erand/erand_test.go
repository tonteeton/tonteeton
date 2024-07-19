package erand

import (
	"math"
	"testing"
)

func TestRandUInt64(t *testing.T) {
	tests := []struct {
		name     string
		maxValue uint64
	}{
		{"Max Value 0", 0},
		{"Max Value 1", 1},
		{"Max Value 100", 100},
		{"Max Value MaxUint64 Minus 1", math.MaxUint64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 1000; i++ { // Run the test multiple times to check the randomness
				result, err := RandUInt64(tt.maxValue)
				if tt.maxValue == 0 {
					if err == nil {
						t.Errorf("RandUInt64(%d) did not return an error", tt.maxValue)
					}
					if result != math.MaxUint64 {
						t.Errorf("RandUInt64(%d) = %d, expected math.MaxUint64 on error", tt.maxValue, result)
					}
				} else {
					if err != nil {
						t.Errorf("RandUInt64(%d) returned error: %v", tt.maxValue, err)
					}
					if result >= tt.maxValue {
						t.Errorf("RandUInt64(%d) = %d, out of range [0, %d)", tt.maxValue, result, tt.maxValue)
					}
				}
			}
		})
	}
}
