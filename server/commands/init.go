package commands

import (
	"math/big"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"database/sql"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	emtUtils "github.com/CyberMiles/travis/vm/cmd/utils"
	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tendermint/tendermint/libs/cli"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/p2p"
	pv "github.com/tendermint/tendermint/privval"
	"os/exec"
)

const (
	FlagChainID   = "chain-id"
	FlagENV       = "env"
	FlagVMGenesis = "vm-genesis"

	defaultEnv = "private"
)

var InitCmd = GetInitCmd()

func GetInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize",
		RunE:  initFiles,
	}
	initCmd.Flags().String(FlagChainID, "local", "Chain ID")
	initCmd.Flags().String(FlagENV, defaultEnv, "Environment (mainnet|staging|testnet|private)")
	initCmd.Flags().String(FlagVMGenesis, "", "VM genesis file")
	return initCmd
}

func initFiles(cmd *cobra.Command, args []string) error {
	initTendermint()
	initTravisDb()
	// initTravisCmd()
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
		genDoc := types.GenesisDoc{
			ChainID: viper.GetString(FlagChainID),
			Params:  utils.DefaultParams(),
		}

		genDoc.Validators = []types.GenesisValidator{{
			PubKey:    types.PubKey{privValidator.GetPubKey()},
			Power:     "1",
			Shares:    1000000,
			Address:   "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc",
			CompRate:  sdk.NewRat(2, 10),
			MaxAmount: 10000000,
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			panic(err)
		}
		logger.Info("Genetated genesis file", "path", genFile)
	}
}

func initEthermint() error {
	genesisPath := viper.GetString(FlagVMGenesis)
	genesis, err := emtUtils.ParseGenesisOrDefault(genesisPath, config.EMConfig.ChainId)
	if err != nil {
		ethUtils.Fatalf("genesisJSON err: %v", err)
	}
	// override ethermint's chain_id
	genesis.Config.ChainID = new(big.Int).SetUint64(uint64(config.EMConfig.ChainId))

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

func initTravisDb() {
	rootDir := viper.GetString(cli.HomeFlag)
	stakeDbPath := filepath.Join(rootDir, "data", utils.DB_FILE_NAME)

	if _, err := os.OpenFile(stakeDbPath, os.O_RDONLY, 0444); err != nil {
		db, err := sql.Open("sqlite3", stakeDbPath)
		if err != nil {
			ethUtils.Fatalf("Initializing cybermiles database: %s", err.Error())
		}
		defer db.Close()

		sqlStmt := `
	create table candidates(id integer not null primary key autoincrement, address text not null, pub_key text not null, shares text not null default '0', voting_power integer default 0, pending_voting_power integer default 0, max_shares text not null default '0', comp_rate text not null default '0', name text not null default '', website text not null default '', location text not null default '', email text not null default '', profile text not null default '', verified text not null default 'N', active text not null default 'Y', rank integer not null default 0, state text not null default '', hash text not null default '', block_height integer not null, num_of_delegators integer not null default 0, created_at text not null default '');
	create unique index idx_candidates_pub_key on candidates(pub_key);
	create unique index idx_candidates_address on candidates(address);
	create index idx_candidates_hash on candidates(hash);
	create table delegations(id integer not null primary key autoincrement, delegator_address text not null, candidate_id integer not null, delegate_amount text not null default '0', award_amount text not null default '0', withdraw_amount text not null default '0', pending_withdraw_amount text not null default '0', slash_amount text not null default '0', comp_rate text not null default '0', hash text not null default '',  voting_power integer not null default 0, state text not null default 'Y', block_height integer not null, average_staking_date integer not null default 0, created_at text not null);
	create unique index idx_delegations_delegator_address_candidate_id on delegations(delegator_address, candidate_id);
	create index idx_delegations_hash on delegations(hash);
 	create table delegate_history(id integer not null primary key autoincrement, delegator_address text not null, candidate_id integer not null, amount text not null default '0', op_code text not null default '', block_height integer not null, hash text not null default '');
	create index idx_delegate_history_delegator_address on delegate_history(delegator_address);
	create index idx_delegate_history_candidate_id on delegate_history(candidate_id);
	create index idx_delegate_history_hash on delegate_history(hash);
	create table slashes(id integer not null primary key autoincrement, candidate_id integer not null, slash_ratio integer default 0, slash_amount text not null, reason text not null default '', created_at text not null, block_height integer not null, hash text not null default '');
	create index idx_slashes_candidate_id on slashes(candidate_id);
	create index idx_slashes_hash on slashes(hash);
 	create table unstake_requests(id integer not null primary key autoincrement, delegator_address text not null, candidate_id integer not null, initiated_block_height integer default 0, performed_block_height integer default 0, amount text not null default '0', state text not null default 'PENDING', hash text not null default '');
 	create index idx_unstake_requests_delegator_address on unstake_requests(delegator_address);
 	create table candidate_daily_stakes(id integer not null primary key autoincrement, candidate_id integer not null, amount text not null default '0', block_height integer not null, hash text not null default '');
	create index idx_candidate_daily_stakes_candidate_id on candidate_daily_stakes(candidate_id);
	create index idx_candidate_daily_stakes_hash on candidate_daily_stakes(hash);
	create table candidate_account_update_requests(id integer primary key autoincrement, candidate_id integer not null, from_address text not null, to_address text not null, created_block_height integer not null, accepted_block_height integer not null, state text not null, hash text not null default '');
	create index idx_candidate_account_update_requests_to_address on candidate_account_update_requests(to_address);
	create index idx_candidate_account_update_requests_hash on candidate_account_update_requests(hash);

 	create table governance_proposal(id text not null primary key, type text not null, proposer text not null, block_height integer not null, expire_timestamp integer not null, expire_block_height integer not null, hash text not null default '', result text not null default '', result_msg text not null default '', result_block_height integer not null default 0);
	create index idx_governance_proposal_hash on governance_proposal(hash);
 	create table governance_transfer_fund_detail(proposal_id text not null, from_address text not null, to_address text not null, amount text not null, reason text not null);
	create index idx_governance_transfer_fund_detail_proposal_id on governance_transfer_fund_detail(proposal_id);
 	create table governance_change_param_detail(proposal_id text not null, param_name text not null, param_value text not null, reason text not null);
	create index idx_governance_change_param_detail_proposal_id on governance_change_param_detail(proposal_id);
	create table governance_deploy_libeni_detail(proposal_id text not null, name text not null, version text not null, fileurl text not null, md5 text not null, reason text not null, status text not null);
	create index idx_governance_deploy_libeni_detail_proposal_id on governance_deploy_libeni_detail(proposal_id);
	create table governance_retire_program_detail(proposal_id text not null, retired_version text not null, preserved_validators text not null, reason text not null);
	create index idx_governance_retire_program_detail_proposal_id on governance_retire_program_detail(proposal_id);
	create table governance_upgrade_program_detail(proposal_id text not null, retired_version text not null, name text not null, version text not null, fileurl text not null, md5 text not null, reason text not null);
	create index idx_governance_upgrade_program_detail_proposal_id on governance_retire_program_detail(proposal_id);
 	create table governance_vote(proposal_id text not null, voter text not null, block_height integer not null, answer text not null,  hash text not null default '', unique(proposal_id, voter) ON conflict replace);
	create index idx_governance_vote_voter on governance_vote(voter);
	create index idx_governance_vote_proposal_id on governance_vote(proposal_id);
	create index idx_governance_vote_hash on governance_vote(hash);
	`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			//os.Remove(stakeDbPath)
			ethUtils.Fatalf("Create travis database tables: %s", err.Error())
		}
		log.Info("Successfully init travis database and create tables!")
	} else {
		log.Warn("The travis database already exists!")
	}
}

func initTravisCmd() {
	rootDir := viper.GetString(cli.HomeFlag)
	binPath := filepath.Join(rootDir, "bin")
	if err := cmn.EnsureDir(binPath, 0700); err != nil {
		cmn.PanicSanity(err.Error())
	}
	execPath, _ := exec.LookPath(os.Args[0])
	if err := exec.Command("cp", execPath, binPath).Run(); err != nil {
		log.Error("copy bin file error %s", execPath, err)
	}
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
