package commands

import (
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/tendermint/tmlibs/cli"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/CyberMiles/travis/app"
	"github.com/CyberMiles/travis/genesis"
	ethapp "github.com/CyberMiles/travis/modules/vm/app"
)

// GetTickStartCmd - initialize a command as the start command with tick
func GetTickStartCmd(tick sdk.Ticker) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start this full node",
		RunE:  tickStartCmd(tick),
	}
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

		return start(rootDir, storeApp)
	}
}

func start(rootDir string, storeApp *app.StoreApp) error {
	srvs, err := startServices(rootDir, storeApp)
	if err != nil {
		return errors.Errorf("Error in start services: %v\n", err)
	}

	// wait forever
	cmn.TrapSignal(func() {
		srvs.tmNode.Stop()
	})

	return nil
}

func createBaseCoinApp(rootDir string, storeApp *app.StoreApp, ethApp *ethapp.EthermintApplication) (*app.BaseApp, error) {
	basecoinApp, err := app.NewBaseApp(storeApp, ethApp, Handler, nil)
	if err != nil {
		return nil, err
	}
	// if chain_id has not been set yet, load the genesis.
	// else, assume it's been loaded
	//if basecoinApp.GetChainID() == "" {
		// If genesis file exists, set key-value options
		genesisFile := path.Join(rootDir, "genesis.json")
		if _, err := os.Stat(genesisFile); err == nil {
			err = genesis.Load(basecoinApp, genesisFile)
			if err != nil {
				return nil, errors.Errorf("Error in LoadGenesis: %v\n", err)
			}
		} else {
			fmt.Printf("No genesis file at %s, skipping...\n", genesisFile)
		}
	//}

	chainID := basecoinApp.GetChainID()
	logger.Info("Starting Travis", "chain_id", chainID)

	return basecoinApp, nil
}
