package priceresp

import (
	"testing"
)

func TestPrice(t *testing.T) {
	t.Run("ValidCellOutput", func(t *testing.T) {
		price := Price{
			LastUpdatedAt: 1715092161,
			Ticker:        0x72716023,
			USD:           345,
			USD24HVol:     81968225604,
			USD24HChange:  1566,
			BTC:           10967,
		}
		boc := price.ToCell().ToBOC()
		expectedLength := 65
		if len(boc) != expectedLength {
			t.Fatalf("Unexpected BOC length: got %d, expected %d. BOC: %x", len(boc), expectedLength, boc)
		}
	})
}
