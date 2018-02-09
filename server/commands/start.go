package commands

import (
	"fmt"
	"os"
	"path"
	"time"
	"math/big"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tmlibs/cli"
	cmn "github.com/tendermint/tmlibs/common"
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/genesis"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/CyberMiles/travis/app"

	"github.com/ethereum/go-ethereum/common"
)

// StartCmd - command to start running the abci app (and tendermint)!
var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start this full node",
	RunE:  startCmd,
}

// GetTickStartCmd - initialize a command as the start command with tick
func GetTickStartCmd(tick sdk.Ticker) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start this full node",
		RunE:  startCmd,
	}
	startCmd.RunE = tickStartCmd(tick)
	return startCmd
}

// nolint TODO: move to config file
const EyesCacheSize = 10000


var (
	// Handler - use a global to store the handler, so we can set it in main.
	// TODO: figure out a cleaner way to register plugins
	Handler sdk.Handler
)

//returns the start command which uses the tick
func tickStartCmd(clock sdk.Ticker) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		rootDir := viper.GetString(cli.HomeFlag)

		cmdName := cmd.Root().Name()
		appName := fmt.Sprintf("%s v%v", cmdName, version.Version)
		storeApp, err := app.NewStoreApp(
			appName,
			path.Join(rootDir, "data", "merkleeyes.db"),
			EyesCacheSize,
			logger.With("module", "app"))
		if err != nil {
			return err
		}

		// Create Basecoin app
		basecoinApp := app.NewBaseApp(storeApp, Handler, clock)
		return start(rootDir, basecoinApp)
	}
}

func startCmd(cmd *cobra.Command, args []string) error {
	rootDir := viper.GetString(cli.HomeFlag)

	cmdName := cmd.Root().Name()
	appName := fmt.Sprintf("%s v%v", cmdName, version.Version)
	storeApp, err := app.NewStoreApp(
		appName,
		path.Join(rootDir, "data", "merkleeyes.db"),
		EyesCacheSize,
		logger.With("module", "app"))
	if err != nil {
		return err
	}

	// Create Basecoin app
	basecoinApp := app.NewBaseApp(storeApp, Handler, nil)
	return start(rootDir, basecoinApp)
}

func start(rootDir string, basecoinApp *app.BaseApp) error {

	// if chain_id has not been set yet, load the genesis.
	// else, assume it's been loaded
	if basecoinApp.GetChainID() == "" {
		// If genesis file exists, set key-value options
		genesisFile := path.Join(rootDir, "genesis.json")
		if _, err := os.Stat(genesisFile); err == nil {
			err = genesis.Load(basecoinApp, genesisFile)
			if err != nil {
				return errors.Errorf("Error in LoadGenesis: %v\n", err)
			}
		} else {
			fmt.Printf("No genesis file at %s, skipping...\n", genesisFile)
		}
	}

	chainID := basecoinApp.GetChainID()
	logger.Info("Starting Travis", "chain_id", chainID)

	srvs, err := startServices(rootDir, basecoinApp)
	if err != nil {
		return errors.Errorf("Error in start services: %v\n", err)
	}

// test change balance -->
state, err := srvs.backend.Ethereum().BlockChain().State()
if err != nil {
	return errors.Errorf("Error in get state: %v\n", err)
}
addr := "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc"
fmt.Printf("===================== balance before set: %v\n", state.GetBalance(common.HexToAddress(addr)))
state.SetBalance(common.HexToAddress(addr), big.NewInt(int64(111)))
fmt.Printf("===================== balance after set: %v\n", state.GetBalance(common.HexToAddress(addr)))
// <---

	// wait forever
	cmn.TrapSignal(func() {
	  // cleanup
	  srvs.emt.Stop()
	  for {
	    pauseDuration := 1 * time.Second
	    time.Sleep(pauseDuration)
	    if !srvs.emt.IsRunning() {
	      break
	    }
	  }
	  srvs.tmNode.Stop()
	})

	return nil
}
