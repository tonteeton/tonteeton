// Package coingecko provides a client for interacting with the CoinGecko API to fetch cryptocurrency price data.
package coingecko

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// SimplePrice represents the price information for a coin.
type SimplePrice struct {
	LastUpdatedAt uint64  `json:"last_updated_at"`
	USD           float64 `json:"usd"`
	USD24HVol     float64 `json:"usd_24h_vol"`
	USD24HChange  float64 `json:"usd_24h_change"`
	BTC           float64 `json:"btc"`
}

// SimplePriceResponse represents the response from the `Coin Price by IDs` API endpoint.
type SimplePriceResponse struct {
	TON SimplePrice `json:"the-open-network"`
}

// GeckoClient represents a client for interacting with the CoinGecko API.
type GeckoClient struct {
	host         string
	apiKeyHeader string
	apiKey       string
}

// NewGecko creates a new GeckoClient instance.
// proAPIKey and demoAPIKey are keys for PRO and Demo APIs, may be empty.
func NewGecko(demoAPIKey string, proAPIKey string) GeckoClient {
	var host, apiKeyHeader, apiKey string
	switch {
	case proAPIKey != "":
		host = "pro-api.coingecko.com"
		apiKeyHeader = "x-cg-pro-api-key"
		apiKey = proAPIKey
	case demoAPIKey != "":
		host = "api.coingecko.com"
		apiKeyHeader = "x-cg-demo-api-key"
		apiKey = demoAPIKey
	default:
		host = "api.coingecko.com"
		apiKeyHeader = ""
		apiKey = ""
	}

	return GeckoClient{host, apiKeyHeader, apiKey}
}

// GetTONPrice queries CoinGecko for the prices of TON.
// Reference: https://docs.coingecko.com/reference/simple-price
func (gecko GeckoClient) GetTONPrice() (SimplePriceResponse, error) {
	query := url.Values{
		"include_24hr_vol":        {"true"},
		"include_24hr_change":     {"true"},
		"include_last_updated_at": {"true"},
		"precision":               {"18"},
	}
	query.Set("ids", "the-open-network")
	query.Set("vs_currencies", "USD,BTC")

	apiURL, err := gecko.buildURL(`/api/v3/simple/price`, query)
	if err != nil {
		return SimplePriceResponse{}, err
	}

	data, err := gecko.get(apiURL)
	if err != nil {
		return SimplePriceResponse{}, err
	}
	if len(data) == 0 {
		return SimplePriceResponse{}, errors.New("Empty response body")
	}
	var prices SimplePriceResponse
	err = json.Unmarshal(data, &prices)
	if err != nil {
		return SimplePriceResponse{}, err
	}

	return prices, nil
}

func (gecko GeckoClient) buildURL(path string, query url.Values) (string, error) {
	apiURL := &url.URL{
		Scheme:   "https",
		Host:     gecko.host,
		Path:     path,
		RawQuery: query.Encode(),
	}
	if _, err := url.ParseRequestURI(apiURL.String()); err != nil {
		return "", err
	}
	return apiURL.String(), nil
}

func (gecko GeckoClient) get(apiURL string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if gecko.apiKeyHeader != "" && gecko.apiKey != "" {
		req.Header.Set(gecko.apiKeyHeader, gecko.apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
