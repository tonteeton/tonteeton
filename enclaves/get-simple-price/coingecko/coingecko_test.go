package coingecko

import (
	"net/url"
	"reflect"
	"testing"
)

func parametrize[V any, T any](fn T, allValues [][]V) {
	v := reflect.ValueOf(fn)
	for _, a := range allValues {
		vargs := make([]reflect.Value, len(a))

		for i, b := range a {
			vargs[i] = reflect.ValueOf(b)
		}
		v.Call(vargs)
	}
}

func TestCoingeckoFunctions(t *testing.T) {
	t.Run("buildURL", func(t *testing.T) {
		testsArgs := [][]any{
			{
				NewGecko("demokey", ""),
				"/api/v3/simple/price",
				url.Values{"ids": {"ethereum"}},
				"https://api.coingecko.com/api/v3/simple/price?ids=ethereum",
			},
			{
				NewGecko("", ""),
				"/api/v3/simple/price",
				url.Values{"ids": {"ethereum"}},
				"https://api.coingecko.com/api/v3/simple/price?ids=ethereum",
			},
			{
				NewGecko("", "prokey"),
				"/api/v3/simple/price",
				url.Values{"ids": {"ethereum"}},
				"https://pro-api.coingecko.com/api/v3/simple/price?ids=ethereum",
			},
		}
		test := func(gecko GeckoClient, path string, query url.Values, expected string) {
			u, err := gecko.buildURL(path, query)
			if err != nil {
				t.Errorf("Error: %v", err)
			}
			if u != expected {
				t.Errorf("Built unexpected URL: %v,\n expected: %v", u, expected)
			}
		}
		parametrize(test, testsArgs)
	})
}
