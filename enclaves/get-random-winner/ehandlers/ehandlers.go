package ehandlers

import (
	"context"
	"enclave/emessages"
	"enclave/eprojects"
	"enclave/erand"
	"errors"
	"github.com/tonteeton/golib/eresp"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"math"
)

// Config contains configuration parameters for generating an enclave response.
type Config struct {
	SenderWallet    *wallet.Wallet
	ContractAddress *address.Address
	Response        eresp.Config
}

type Handlers struct {
	config        Config
	projects      eprojects.Projects
	revealedValue emessages.RevealedValue
}

func Init(config Config) (*Handlers, error) {
	projects, err := eprojects.LoadProjects("./buidls.json")
	if err != nil {
		return nil, err
	}
	return &Handlers{config: config, projects: projects}, nil
}

func (handlers *Handlers) RandomCommit(tx *tlb.Transaction) error {
	cfg := handlers.config
	randIndex, err := erand.RandUInt64(uint64(handlers.projects.Len()))
	if err != nil {
		return err
	}
	project := handlers.projects.GetByIndex(int(randIndex))
	if project == nil {
		return errors.New("can't get project by index")
	}

	randNonce, err := erand.RandUInt64(math.MaxUint64)
	if err != nil {
		return err
	}

	handlers.revealedValue = emessages.RevealedValue{
		Timestamp: tx.Now,
		Recipient: handlers.config.ContractAddress,
		Nonce:     randNonce,
		DoraID:    uint64(project.ID),
		Name:      project.Name,
	}

	resp := emessages.RandomCommit{
		Timestamp: handlers.revealedValue.Timestamp,
		Recipient: handlers.revealedValue.Recipient,
		ValueHash: handlers.revealedValue.Hash(),
	}
	responseCell, err := eresp.PackResponseToCell(cfg.Response, resp.ToCell(), resp.GetOpcode())
	if err != nil {
		return err
	}

	return sendResponse(context.Background(), cfg.SenderWallet, cfg.ContractAddress, responseCell)
}

func (handlers *Handlers) RandomReveal(tx *tlb.Transaction) error {
	if len(tx.Hash) != 32 {
		return errors.New("unexpected transaction hash size")
	}
	cfg := handlers.config
	resp := emessages.RandomReveal{
		//DoraID: uint64(project.ID),
		//Name: project.Name,
		DoraID:          handlers.revealedValue.DoraID,
		Name:            handlers.revealedValue.Name,
		RevealTimestamp: tx.Now,
		Nonce:           handlers.revealedValue.Nonce,
		TxHash:          tx.Hash,
	}

	responseCell, err := eresp.PackResponseToCell(cfg.Response, resp.ToCell(), resp.GetOpcode())
	if err != nil {
		return err
	}
	return sendResponse(context.Background(), cfg.SenderWallet, cfg.ContractAddress, responseCell)
}

func sendResponse(ctx context.Context, senderWallet *wallet.Wallet, address *address.Address, payload *cell.Cell) error {
	msg := wallet.SimpleMessage(address, tlb.MustFromTON("0.025"), payload)

	tx, _, err := senderWallet.SendWaitTransaction(ctx, msg)
	if err != nil {
		return err
	}
	_ = tx
	return nil
}
