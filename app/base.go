package app

import (
	goerr "errors"
	"fmt"
	"math/big"

	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/errors"
	"github.com/CyberMiles/travis/sdk/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	abci "github.com/tendermint/abci/types"

	"github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/modules/stake"
	ttypes "github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"golang.org/x/crypto/ripemd160"
	"bytes"
)

// BaseApp - The ABCI application
type BaseApp struct {
	*StoreApp
	EthApp              *EthermintApplication
	checkedTx           map[common.Hash]*types.Transaction
	ethereum            *eth.Ethereum
	AbsentValidators    *stake.AbsentValidators
	ByzantineValidators []abci.Evidence
	LastValidators      []stake.Validator
}

const (
	BLOCK_AWARD_STR = "10000000000000000000000"
)

var (
	blockAward, _                  = big.NewInt(0).SetString(BLOCK_AWARD_STR, 10)
	_             abci.Application = &BaseApp{}
)

// NewBaseApp extends a StoreApp with a handler and a ticker,
// which it binds to the proper abci calls
func NewBaseApp(store *StoreApp, ethApp *EthermintApplication, ethereum *eth.Ethereum) (*BaseApp, error) {
	// init pending proposals
	pendingProposals := governance.GetPendingProposals()
	if len(pendingProposals) > 0 {
		proposals := make(map[string]uint64)
		for _, pp := range pendingProposals {
			proposals[pp.Id] = pp.ExpireBlockHeight
		}
		utils.PendingProposal.BatchAdd(proposals)
	}

	app := &BaseApp{
		StoreApp:         store,
		EthApp:           ethApp,
		checkedTx:        make(map[common.Hash]*types.Transaction),
		ethereum:         ethereum,
		AbsentValidators: stake.NewAbsentValidators(),
	}

	return app, nil
}

// InitChain - ABCI
func (app *StoreApp) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {
	return
}

// Info implements abci.Application. It returns the height and hash,
// as well as the abci name and version.
//
// The height is the block that holds the transactions, not the apphash itself.
func (app *BaseApp) Info(req abci.RequestInfo) abci.ResponseInfo {
	ethInfoRes := app.EthApp.Info(req)

	if big.NewInt(ethInfoRes.LastBlockHeight).Cmp(bigZero) == 0 {
		return ethInfoRes
	}

	travisInfoRes := app.StoreApp.Info(req)

	travisInfoRes.LastBlockAppHash = finalAppHash(ethInfoRes.LastBlockAppHash, travisInfoRes.LastBlockAppHash, travisInfoRes.LastBlockHeight, nil)
	return travisInfoRes
}

// DeliverTx - ABCI
func (app *BaseApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {
	tx, err := decodeTx(txBytes)
	if err != nil {
		app.logger.Error("DeliverTx: Received invalid transaction", "err", err)
		return errors.DeliverResult(err)
	}

	if isEthTx(tx) {
		if checkedTx, ok := app.checkedTx[tx.Hash()]; ok {
			tx = checkedTx
		} else {
			// force cache from of tx
			// TODO: Get chainID from config
			if _, err := types.Sender(types.NewEIP155Signer(big.NewInt(111)), tx); err != nil {
				app.logger.Debug("DeliverTx: Received invalid transaction", "tx", tx, "err", err)
				return errors.DeliverResult(err)
			}
		}
		resp := app.EthApp.DeliverTx(tx)
		app.logger.Debug("EthApp DeliverTx response", "resp", resp)
		return resp
	}

	app.logger.Info("DeliverTx: Received valid transaction", "tx", tx)

	ctx := ttypes.NewContext(app.GetChainID(), app.WorkingHeight(), app.EthApp.DeliverTxState())
	return app.deliverHandler(ctx, app.Append(), tx)
}

// CheckTx - ABCI
func (app *BaseApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	tx, err := decodeTx(txBytes)
	if err != nil {
		app.logger.Error("CheckTx: Received invalid transaction", "err", err)
		return errors.CheckResult(err)
	}

	if isEthTx(tx) {
		resp := app.EthApp.CheckTx(tx)
		app.logger.Debug("EthApp CheckTx response", "resp", resp)
		if resp.IsErr() {
			return errors.CheckResult(goerr.New(resp.String()))
		}
		app.checkedTx[tx.Hash()] = tx
		return sdk.NewCheck(0, "").ToABCI()
	}

	app.logger.Info("CheckTx: Received valid transaction", "tx", tx)

	ctx := ttypes.NewContext(app.GetChainID(), app.WorkingHeight(), app.EthApp.checkTxState)
	return app.checkHandler(ctx, app.Check(), tx)
}

// BeginBlock - ABCI
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	app.EthApp.BeginBlock(req)

	// handle the absent validators
	for _, i := range req.AbsentValidators {
		app.logger.Debug("BeginBlock", "index", i)
		if i >= int32(len(app.LastValidators)) {
			continue
		}

		v := app.LastValidators[i]
		app.AbsentValidators.Add(v.PubKey, app.WorkingHeight())
	}
	app.AbsentValidators.Clear(app.WorkingHeight())

	app.logger.Info("BeginBlock", "absent_validators", app.AbsentValidators)
	app.ByzantineValidators = req.ByzantineValidators

	return abci.ResponseBeginBlock{}
}

// EndBlock - ABCI - triggers Tick actions
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	app.EthApp.EndBlock(req)

	cs := stake.GetCandidates()
	cs.Sort()
	app.LastValidators = cs.Validators()
	validators := cs.Validators()

	// block award
	stake.NewAwardDistributor(app.WorkingHeight(), validators, utils.BlockGasFee, app.AbsentValidators, app.logger).DistributeAll()

	// punish Byzantine validators
	if len(app.ByzantineValidators) > 0 {
		for _, bv := range app.ByzantineValidators {
			pk, err := ttypes.GetPubKey(string(bv.PubKey))
			if err != nil {
				continue
			}

			stake.PunishByzantineValidator(pk)
			app.ByzantineValidators = app.ByzantineValidators[:0]
		}
	}

	// punish the absent validators
	for k, v := range app.AbsentValidators.Validators {
		stake.PunishAbsentValidator(k, v)
	}

	// execute tick if present
	diff, err := tick(app.Append())
	if err != nil {
		panic(err)
	}
	app.AddValChange(diff)

	return app.StoreApp.EndBlock(req)
}

func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	app.checkedTx = make(map[common.Hash]*types.Transaction)
	ethAppCommit := app.EthApp.Commit()

	workingHeight := app.WorkingHeight()

	res = app.StoreApp.Commit()
	res.Data = finalAppHash(ethAppCommit.Data, res.Data, workingHeight, nil)
	return
}

func (app *BaseApp) InitState(module, key string, value interface{}) error {
	state := app.Append()
	logger := app.Logger().With("module", module, "key", key)

	if module == sdk.ModuleNameBase {
		if key == sdk.ChainKey {
			app.info.SetChainID(state, value.(string))
			return nil
		}
		logger.Error("Invalid genesis option")
		return fmt.Errorf("unknown base option: %s", key)
	}

	err := stake.InitState(key, value, state)
	if err != nil {
		logger.Error("Invalid genesis option", "err", err)
	}
	return err
}

func isEthTx(tx *types.Transaction) bool {
	zero := big.NewInt(0)
	return tx.Data() == nil ||
		tx.GasPrice().Cmp(zero) != 0 ||
		tx.Gas().Cmp(zero) != 0 ||
		tx.Value().Cmp(zero) != 0 ||
		tx.To() != nil
}

// Tick - Called every block even if no transaction, process all queues,
// validator rewards, and calculate the validator set difference
func tick(store state.SimpleDB) (change []abci.Validator, err error) {
	change, err = stake.UpdateValidatorSet(store)
	return
}

func finalAppHash(ethCommitHash []byte, travisCommitHash []byte, workingHeight int64, store *state.SimpleDB) []byte{

	hasher := ripemd160.New()
	buf := new(bytes.Buffer)
	buf.Write(ethCommitHash)
	buf.Write(travisCommitHash)
	hasher.Write(buf.Bytes())
	hash := hasher.Sum(nil)

	if store != nil {
		// TODO: save to DB
	}
	return hash
}
