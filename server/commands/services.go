package commands

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/node"
	pv "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"

	"github.com/CyberMiles/travis/api"
	"github.com/CyberMiles/travis/app"
	travisUtils "github.com/CyberMiles/travis/utils"
	"github.com/CyberMiles/travis/vm/cmd/utils"
	emtUtils "github.com/CyberMiles/travis/vm/cmd/utils"
	"github.com/CyberMiles/travis/vm/ethereum"
)

type Services struct {
	backend *api.Backend
	tmNode  *node.Node
	emNode  *ethereum.Node
}

func startServices(rootDir string, storeApp *app.StoreApp) (*Services, error) {
	// Setup the go-ethereum node and start it
	emNode := emtUtils.MakeFullNode(context)
	startNode(context, emNode)

	// Fetch the registered service of this type
	var backend *api.Backend
	if err := emNode.Service(&backend); err != nil {
		ethUtils.Fatalf("ethereum backend service not running: %v", err)
	}

	// In-proc RPC connection so ABCI.Query can be forwarded over the ethereum rpc
	rpcClient, err := emNode.Attach()
	if err != nil {
		ethUtils.Fatalf("Failed to attach to the inproc geth: %v", err)
	}

	// Create the ABCI app
	ethApp, err := app.NewEthermintApplication(backend, rpcClient, nil)
	if err != nil {
		log.Warn(err.Error())
		os.Exit(1)
	}
	ethApp.SetLogger(emtUtils.EthermintLogger().With("module", "vm"))

	// Alter database if needed
	if err = alterDatabaseIfNeeded(rootDir); err != nil {
		log.Warn(err.Error())
		os.Exit(1)
	}

	// Alter database if needed
	if err = alterDatabaseIfNeeded2(rootDir); err != nil {
		log.Warn(err.Error())
		os.Exit(1)
	}

	// Create Basecoin app
	basecoinApp, err := createBaseApp(rootDir, storeApp, ethApp, backend.Ethereum())
	if err != nil {
		log.Warn(err.Error())
		os.Exit(1)
	}

	// Create & start tendermint node
	tmNode, err := startTendermint(basecoinApp)
	if err != nil {
		log.Warn(err.Error())
		os.Exit(1)
	}
	backend.SetTMNode(tmNode)

	return &Services{backend, tmNode, emNode}, nil
}

// startNode copies the logic from go-ethereum
func startNode(ctx *cli.Context, stack *ethereum.Node) {
	emtUtils.StartNode(stack)

	// Unlock any account specifically requested
	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	passwords := ethUtils.MakePasswordList(ctx)
	unlocks := strings.Split(ctx.GlobalString(ethUtils.UnlockedAccountFlag.Name), ",")
	for i, account := range unlocks {
		if trimmed := strings.TrimSpace(account); trimmed != "" {
			utils.UnlockAccount(ctx, ks, trimmed, i, passwords)
		}
	}
	// Register wallet event handlers to open and auto-derive wallets
	events := make(chan accounts.WalletEvent, 16)
	stack.AccountManager().Subscribe(events)

	go func() {
		// Create an chain state reader for self-derivation
		rpcClient, err := stack.Attach()
		if err != nil {
			ethUtils.Fatalf("Failed to attach to self: %v", err)
		}
		stateReader := ethclient.NewClient(rpcClient)

		// Open and self derive any wallets already attached
		for _, wallet := range stack.AccountManager().Wallets() {
			if err := wallet.Open(""); err != nil {
				log.Warn("Failed to open wallet", "url", wallet.URL(), "err", err)
			} else {
				wallet.SelfDerive(accounts.DefaultBaseDerivationPath, stateReader)
			}
		}
		// Listen for wallet event till termination
		for event := range events {
			if event.Kind == accounts.WalletArrived {
				if err := event.Wallet.Open(""); err != nil {
					log.Warn("New wallet appeared, failed to open", "url",
						event.Wallet.URL(), "err", err)
				} else {
					status, _ := event.Wallet.Status()
					log.Info("New wallet appeared", "url", event.Wallet.URL(),
						"status", status)
					event.Wallet.SelfDerive(accounts.DefaultBaseDerivationPath,
						stateReader)
				}
			} else {
				log.Info("Old wallet dropped", "url", event.Wallet.URL())
				event.Wallet.Close()
			}
		}
	}()
}

func startTendermint(basecoinApp abcitypes.Application) (*node.Node, error) {
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return nil, err
	}

	var papp proxy.ClientCreator
	if basecoinApp != nil {
		papp = proxy.NewLocalClientCreator(basecoinApp)
	} else {
		papp = proxy.DefaultClientCreator(cfg.ProxyApp, cfg.ABCI, cfg.DBDir())
	}

	// Create & start tendermint node
	n, err := node.NewNode(cfg,
		pv.LoadOrGenFilePV(cfg.PrivValidatorFile()),
		papp,
		node.DefaultGenesisDocProviderFunc(cfg),
		node.DefaultDBProvider,
		node.DefaultMetricsProvider,
		logger.With("module", "node"))
	if err != nil {
		return nil, err
	}

	err = n.Start()
	if err != nil {
		return nil, err
	}

	return n, nil
}

func alterDatabaseIfNeeded(rootDir string) error {
	stakeDbPath := filepath.Join(rootDir, "data", travisUtils.DB_FILE_NAME)
	db, err := sql.Open("sqlite3", stakeDbPath)
	if err != nil {
		return err
	}

	defer db.Close()

	var cnt int64
	err = db.QueryRow("SELECT COUNT(*) AS cnt FROM pragma_table_info('delegations') WHERE name='source'").Scan(&cnt)
	if err != nil {
		return err
	}

	if cnt == 1 {
		return nil
	}

	sqlStmt := "alter table delegations add column source text not null default 'cube'"
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	log.Info("Successfully altered database #1!")

	return nil
}

func alterDatabaseIfNeeded2(rootDir string) error {
	stakeDbPath := filepath.Join(rootDir, "data", travisUtils.DB_FILE_NAME)
	db, err := sql.Open("sqlite3", stakeDbPath)
	if err != nil {
		return err
	}

	defer db.Close()

	// add the completely_withdraw field to delegations table
	var cnt int64
	err = db.QueryRow("SELECT COUNT(*) AS cnt FROM pragma_table_info('delegations') WHERE name='completely_withdraw'").Scan(&cnt)
	if err != nil {
		return err
	}

	if cnt == 1 {
		return nil
	}

	sqlStmt := "alter table delegations add column completely_withdraw text not null default 'N'"
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	// add the actual_amount field to unstake_requests table
	err = db.QueryRow("SELECT COUNT(*) AS cnt FROM pragma_table_info('unstake_requests') WHERE name='actual_amount'").Scan(&cnt)
	if err != nil {
		return err
	}

	if cnt == 1 {
		return nil
	}

	sqlStmt = "alter table unstake_requests add column actual_amount text not null default '0'"
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return err
	}

	log.Info("Successfully altered database #2!")

	return nil
}
