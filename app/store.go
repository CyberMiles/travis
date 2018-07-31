package app

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"path"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ripemd160"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	cmn "github.com/tendermint/tendermint/libs/common"
	tDB "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/sdk/errors"
	sm "github.com/CyberMiles/travis/sdk/state"
	"github.com/CyberMiles/travis/sdk/dbm"
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
	pending []abci.Validator

	// height is last committed block, DeliverTx is the next one
	height int64

	TotalUsedGasFee *big.Int

	logger log.Logger

	sqliter *dbm.Sqliter

	BlockEnd            bool
}

// NewStoreApp creates a data store to handle queries
func NewStoreApp(appName, dbName string, cacheSize int, logger log.Logger) (*StoreApp, error) {
	state, err := loadState(dbName, cacheSize, DefaultHistorySize)
	if err != nil {
		return nil, err
	}

	rootDir := viper.GetString(cli.HomeFlag)
	dbPath := path.Join(rootDir, "data", "travis.db")

	sqliter := dbm.NewSqliter(dbPath)

	app := &StoreApp{
		Name:            appName,
		state:           state,
		height:          state.LatestHeight(),
		info:            sm.NewChainState(),
		TotalUsedGasFee: big.NewInt(0),
		logger:          logger.With("module", "app"),
		sqliter:		 sqliter,
	}
	return app, nil
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

// Check returns the working state for Chec
// kTx
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

// WorkingHeight gets the current block we are writing
func (app *StoreApp) GetSqliter() *dbm.Sqliter {
	return app.sqliter
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
		if tree.Tree.VersionExists(withProof) {
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
		value := app.state.Check().Get(key)
		fmt.Printf("Check Value: %s\n", hex.EncodeToString(value))
		resQuery.Value = value

		if reqQuery.Prove {
			value, proof, err := tree.GetVersionedWithProof(key, height)
			if err != nil {
				resQuery.Log = err.Error()
				break
			}
			resQuery.Value = value
			resQuery.Proof = proof.ComputeRootHash()
		} else {
			value := tree.Get(key)
			resQuery.Value = value
		}
	case "/validators":
		candidates := stake.GetCandidates()
		b, _ := json.Marshal(candidates)
		resQuery.Value = b
	case "/validator":
		address := common.HexToAddress(string(reqQuery.Data))
		candidate := stake.GetCandidateByAddress(address)
		if candidate != nil {
			b, _ := json.Marshal(candidate)
			resQuery.Value = b
		} else {
			resQuery.Value = []byte{}
		}
	case "/delegator":
		address := common.HexToAddress(string(reqQuery.Data))
		delegations := stake.GetDelegationsByDelegator(address)
		b, _ := json.Marshal(delegations)
		resQuery.Value = b
	case "/governance/proposals":
		proposals := governance.GetProposals()
		b, _ := json.Marshal(proposals)
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
		return abci.ResponseCommit{}
	}

	return abci.ResponseCommit{Data: hash}
}

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
func (app *StoreApp) AddValChange(diffs []abci.Validator) {
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
func pubKeyIndex(val abci.Validator, list []abci.Validator) int {
	for i, v := range list {
		if bytes.Equal(val.PubKey.Data, v.PubKey.Data) {
			return i
		}
	}
	return -1
}

func loadState(dbName string, cacheSize int, historySize int64) (*sm.State, error) {
	// memory backed case, just for testing
	if dbName == "" {
		tree := iavl.NewVersionedTree(tDB.NewMemDB(), 0)
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
	db := tDB.NewDB(name, tDB.LevelDBBackend, dir)
	tree := iavl.NewVersionedTree(db, cacheSize)
	if _, err = tree.Load(); err != nil {
		return nil, errors.ErrInternal("Loading tree: " + err.Error())
	}

	return sm.NewState(tree, historySize), nil
}

func getDb() *sql.DB {
	rootDir := viper.GetString(cli.HomeFlag)
	dbPath := path.Join(rootDir, "data", "travis.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	return db
}

func (app *StoreApp) GetDbHash() []byte {
	db, _ := app.sqliter.GetDB()
	defer app.sqliter.CloseDB(db)

	tables := []string{"candidates", "delegations", "governance_proposal", "governance_vote", "unstake_requests"}
	hashes := make([]byte, len(tables))
	for _, table := range tables {
		hashes = append(hashes, getTableHash(db, table)...)
	}
	return hashing(hashes)
}

func getTableHash(db *sql.DB, table string) []byte {
	stmt, err := db.Prepare("select hash from " + table + " where 1=1 order by hash")
	if err != nil {
		fmt.Println(err)
	}

	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		panic(err)
	}
	var hash string
	hashes := make([]byte, 80)
	for rows.Next() {
		err = rows.Scan(&hash)
		if err != nil {
			panic(err)
		}
		hashes = append(hashes, []byte(hash)...)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	//fmt.Printf("Table %s, hash: %s\n", table, common.Bytes2Hex(hashing(hashes)))
	return hashing(hashes)
}

func hashing(h []byte) []byte {
	hasher := ripemd160.New()
	buf := new(bytes.Buffer)
	buf.Write(h)
	hasher.Write(buf.Bytes())
	return hasher.Sum(nil)
}
