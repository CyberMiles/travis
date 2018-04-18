package app

import (
	"bytes"
	goerr "errors"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	abci "github.com/tendermint/abci/types"

	"github.com/CyberMiles/travis/modules"
	"github.com/CyberMiles/travis/modules/stake"
	ttypes "github.com/CyberMiles/travis/types"
	ethapp "github.com/CyberMiles/travis/vm/app"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/CyberMiles/travis/utils"
)

// BaseApp - The ABCI application
type BaseApp struct {
	*StoreApp
	handler                modules.Handler
	clock                  sdk.Ticker
	EthApp                 *ethapp.EthermintApplication
	checkedTx              map[common.Hash]*types.Transaction
	ethereum               *eth.Ethereum
	AbsentValidators       []int32
	ByzantineValidators    []*abci.Evidence
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
func NewBaseApp(store *StoreApp, ethApp *ethapp.EthermintApplication, clock sdk.Ticker, ethereum *eth.Ethereum) (*BaseApp, error) {
	app := &BaseApp{
		StoreApp:               store,
		handler:                modules.Handler{},
		clock:                  clock,
		EthApp:                 ethApp,
		checkedTx:              make(map[common.Hash]*types.Transaction),
		ethereum:               ethereum,
	}

	return app, nil
}

// DeliverTx - ABCI
func (app *BaseApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {
	ok, tx := isTravisTx(txBytes)
	if !ok {
		ok, tx, err := isEthTx(txBytes)
		if !ok {
			app.logger.Debug("DeliverTx: Received invalid transaction", "tx", tx, "err", err)
			return errors.DeliverResult(err)
		}

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
		app.logger.Debug("ethermint DeliverTx response: %v\n", resp)
		return resp
	}

	app.logger.Info("DeliverTx: Received valid transaction", "tx", tx)

	ctx := ttypes.NewContext(app.GetChainID(), app.WorkingHeight(), app.ethereum)
	res, err := app.handler.DeliverTx(ctx, app.Append(), tx)
	if err != nil {
		return errors.DeliverResult(err)
	}
	app.AddValChange(res.Diff)
	return res.ToABCI()
}

// CheckTx - ABCI
func (app *BaseApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	ok, tx := isTravisTx(txBytes)
	if !ok {
		ok, tx, err := isEthTx(txBytes)
		if !ok {
			app.logger.Debug("CheckTx: Received invalid transaction", "tx", tx, "err", err)
			return errors.CheckResult(err)
		}

		resp := app.EthApp.CheckTx(tx)
		app.logger.Debug("ethermint CheckTx response: %v\n", resp)
		if resp.IsErr() {
			return errors.CheckResult(goerr.New(resp.Error()))
		}
		app.checkedTx[tx.Hash()] = tx
		return sdk.NewCheck(0, "").ToABCI()
	}

	app.logger.Info("CheckTx: Receivted valid transaction", "tx", tx)

	ctx := ttypes.NewContext(app.GetChainID(), app.WorkingHeight(), app.ethereum)
	res, err := app.handler.CheckTx(ctx, app.Check(), tx)
	if err != nil {
		return errors.CheckResult(err)
	}
	return res.ToABCI()
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
	if app.clock != nil {
		ctx := stack.NewContext(
			app.GetChainID(),
			app.WorkingHeight(),
			app.Logger().With("call", "tick"),
		)

		diff, err := app.clock.Tick(ctx, app.Append())
		if err != nil {
			panic(err)
		}
		app.AddValChange(diff)
	}

	// block award
	validators := stake.GetCandidates().Validators()
	for _, i := range app.AbsentValidators {
		validators.Remove(i)
	}
	stake.NewAwardCalculator(app.WorkingHeight(), validators, utils.BlockGasFee).AwardAll()

	// todo punish Byzantine validators

	return app.StoreApp.EndBlock(req)
}

func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	app.checkedTx = make(map[common.Hash]*types.Transaction)
	app.EthApp.Commit()
	//var hash = resp.Data
	//app.logger.Debug("ethermint Commit response, %v, hash: %v\n", resp, hash.String())

	res = app.StoreApp.Commit()
	return
}

func (app *BaseApp) InitState(module, key, value string) error {
	state := app.Append()
	logger := app.Logger().With("module", module, "key", key)

	if module == sdk.ModuleNameBase {
		if key == sdk.ChainKey {
			app.info.SetChainID(state, value)
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

// rlp decode an ethereum transaction
func decodeEthTx(txBytes []byte) (*types.Transaction, error) {
	tx := new(types.Transaction)
	rlpStream := rlp.NewStream(bytes.NewBuffer(txBytes), 0)
	if err := tx.DecodeRLP(rlpStream); err != nil {
		return nil, err
	}
	return tx, nil
}

func isTravisTx(txBytes []byte) (bool, sdk.Tx) {
	tx, err := sdk.LoadTx(txBytes)
	if err != nil {
		return false, sdk.Tx{}
	}

	return true, tx
}

func isEthTx(txBytes []byte) (bool, *types.Transaction, error) {
	// try to decode with ethereum
	tx, err := decodeEthTx(txBytes)
	if err != nil {
		return false, nil, err
	}
	return true, tx, nil
}
