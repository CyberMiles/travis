package commands

import (
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/abci/server"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/cli"
	cmn "github.com/tendermint/tmlibs/common"
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/CyberMiles/travis/genesis"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/CyberMiles/travis/app"
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
	addStartFlag(startCmd)
	return startCmd
}

// nolint TODO: move to config file
const EyesCacheSize = 10000

//nolint
const (
	FlagAddress           = "address"
)

var (
	// Handler - use a global to store the handler, so we can set it in main.
	// TODO: figure out a cleaner way to register plugins
	Handler sdk.Handler
)

func init() {
	addStartFlag(StartCmd)
}

func addStartFlag(startCmd *cobra.Command) {
	flags := startCmd.Flags()
	flags.String(FlagAddress, "tcp://0.0.0.0:46658", "Listen address")
}

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
	// run just the abci app/server
	return startTravisABCI(basecoinApp)
}

func startTravisABCI(basecoinApp abci.Application) error {
	// Start the ABCI listener
	addr := viper.GetString(FlagAddress)
	svr, err := server.NewServer(addr, "socket", basecoinApp)
	if err != nil {
		return errors.Errorf("Error creating listener: %v\n", err)
	}
	svr.SetLogger(logger.With("module", "abci-server"))
	svr.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})
	return nil
}

