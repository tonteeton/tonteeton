package txparser

import (
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type TransactionParser struct {
	TestNet bool
	Address *address.Address
}

func (parser TransactionParser) hasAddress(addr *address.Address) bool {
	addr.SetTestnetOnly(false)
	addr.SetBounce(false)
	return parser.Address.String() == addr.String()
}

// ParseExternalComments return list of emitted comments.
func (parser TransactionParser) ParseExternalComments(tx *tlb.Transaction) []string {
	var comments []string

	if tx.IO.Out != nil {
		messages, err := tx.IO.Out.ToSlice()
		if err != nil {
			return nil
		}
		for _, m := range messages {
			switch m.MsgType {
			case tlb.MsgTypeExternalOut:
				externalOut := m.AsExternalOut()
				comment := parser.ParseComment(externalOut.Body)
				if parser.hasAddress(externalOut.SrcAddr) && externalOut.DstAddr.IsAddrNone() && comment != "" {
					comments = append(comments, comment)
				}
			}
		}
	}
	return comments
}

func (parser TransactionParser) ParseComment(payload *cell.Cell) string {
	l := payload.BeginParse()
	if val, err := l.LoadUInt(32); err == nil && val == 0 {
		str, _ := l.LoadStringSnake()
		return str
	}
	return ""
}
