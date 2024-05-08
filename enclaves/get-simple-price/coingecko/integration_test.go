//go:build integration

package coingecko

import (
	"testing"
)

func TestCoingeckoIntegration(t *testing.T) {
	t.Helper()

	t.Run("getSimplePrice", func(t *testing.T) {
		got, err := NewGecko("", "").GetTONPrice()
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		if got.TON.USD <= 0 {
			t.Errorf("Unexpected response: %v", got)
		}
	})
}
