package main

import (
	"context"
	"enclave/appconf"
	"enclave/ehandlers"
	"enclave/txparser"
	"fmt"
	"github.com/tonteeton/golib/eresp"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"log"
	"os"
	"slices"

	"errors"
	"flag"
	"github.com/tonteeton/golib/eattest"
	"github.com/tonteeton/golib/ereport"
)

func watchTransactions(cfg *appconf.Config) error {
	contractAddress := cfg.Network.ContractAddress

	client := liteclient.NewConnectionPool()
	clientCfg, err := liteclient.GetConfigFromUrl(context.Background(), cfg.Network.GlobalConfigURL)
	if err != nil {
		return err
	}
	err = client.AddConnectionsFromConfig(context.Background(), clientCfg)
	if err != nil {
		return err
	}

	api := ton.NewAPIClient(client).WithRetry()
	api.SetTrustedBlockFromConfig(clientCfg)

	senderWallet, err := wallet.FromSeed(api, cfg.Wallet.Mnemonic, wallet.V3R2)
	if err != nil {
		return err
	}

	log.Println("fetching and checking proofs since config init block...")
	master, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		return err
	}

	acc, err := api.GetAccount(context.Background(), master, contractAddress)
	if err != nil {
		return err
	}

	lastProcessedLT := acc.LastTxLT
	transactions := make(chan *tlb.Transaction)
	go api.SubscribeOnTransactions(context.Background(), contractAddress, lastProcessedLT, transactions)

	handlers, err := ehandlers.Init(
		ehandlers.Config{
			SenderWallet:    senderWallet,
			ContractAddress: contractAddress,
			Response: eresp.Config{
				Response:      cfg.Response,
				SignatureKeys: cfg.SignatureKeys,
			},
		},
	)
	if err != nil {
		return err
	}

	txParser := txparser.TransactionParser{
		TestNet: cfg.Network.TestNet,
		Address: contractAddress,
	}

	log.Println("waiting for transactions...")
	for tx := range transactions {
		comments := txParser.ParseExternalComments(tx)
		if slices.Contains(comments, "random()") {
			log.Println("random() command detected")
			err := handlers.RandomCommit(tx)
			if err != nil {
				return err
			}
		}
		if slices.Contains(comments, "reveal()") {
			log.Println("raveal() command detected")
			err := handlers.RandomReveal(tx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func executeReportFunc(fn func(ereport.Config, eattest.Attestation) error, cfg *appconf.Config) error {
	reportCfg := ereport.Config{
		Reports:        cfg.Reports,
		SignatureKeys:  cfg.SignatureKeys,
		EncryptionKeys: cfg.EncryptionKeys,
	}
	attest := eattest.NewAttestation()
	return fn(reportCfg, attest)
}

func main() {
	cfg, err := appconf.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	flag.Usage = func() {
		fmt.Println("Usage: [command]")
		fmt.Println("Commands:")
		fmt.Println("  watch            Watch for incoming transactions and handle them")
		fmt.Println("  report-key       Generate SGX-signed report with public keys")
		fmt.Println("  import-key       Import encrypted signature Private key")
		fmt.Println("  export-key       Export encrypted signature Private key")
	}

	cmds := map[string]func(cfg *appconf.Config) error{
		"watch":      watchTransactions,
		"report-key": func(cfg *appconf.Config) error { return executeReportFunc(ereport.ExportPublicKeys, cfg) },
		"import-key": func(cfg *appconf.Config) error { return executeReportFunc(ereport.ImportPrivateSignature, cfg) },
		"export-key": func(cfg *appconf.Config) error { return executeReportFunc(ereport.ExportPrivateSignature, cfg) },
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	cmd, ok := cmds[os.Args[1]]
	if ok {
		err = cmd(cfg)
	} else {
		flag.Usage()
		err = errors.New("Unknown command.")
	}

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
