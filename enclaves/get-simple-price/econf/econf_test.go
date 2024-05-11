package econf

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("LoadConfig", func(t *testing.T) {
		t.Setenv("COINGECKO_API_KEY", "demo")
		t.Setenv("COINGECKO_PRO_API_KEY", "pro")
		cfg, err := LoadConfig()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		} else if cfg.Tickers.TON == 0 || cfg.CoinGecko.DemoKey != "demo" || cfg.CoinGecko.ProKey != "pro" {
			t.Errorf("Unexpected config: %+v", cfg)
		} else if cfg.Keys.PublicKeyPath == "" {
			t.Errorf("Unexpected keys config: %+v", cfg.Keys)
		}
	})
}
