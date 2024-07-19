// Package appconf provides application-specific configuration management by extending the base configuration provided by econf.
package appconf

import (
	"errors"
	"github.com/tonteeton/golib/econf"
	"github.com/xssnick/tonutils-go/address"
	"os"
	"strconv"
	"strings"
)

const (
	APP_VERSION    = "get-random-int-v1r1"
	TESTNET_CONFIG = "https://ton.org/testnet-global.config.json"
	MAINNET_CONFIG = "https://ton.org/global.config.json"
)

// Config extends the econf.Config to include additional application-specific configurations.
type Config struct {
	econf.Config // Embedding main config from econf

	Network struct {
		TestNet         bool
		GlobalConfigURL string
		ContractAddress *address.Address
	}

	Wallet struct {
		Mnemonic []string
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

	testNetEnv := os.Getenv("TON_TESTNET")
	if testNetEnv == "" {
		return nil, errors.New("TON_TESTNET env is not set")
	}
	testNet, err := strconv.ParseBool(testNetEnv)
	if err != nil {
		return nil, err
	}
	cfg.Network.TestNet = testNet
	if testNet {
		cfg.Network.GlobalConfigURL = TESTNET_CONFIG
	} else {
		cfg.Network.GlobalConfigURL = MAINNET_CONFIG
	}

	contractAddress := os.Getenv("TON_CONTRACT_ADDRESS")
	if contractAddress == "" {
		return nil, errors.New("TON_CONTRACT_ADDRESS env is not set")
	}
	parsedAddress, err := address.ParseAddr(contractAddress)
	if err != nil {
		return nil, err
	}
	parsedAddress.SetTestnetOnly(false)
	parsedAddress.SetBounce(false)
	cfg.Network.ContractAddress = parsedAddress

	mnemonic := os.Getenv("TON_WALLET_MNEMONIC")
	if mnemonic == "" {
		return nil, errors.New("TON_WALLET_MNEMONIC env is not set")
	}
	cfg.Wallet.Mnemonic = strings.Split(mnemonic, " ")

	return &cfg, nil
}
