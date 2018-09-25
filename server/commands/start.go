package commands

import (
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ethereum/go-ethereum/eth"
	"github.com/tendermint/tendermint/libs/cli"
	cmn "github.com/tendermint/tendermint/libs/common"

	"encoding/json"
	"github.com/CyberMiles/travis/app"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/sdk/dbm"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"github.com/CyberMiles/travis/version"
)

// GetStartCmd - initialize a command as the start command with tick
func GetStartCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start this full node",
		RunE:  startCmd(),
	}
	return startCmd
}

// nolint TODO: move to config file
const EyesCacheSize = 10000

//returns the start command which uses the tick
func startCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		rootDir := viper.GetString(cli.HomeFlag)

		if err := dbm.InitSqliter(path.Join(rootDir, "data", "travis.db")); err != nil {
			return err
		}

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
		srvs.emNode.Stop()
		dbm.Sqliter.CloseDB()
	})

	return nil
}

func createBaseApp(rootDir string, storeApp *app.StoreApp, ethApp *app.EthermintApplication, ethereum *eth.Ethereum) (*app.BaseApp, error) {
	app, err := app.NewBaseApp(storeApp, ethApp, ethereum)
	if err != nil {
		return nil, err
	}
	// if chain_id has not been set yet, load the genesis.
	// else, assume it's been loaded
	if app.GetChainID() == "" {
		// If genesis file exists, set key-value options
		genesisFile := path.Join(rootDir, DefaultConfig().TMConfig.GenesisFile())
		if _, err := os.Stat(genesisFile); err == nil {
			genDoc, err := loadGenesis(genesisFile)
			if err != nil {
				return nil, errors.Errorf("Error in LoadGenesis: %v\n", err)
			}

			app.SetChainId(genDoc.ChainID)
			utils.SetParams(genDoc.Params)
			for _, val := range genDoc.Validators {
				stake.SetGenesisValidator(val, app.Append())
			}
		} else {
			fmt.Printf("No genesis file at %s, skipping...\n", genesisFile)
		}
	}

	chainID := app.GetChainID()
	logger.Info("Starting Travis", "chain_id", chainID)

	return app, nil
}

func loadGenesis(filePath string) (*types.GenesisDoc, error) {
	bytes, err := cmn.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "loading genesis file")
	}

	genDoc := new(types.GenesisDoc)
	err = json.Unmarshal(bytes, genDoc)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling genesis file")
	}

	return genDoc, nil
}
