package app

import (
	goerr "errors"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	abci "github.com/tendermint/abci/types"

	"github.com/CyberMiles/travis/modules"
	"github.com/CyberMiles/travis/modules/stake"
	ttypes "github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
)

// BaseApp - The ABCI application
type BaseApp struct {
	*StoreApp
	handler                modules.Handler
	clock                  sdk.Ticker
	EthApp                 *EthermintApplication
	checkedTx              map[common.Hash]*types.Transaction
	AbsentValidatorPubKeys [][]byte
	ethereum               *eth.Ethereum
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
func NewBaseApp(store *StoreApp, ethApp *EthermintApplication, clock sdk.Ticker, ethereum *eth.Ethereum) (*BaseApp, error) {
	app := &BaseApp{
		StoreApp:               store,
		handler:                modules.Handler{},
		clock:                  clock,
		EthApp:                 ethApp,
		checkedTx:              make(map[common.Hash]*types.Transaction),
		AbsentValidatorPubKeys: [][]byte{},
		ethereum:               ethereum,
	}

	return app, nil
}

// DeliverTx - ABCI
func (app *BaseApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {
	app.logger.Debug("DeliverTx")

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

	// travis tx
	app.logger.Info("DeliverTx: Received valid transaction", "tx", tx)

	ctx := ttypes.NewContext(app.GetChainID(), app.WorkingHeight(), app.ethereum)
	return app.deliverHandler(ctx, app.Append(), tx)
}

// CheckTx - ABCI
func (app *BaseApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	app.logger.Debug("CheckTx")

	tx, err := decodeTx(txBytes)
	if err != nil {
		app.logger.Error("CheckTx: Received invalid transaction", "err", err)
		return errors.CheckResult(err)
	}

	if isEthTx(tx) {
		resp := app.EthApp.CheckTx(tx)
		app.logger.Debug("EthApp CheckTx response: %v\n", resp)
		if resp.IsErr() {
			return errors.CheckResult(goerr.New(resp.Error()))
		}
		app.checkedTx[tx.Hash()] = tx
		return sdk.NewCheck(0, "").ToABCI()
	}

	// travis tx
	app.logger.Info("CheckTx: Received valid transaction", "tx", tx)

	ctx := ttypes.NewContext(app.GetChainID(), app.WorkingHeight(), app.ethereum)
	return app.checkHandler(ctx, app.Check(), tx)
}

// BeginBlock - ABCI
func (app *BaseApp) BeginBlock(beginBlock abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	app.logger.Debug("BeginBlock")

	resp := app.EthApp.BeginBlock(beginBlock)
	app.logger.Debug("ethermint BeginBlock response: %v\n", resp)

	app.AbsentValidatorPubKeys = [][]byte{}
	evidences := beginBlock.ByzantineValidators
	for _, evidence := range evidences {
		app.AbsentValidatorPubKeys = append(app.AbsentValidatorPubKeys, evidence.GetPubKey())
	}
	return abci.ResponseBeginBlock{}
}

// EndBlock - ABCI - triggers Tick actions
func (app *BaseApp) EndBlock(endBlock abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	app.logger.Debug("EndBlock")

	//resp, _ := app.client.EndBlockSync(endBlock)
	resp := app.EthApp.EndBlock(endBlock)
	app.logger.Debug("ethermint EndBlock response: %v\n", resp)

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
	// fixme exclude absent validators
	var validatorPubKeys [][]byte
	ratioMap := stake.CalValidatorsStakeRatio(app.Append(), validatorPubKeys)
	for k, v := range ratioMap {
		awardAmount := big.NewInt(0)
		intv := int64(1000 * v)
		awardAmount.Mul(blockAward, big.NewInt(intv))
		awardAmount.Div(awardAmount, big.NewInt(1000))
		utils.StateChangeQueue = append(utils.StateChangeQueue, utils.StateChangeObject{
			From: stake.DefaultHoldAccount, To: common.HexToAddress(k), Amount: awardAmount})
	}

	// todo send StateChangeQueue to VM

	return app.StoreApp.EndBlock(endBlock)
}

func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	app.logger.Debug("Commit")

	app.checkedTx = make(map[common.Hash]*types.Transaction)
	app.EthApp.Commit()
	//var hash = resp.Data
	//app.logger.Debug("ethermint Commit response, %v, hash: %v\n", resp, hash.String())

	resp := app.StoreApp.Commit()
	return resp
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

func isEthTx(tx *types.Transaction) bool {
	zero := big.NewInt(0)
	return tx.Data() == nil ||
		tx.GasPrice().Cmp(zero) != 0 ||
		tx.Gas().Cmp(zero) != 0 ||
		tx.Value().Cmp(zero) != 0 ||
		tx.To() != nil
}
