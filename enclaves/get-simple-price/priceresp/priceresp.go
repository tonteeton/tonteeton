// Package priceresp provides a data structure for representing cryptocurrency prices
// and a method to serialize this data structure into a TON-compatible TVM cell.
package priceresp

import (
	"github.com/xssnick/tonutils-go/tvm/cell"
)

// Price represents cryptocurrency price information for serialization into a TVM cell.
type Price struct {
	LastUpdatedAt uint64
	Ticker        uint64
	USD           uint64
	USD24HVol     uint64
	USD24HChange  int64
	BTC           uint64
}

// ToCell serializes the Price struct into a TVM cell.
func (price Price) ToCell() *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(price.LastUpdatedAt, 64).
		MustStoreUInt(price.Ticker, 64).
		MustStoreUInt(price.USD, 64).
		MustStoreUInt(price.USD24HVol, 64).
		MustStoreInt(price.USD24HChange, 64).
		MustStoreUInt(price.BTC, 64).
		EndCell()
}
