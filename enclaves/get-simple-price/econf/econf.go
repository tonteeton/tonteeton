// Package econf provides functionality to load configuration from environment variables.
package econf

import (
	"os"
)

const (
	// TON_TICKER represents the ticker ID known to the contract.
	TON_TICKER  = uint64(0x72716023)
	APP_VERSION = "get-simple-price-v1r1"
)

type KeysConfig struct {
	PublicKeyPath  string
	PrivateKeyPath string
	SealedDatePath string
	Version        string
}

type ReportsConfig struct {
	PublicKeysPath       string
	SignatureRequestPath string
	SignatureImportPath  string
	SignatureExportPath  string
}

type Config struct {
	CoinGecko struct {
		DemoKey string
		ProKey  string
	}
	Tickers struct {
		TON uint64
	}
	Reports        ReportsConfig
	SignatureKeys  KeysConfig
	EncryptionKeys KeysConfig
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() (*Config, error) {
	cfg := Config{}
	cfg.Tickers.TON = TON_TICKER
	cfg.CoinGecko.DemoKey = os.Getenv("COINGECKO_API_KEY")
	cfg.CoinGecko.ProKey = os.Getenv("COINGECKO_PRO_API_KEY")
	cfg.Reports = ReportsConfig{
		PublicKeysPath:       "mount/report_keys.pub",
		SignatureRequestPath: "mount/report_signature_request.pub",
		SignatureImportPath:  "mount/report_signature_import.enc",
		SignatureExportPath:  "mount/report_signature_export.enc",
	}
	cfg.SignatureKeys = KeysConfig{
		PublicKeyPath:  "mount/signature_key.pub",
		PrivateKeyPath: "mount/signature_key.priv.enc",
		SealedDatePath: "mount/signature_created.enc",
		Version:        APP_VERSION,
	}
	cfg.EncryptionKeys = KeysConfig{
		PublicKeyPath:  "mount/box_key.pub",
		PrivateKeyPath: "mount/box_key.priv.enc",
		SealedDatePath: "mount/box_created.enc",
		Version:        APP_VERSION,
	}
	return &cfg, nil
}
