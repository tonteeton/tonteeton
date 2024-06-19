// Package appconf provides application-specific configuration management by extending the base configuration provided by econf.
package appconf

import (
	"github.com/tonteeton/golib/econf"
	"os"
)

const (
	TON_TICKER  = uint64(0x72716023)
	APP_VERSION = "get-simple-price-v1r1"
)

// Config extends the econf.Config to include additional application-specific configurations.
type Config struct {
	econf.Config // Embedding main config from econf

	// CoinGecko holds API keys for accessing CoinGecko services.
	CoinGecko struct {
		DemoKey string
		ProKey  string
	}

	// Tickers holds cryptocurrency ticker values.
	Tickers struct {
		TON uint64
	}
}

// LoadConfig loads the application configuration.
func LoadConfig() (*Config, error) {
	cfg := Config{}

	// Load econf.Config sections
	econfConfig, err := econf.LoadConfig(APP_VERSION)
	if err != nil {
		return nil, err
	}
	cfg.Config = *econfConfig

	// Load app-specific configurations
	cfg.Tickers.TON = TON_TICKER
	cfg.CoinGecko.DemoKey = os.Getenv("COINGECKO_API_KEY")
	cfg.CoinGecko.ProKey = os.Getenv("COINGECKO_PRO_API_KEY")

	return &cfg, nil
}
