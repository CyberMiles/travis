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
)

// BaseApp - The ABCI application
type BaseApp struct {
	*StoreApp
	EthApp              *EthermintApplication
	checkedTx           map[common.Hash]*types.Transaction
	ethereum            *eth.Ethereum
	AbsentValidators    []int32
	ByzantineValidators []abci.Evidence
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
		StoreApp:  store,
		EthApp:    ethApp,
		checkedTx: make(map[common.Hash]*types.Transaction),
		ethereum:  ethereum,
	}

	return app, nil
}

// InitChain - ABCI
func (app *StoreApp) InitChain(req abci.RequestInitChain) (res abci.ResponseInitChain) {
	return
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
		app.logger.Debug("EthApp DeliverTx response: %v\n", resp)
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
		app.logger.Debug("EthApp CheckTx response: %v\n", resp)
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
	app.AbsentValidators = req.AbsentValidators
	app.ByzantineValidators = req.ByzantineValidators

	return abci.ResponseBeginBlock{}
}

// EndBlock - ABCI - triggers Tick actions
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	app.EthApp.EndBlock(req)

	// execute tick if present
	diff, err := tick(app.Append())
	if err != nil {
		panic(err)
	}
	app.AddValChange(diff)

	// block award
	cs := stake.GetCandidates()
	cs.Sort()
	validators := cs.Validators()
	//for _, i := range app.AbsentValidators {
	//	validators.Remove(i)
	//}
	stake.NewAwardCalculator(app.WorkingHeight(), validators, utils.BlockGasFee).AwardAll()

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

	// todo punish those validators who has been absent for up to 3 hours

	return app.StoreApp.EndBlock(req)
}

func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	app.checkedTx = make(map[common.Hash]*types.Transaction)
	app.EthApp.Commit()
	res = app.StoreApp.Commit()
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
