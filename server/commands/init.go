package commands

import (
	"math/big"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/tendermint/tendermint/p2p"
	pv "github.com/tendermint/tendermint/types/priv_validator"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/CyberMiles/travis/types"
	emtUtils "github.com/CyberMiles/travis/vm/cmd/utils"
)

var (
	FlagChainID = "chain-id"
)

var InitCmd = GetInitCmd()

func GetInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize",
		RunE:  initFiles,
	}
	initCmd.Flags().String(FlagChainID, "local", "Chain ID")
	return initCmd
}

func initFiles(cmd *cobra.Command, args []string) error {
	initTendermint()
	return initEthermint()
}

func initTendermint() {
	// private validator
	privValFile := config.TMConfig.PrivValidatorFile()
	var privValidator *pv.FilePV
	if cmn.FileExists(privValFile) {
		privValidator = pv.LoadFilePV(privValFile)
		logger.Info("Found private validator", "path", privValFile)
	} else {
		dir := filepath.Dir(privValFile)
		os.Mkdir(dir, os.ModePerm)
		privValidator = pv.GenFilePV(privValFile)
		privValidator.Save()
		logger.Info("Genetated private validator", "path", privValFile)
	}

	nodeKeyFile := config.TMConfig.NodeKeyFile()
	if cmn.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			panic(err)
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}

	// genesis file
	genFile := config.TMConfig.GenesisFile()
	if cmn.FileExists(genFile) {
		logger.Info("Found genesis file", "path", genFile)
	} else {
		genDoc := GenesisDoc{
			ChainID:          viper.GetString(FlagChainID),
			MaxVals:          4,
			SelfStakingRatio: "0.1",
		}
		genDoc.Validators = []types.GenesisValidator{{
			PubKey:    types.PubKey{privValidator.GetPubKey()},
			Power:     1000,
			Address:   "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
			CompRate:  "0.5",
			MaxAmount: 10000,
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			panic(err)
		}
		logger.Info("Genetated genesis file", "path", genFile)
	}
}

func initEthermint() error {
	genesisPath := context.Args().First()
	genesis, err := emtUtils.ParseGenesisOrDefault(genesisPath)
	if err != nil {
		ethUtils.Fatalf("genesisJSON err: %v", err)
	}
	// override ethermint's chain_id
	genesis.Config.ChainId = new(big.Int).SetUint64(uint64(config.EMConfig.EthChainId))

	ethermintDataDir := emtUtils.MakeDataDir(context)

	chainDb, err := ethdb.NewLDBDatabase(filepath.Join(ethermintDataDir,
		"vm/chaindata"), 0, 0)
	if err != nil {
		ethUtils.Fatalf("could not open database: %v", err)
	}

	_, hash, err := core.SetupGenesisBlock(chainDb, genesis)
	if err != nil {
		ethUtils.Fatalf("failed to write genesis block: %v", err)
	}

	log.Info("successfully wrote genesis block and/or chain rule set", "hash", hash)

	// As per https://github.com/tendermint/ethermint/issues/244#issuecomment-322024199
	// Let's implicitly add in the respective keystore files
	// to avoid manually doing this step:
	// $ cp -r $GOPATH/src/github.com/tendermint/ethermint/setup/keystore $(DATADIR)
	keystoreDir := filepath.Join(ethermintDataDir, "keystore")
	if err := os.MkdirAll(keystoreDir, 0777); err != nil {
		ethUtils.Fatalf("mkdirAll keyStoreDir: %v", err)
	}

	for filename, content := range keystoreFilesMap {
		storeFileName := filepath.Join(keystoreDir, filename)
		f, err := os.Create(storeFileName)
		if err != nil {
			log.Error("create %q err: %v", storeFileName, err)
			continue
		}
		if _, err := f.Write([]byte(content)); err != nil {
			log.Error("write content %q err: %v", storeFileName, err)
		}
		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}

var keystoreFilesMap = map[string]string{
	// https://github.com/tendermint/ethermint/blob/edc95f9d47ba1fb7c8161182533b5f5d5c5d619b/setup/keystore/UTC--2016-10-21T22-30-03.071787745Z--7eff122b94897ea5b0e2a9abf47b86337fafebdc
	// OR
	// $GOPATH/src/github.com/ethermint/setup/keystore/UTC--2016-10-21T22-30-03.071787745Z--7eff122b94897ea5b0e2a9abf47b86337fafebdc
	"UTC--2016-10-21T22-30-03.071787745Z--7eff122b94897ea5b0e2a9abf47b86337fafebdc": `
{
  "address":"7eff122b94897ea5b0e2a9abf47b86337fafebdc",
  "id":"f86a62b4-0621-4616-99af-c4b7f38fcc48","version":3,
  "crypto":{
    "cipher":"aes-128-ctr","ciphertext":"19de8a919e2f4cbdde2b7352ebd0be8ead2c87db35fc8e4c9acaf74aaaa57dad",
    "cipherparams":{"iv":"ba2bd370d6c9d5845e92fbc6f951c792"},
    "kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"c7cc2380a96adc9eb31d20bd8d8a7827199e8b16889582c0b9089da6a9f58e84"},
    "mac":"ff2c0caf051ca15d8c43b6f321ec10bd99bd654ddcf12dd1a28f730cc3c13730"
  }
}
`,
	"UTC--2018-04-09T09-48-47.241470000Z--77beb894fc9b0ed41231e51f128a347043960a9d": `
{"address":"77beb894fc9b0ed41231e51f128a347043960a9d","crypto":{"cipher":"aes-128-ctr","ciphertext":"a559667b38ab5b38aeadd22aef5b0582ef28bc86e9899d058ec41a5f8193ffd6","cipherparams":{"iv":"0efe57fb91b2eaf3f7c630037c531b13"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"21201b8cb9ad02be8744c44f82a6668d30eb9c2811b6038d9727bbd1034015aa"},"mac":"9f9e1e1877f55f22f8ca3fb36b0667caf02db3bc98717e62b40df18bdb33e766"},"id":"745b91e0-e760-4fa9-97b2-22a76745a25f","version":3}
`,
}
