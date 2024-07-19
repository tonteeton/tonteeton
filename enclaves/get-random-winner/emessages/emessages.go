// Package emessages provides a data structures for Random Updates protocol.
package emessages

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type RandomCommit struct {
	Timestamp uint32
	Recipient *address.Address
	ValueHash []byte // The hash of RevealedValue value.
}

func (msg RandomCommit) GetOpcode() uint32 {
	return 0xbb15fe7d
}

func (msg RandomCommit) ToCell() *cell.Cell {
	hashCell := cell.BeginCell().
		MustStoreSlice(msg.ValueHash, 256).
		EndCell()

	return cell.BeginCell().
		MustStoreUInt(uint64(msg.Timestamp), 32).
		MustStoreAddr(msg.Recipient).
		MustStoreRef(hashCell).
		EndCell()
}

type RandomReveal struct {
	DoraID          uint64
	Name            string
	RevealTimestamp uint32
	Nonce           uint64
	TxHash          []byte
}

func (msg RandomReveal) GetOpcode() uint32 {
	return 0x6b91a49a
}

func (msg RandomReveal) ToCell() *cell.Cell {
	txHashCell := cell.BeginCell().MustStoreSlice(msg.TxHash, 256).EndCell()
	return cell.BeginCell().
		MustStoreUInt(uint64(msg.DoraID), 64).
		MustStoreRef(cell.BeginCell().MustStoreStringSnake(msg.Name).EndCell()).
		MustStoreUInt(uint64(msg.RevealTimestamp), 32).
		MustStoreUInt(msg.Nonce, 64).
		MustStoreRef(txHashCell).
		EndCell()
}


type RevealedValue struct {
	Timestamp uint32
	Recipient *address.Address
	Nonce     uint64
	DoraID    uint64
	Name      string
}

func (msg RevealedValue) ToCell() *cell.Cell {
	return cell.BeginCell().
		MustStoreUInt(uint64(msg.Timestamp), 32).
		MustStoreAddr(msg.Recipient).
		MustStoreUInt(msg.Nonce, 64).
		MustStoreUInt(uint64(msg.DoraID), 64).
		MustStoreRef(cell.BeginCell().MustStoreStringSnake(msg.Name).EndCell()).
		EndCell()
}

func (msg RevealedValue) Hash() []byte {
	return msg.ToCell().Hash()
}
