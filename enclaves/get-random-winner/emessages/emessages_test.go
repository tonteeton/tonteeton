package emessages

import (
	"encoding/hex"
	"testing"
)

func TestRevealedValue(t *testing.T) {
	t.Run("Hash as expected", func(t *testing.T) {
		msg := RevealedValue{
			Timestamp: 1721233023,
			DoraID:    0xaaaaaa,
			Name:      "Test project",
		}
		hash := hex.EncodeToString(msg.Hash())
		expected := "258de9da53e437e731e0c198d62d4749da3febfda671f28d3f90cc57047d61ad"
		if hash != expected {
			t.Errorf("Unexpected hash: %#v. Expected: %#v", hash, expected)
		}
	})
}

func TestRandomReveal(t *testing.T) {
	t.Run("Packed as expected", func(t *testing.T) {
		txHash := make([]byte, 256)
		for i := range txHash {
			txHash[i] = 0xcc
		}
		msg := RandomReveal{
			DoraID:          0xaaaaaa,
			Name:            "Test project",
			RevealTimestamp: 1721233100,
			Nonce:           0xbbbbbb,
			TxHash:          txHash,
		}
		boc := hex.EncodeToString(msg.ToCell().ToBOC())
		expectedBOC := "b5ee9c724101030100480002280000000000aaaaaa6697eecc0000000000bbbbbb01020018546573742070726f6a6563740040cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc885d3b1e"
		if boc != expectedBOC {
			t.Errorf("Unexpected BOC: %#v. Expected: %#v", boc, expectedBOC)
		}
	})
}
