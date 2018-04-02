package app

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/errors"
	sm "github.com/cosmos/cosmos-sdk/state"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"
	"database/sql"
	"os"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tendermint/go-wire"
	"github.com/ethereum/go-ethereum/common"
	"github.com/CyberMiles/travis/modules/stake"
)

// DefaultHistorySize is how many blocks of history to store for ABCI queries
const DefaultHistorySize = 10

// StoreApp contains a data store and all info needed
// to perform queries and handshakes.
//
// It should be embeded in another struct for CheckTx,
// DeliverTx and initializing state from the genesis.
type StoreApp struct {
	// Name is what is returned from info
	Name string

	// this is the database state
	info  *sm.ChainState
	state *sm.State

	// cached validator changes from DeliverTx
	pending []*abci.Validator

	// height is last committed block, DeliverTx is the next one
	height int64

	logger log.Logger
}

// TODO should satisfy?
//var _ abci.Application = &StoreApp{}

// NewStoreApp creates a data store to handle queries
func NewStoreApp(appName, dbName string, cacheSize int, logger log.Logger) (*StoreApp, error) {
	state, err := loadState(dbName, cacheSize, DefaultHistorySize)
	if err != nil {
		return nil, err
	}

	err = initStakeDb()
	if err != nil {
		return nil, err
	}

	app := &StoreApp{
		Name:   appName,
		state:  state,
		height: state.LatestHeight(),
		info:   sm.NewChainState(),
		logger: logger.With("module", "app"),
	}
	return app, nil
}

// MockStoreApp returns a Store app with no persistence
func MockStoreApp(appName string, logger log.Logger) (*StoreApp, error) {
	return NewStoreApp(appName, "", 0, logger)
}

// GetChainID returns the currently stored chain
func (app *StoreApp) GetChainID() string {
	return app.info.GetChainID(app.state.Committed())
}

// Logger returns the application base logger
func (app *StoreApp) Logger() log.Logger {
	return app.logger
}

// Hash gets the last hash stored in the database
func (app *StoreApp) Hash() []byte {
	return app.state.LatestHash()
}

// Committed returns the committed state,
// also exposing historical queries
// func (app *StoreApp) Committed() *Bonsai {
// 	return app.state.committed
// }

// Append returns the working state for DeliverTx
func (app *StoreApp) Append() sm.SimpleDB {
	return app.state.Append()
}

// Check returns the working state for CheckTx
func (app *StoreApp) Check() sm.SimpleDB {
	return app.state.Check()
}

// CommittedHeight gets the last block height committed
// to the db
func (app *StoreApp) CommittedHeight() int64 {
	return app.height
}

// WorkingHeight gets the current block we are writing
func (app *StoreApp) WorkingHeight() int64 {
	return app.height + 1
}

// Info implements abci.Application. It returns the height and hash,
// as well as the abci name and version.
//
// The height is the block that holds the transactions, not the apphash itself.
func (app *StoreApp) Info(req abci.RequestInfo) abci.ResponseInfo {
	hash := app.Hash()

	app.logger.Info("Info synced",
		"height", app.CommittedHeight(),
		"hash", fmt.Sprintf("%X", hash))

	return abci.ResponseInfo{
		Data:             app.Name,
		LastBlockHeight:  app.CommittedHeight(),
		LastBlockAppHash: hash,
	}
}

// SetOption - ABCI
func (app *StoreApp) SetOption(res abci.RequestSetOption) abci.ResponseSetOption {
	return abci.ResponseSetOption{Log: "Not Implemented"}
}

// Query - ABCI
func (app *StoreApp) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	if len(reqQuery.Data) == 0 {
		resQuery.Log = "Query cannot be zero length"
		resQuery.Code = errors.CodeTypeEncodingErr
		return
	}

	// set the query response height to current
	tree := app.state.Committed()

	height := reqQuery.Height
	if height == 0 {
		// TODO: once the rpc actually passes in non-zero
		// heights we can use to query right after a tx
		// we must retrun most recent, even if apphash
		// is not yet in the blockchain

		withProof := app.CommittedHeight() - 1
		if tree.Tree.VersionExists(uint64(withProof)) {
			height = withProof
		} else {
			height = app.CommittedHeight()
		}
	}
	resQuery.Height = height

	switch reqQuery.Path {
	case "/store", "/key": // Get by key
		key := reqQuery.Data // Data holds the key bytes
		resQuery.Key = key
		if reqQuery.Prove {
			value, proof, err := tree.GetVersionedWithProof(key, height)
			if err != nil {
				resQuery.Log = err.Error()
				break
			}
			resQuery.Value = value
			resQuery.Proof = proof.Bytes()
		} else {
			value := tree.Get(key)
			resQuery.Value = value
		}
	case "/slot":
		slotId := string(reqQuery.Data)
		slot := stake.GetSlot(slotId)
		if slot != nil && slot.State == "Y" {
			resQuery.Value = wire.BinaryBytes(*slot)
		} else {
			resQuery.Value = []byte{}
		}
	case "/slots":
		slots := stake.GetSlots()
		var filtered []*stake.Slot
		for _, slot := range slots {
			if slot.State == "Y" {
				filtered = append(filtered, slot)
			}
		}
		b := wire.BinaryBytes(filtered)
		resQuery.Value = b
	case "/validators":
		candidates := stake.GetCandidates()
		b := wire.BinaryBytes(candidates)
		resQuery.Value = b
	case "/validator":
		address := common.HexToAddress(string(reqQuery.Data))
		candidate := stake.GetCandidateByAddress(address)
		if candidate != nil {
			b := wire.BinaryBytes(*candidate)
			resQuery.Value = b
		} else {
			resQuery.Value = []byte{}
		}
	case "/delegator":
		address := common.HexToAddress(string(reqQuery.Data))
		slotDelegates := stake.GetSlotDelegatesByAddress(address)
		b := wire.BinaryBytes(slotDelegates)
		resQuery.Value = b
	default:
		resQuery.Code = errors.CodeTypeUnknownRequest
		resQuery.Log = cmn.Fmt("Unexpected Query path: %v", reqQuery.Path)
	}

	return
}

// Commit implements abci.Application
func (app *StoreApp) Commit() (res abci.ResponseCommit) {
	app.height++

	hash, err := app.state.Commit(app.height)

	if err != nil {
		// die if we can't commit, not to recover
		panic(err)
	}
	app.logger.Debug("Commit synced",
		"height", app.height,
		"hash", fmt.Sprintf("%X", hash),
	)

	if app.state.Size() == 0 {
		return abci.ResponseCommit{Log: "Empty hash for empty tree"}
	}

	return abci.ResponseCommit{Data: hash}
}

// InitChain - ABCI
func (app *StoreApp) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {
	// create and save the empty candidate
	//for _, validator := range req.Validators {
	//	pk, _ := stake.GetPubKey(string(validator.PubKey))
	//	candidate := stake.NewCandidate(pk, d.sender, uint64(validator.Power), uint64(validator.Power))
	//	stake.SaveCandidate(candidate)
	//}

	return
}

// BeginBlock - ABCI
func (app *StoreApp) BeginBlock(_ abci.RequestBeginBlock) (res abci.ResponseBeginBlock) { return }

// EndBlock - ABCI
// Returns a list of all validator changes made in this block
func (app *StoreApp) EndBlock(_ abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	// TODO: cleanup in case a validator exists multiple times in the list
	res.ValidatorUpdates = app.pending
	app.pending = nil
	return
}

// AddValChange is meant to be called by apps on DeliverTx
// results, this is added to the cache for the endblock
// changeset
func (app *StoreApp) AddValChange(diffs []*abci.Validator) {
	for _, d := range diffs {
		idx := pubKeyIndex(d, app.pending)
		if idx >= 0 {
			app.pending[idx] = d
		} else {
			app.pending = append(app.pending, d)
		}
	}
}

// return index of list with validator of same PubKey, or -1 if no match
func pubKeyIndex(val *abci.Validator, list []*abci.Validator) int {
	for i, v := range list {
		if bytes.Equal(val.PubKey, v.PubKey) {
			return i
		}
	}
	return -1
}

func loadState(dbName string, cacheSize int, historySize int64) (*sm.State, error) {
	// memory backed case, just for testing
	if dbName == "" {
		tree := iavl.NewVersionedTree(0, dbm.NewMemDB())
		return sm.NewState(tree, historySize), nil
	}

	// Expand the path fully
	dbPath, err := filepath.Abs(dbName)
	if err != nil {
		return nil, errors.ErrInternal("Invalid Database Name")
	}

	// Some external calls accidently add a ".db", which is now removed
	dbPath = strings.TrimSuffix(dbPath, path.Ext(dbPath))

	// Split the database name into it's components (dir, name)
	dir := path.Dir(dbPath)
	name := path.Base(dbPath)

	// Open database called "dir/name.db", if it doesn't exist it will be created
	db := dbm.NewDB(name, dbm.LevelDBBackendStr, dir)
	tree := iavl.NewVersionedTree(cacheSize, db)
	if err = tree.Load(); err != nil {
		return nil, errors.ErrInternal("Loading tree: " + err.Error())
	}

	return sm.NewState(tree, historySize), nil
}

func initStakeDb() error {
	rootDir := viper.GetString(cli.HomeFlag)
	stakeDbPath := path.Join(rootDir, "data", "stake.db")
	_, err := os.OpenFile(stakeDbPath, os.O_RDONLY, 0444)
	if err != nil {
		db, err := sql.Open("sqlite3", stakeDbPath)
		if err != nil {
			return errors.ErrInternal("Initializing stake database: " + err.Error())
		}
		defer db.Close()

		sqlStmt := `
		create table candidates(address text not null primary key, pub_key text not null, shares integer default 0, voting_power integer default 0, state text not null default 'Y', created_at text not null, updated_at text);
		create table delegators(address text not null primary key, created_at text not null);
		create table slots(id text not null primary key, validator_address text not null, total_amount integer not null default 0, available_amount integer not null default 0, proposed_roi integer not null default 0, state text not null default 'Y', created_at text not null, updated_at text);
		create table delegate_history(id integer not null primary key autoincrement, delegator_address text not null, slot_id text not null, amount integer not null default 0, op_code text not null default '', created_at text not null);
		create table slot_delegates (delegator_address text not null, slot_id text not null, amount integer not null default 0, created_at text not null, updated_at text);
		create index idx_slots_validator_address on slots(validator_address);
		create index idx_candidates_pub_key on candidates(pub_key);
		create index idx_slot_delegates_delegator_address on slot_delegates(delegator_address);
		create index idx_slot_delegates_slot_id on slot_delegates(slot_id);
		`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			//os.Remove(stakeDbPath)
			return errors.ErrInternal("Initializing stake database: " + err.Error())
		}
	}

	return nil
}
