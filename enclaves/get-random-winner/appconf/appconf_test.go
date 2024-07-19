package appconf

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("LoadConfig", func(t *testing.T) {
		t.Setenv("TON_TESTNET", "1")
		t.Setenv("TON_CONTRACT_ADDRESS", "EQDtFpEwcFAEcRe5mLVh2N6C0x-_hJEM7W61_JLnSF74p4q2")
		t.Setenv("TON_WALLET_MNEMONIC", "test")
		cfg, err := LoadConfig()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		} else if cfg.SignatureKeys.PublicKeyPath == "" {
			t.Errorf("Unexpected keys config: %+v", cfg.SignatureKeys)
		}
	})
}
