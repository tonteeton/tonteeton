// Package econf provides functionality to load configuration from environment variables.
package econf

import (
	"os"
)

const (
	// TON_TICKER represents the ticker ID known to the contract.
	TON_TICKER = uint64(0x72716023)
)

type Config struct {
	CoinGecko struct {
		DemoKey string
		ProKey  string
	}
	Tickers struct {
		TON uint64
	}
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	cfg := Config{}
	cfg.Tickers.TON = TON_TICKER
	cfg.CoinGecko.DemoKey = os.Getenv("COINGECKO_API_KEY")
	cfg.CoinGecko.ProKey = os.Getenv("COINGECKO_PRO_API_KEY")
	return &cfg, nil
}
