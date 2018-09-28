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
	"net"
	"net/rpc"
	"net/http"
	"log"
	"time"
)

const (
	MonitorFlag = "monitor"
	RpcPort = "26650"
)

// GetStartCmd - initialize a command as the start command with tick
func GetStartCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start this full node",
		RunE:  startCmd(),
	}
	startCmd.PersistentFlags().Bool(MonitorFlag, false, "start travis as monitor mode")
	return startCmd
}

// nolint TODO: move to config file
const EyesCacheSize = 10000

//returns the start command which uses the tick
func startCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		rootDir := viper.GetString(cli.HomeFlag)
		// start with monitor mode
		if monitorFlag := viper.GetBool(MonitorFlag); monitorFlag {
			return startSubProcess(rootDir)
		}

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

func startSubProcess(rootDir string) error {
	arg := "--" + MonitorFlag

	args := make([]string,0)
	var index int
	for index = 1; index < len(os.Args); index++ {
		if os.Args[index] == arg {
			break
		}
		args = append(args, os.Args[index])
	}
	args = append(args, os.Args[index+1:]...)
	fmt.Println(args)
	fmt.Println(os.Args)
	cmd := types.NewTravisCmd(rootDir, path.Base(os.Args[0]), args...)
	m := types.NewMonitor(cmd)
	startRPC(m)
	cmd.Start()

	go startRoutine(cmd)


	cmn.TrapSignal(func() {
		fmt.Println("Stopping the command ...", cmd.Cmd().Process.Pid)
		cmd.Stop()
		time.Sleep(time.Second * 1)
	})

	return nil
}


func startRPC(m *types.Monitor) error {
	rpc.Register(m)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", "127.0.0.1:"+RpcPort)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
	return nil
}

func startRoutine(c *types.TravisCmd) {
	for {
		select {
		case cmdName := <-c.DownloadChan:
			fmt.Printf("Start to download %s\n", cmdName)
			if err := c.Download(cmdName); err != nil {
				log.Fatalf("Download failed: %s\n", err)
			}
		case cmdInfo := <-c.UpgradeChan:
			fmt.Printf("Start to upgrade %s\n", cmdInfo.Name)
			if c.NextName != cmdInfo.Name {
				log.Fatalf("Upgrade want version (%s) but get version: (%s)\n", cmdInfo.Name, c.NextName)
			}
			if err := c.Upgrade(cmdInfo); err != nil {
				log.Fatalf("Upgrade failed: %s\n", err)
			}
		case <-c.KillChan:
			if err := c.Kill(); err != nil {
				log.Fatalf("Kill process failed: %s\n", err)
			}
		}
	}
}
