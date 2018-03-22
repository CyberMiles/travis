package app

import (
	"bytes"
	goerr "errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	_ "github.com/tendermint/abci/client"
	abci "github.com/tendermint/abci/types"
	//"github.com/tendermint/go-wire/data"
	//auth "github.com/cosmos/cosmos-sdk/modules/auth"

	"github.com/CyberMiles/travis/modules/stake"
	ethapp "github.com/CyberMiles/travis/modules/vm/app"
	"github.com/CyberMiles/travis/utils"
)

// BaseApp - The ABCI application
type BaseApp struct {
	*StoreApp
	handler sdk.Handler
	clock   sdk.Ticker
	//client abcicli.Client
	EthApp *ethapp.EthermintApplication
}

const (
	ETHERMINT_ADDR  = "localhost:8848"
	BLOCK_AWARD_STR = "10000000000000000000000"
)

var (
	blockAward, _ = big.NewInt(0).SetString(BLOCK_AWARD_STR, 10)

	_ abci.Application = &BaseApp{}
	//client, err = abcicli.NewClient(ETHERMINT_ADDR, "socket", true)
	handler = stake.NewHandler()
)

// NewBaseApp extends a StoreApp with a handler and a ticker,
// which it binds to the proper abci calls
func NewBaseApp(store *StoreApp, ethApp *ethapp.EthermintApplication, handler sdk.Handler, clock sdk.Ticker) (*BaseApp, error) {
	app := &BaseApp{
		StoreApp: store,
		handler:  handler,
		clock:    clock,
		EthApp:   ethApp,
	}
	//client, err := abcicli.NewClient(ETHERMINT_ADDR, "socket", true)
	//if err != nil {
	//	return nil, err
	//}
	//if err := client.Start(); err != nil {
	//	return nil, err
	//}
	//
	//app.client = client

	return app, nil
}

// DeliverTx - ABCI - dispatches to the handler
func (app *BaseApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {
	app.logger.Debug("DeliverTx")

	tx, err := sdk.LoadTx(txBytes)
	if err != nil {
		// try to decode with ethereum
		tx, err := decodeTx(txBytes)
		if err != nil {
			app.logger.Debug("DeliverTx: Received invalid transaction", "tx", tx, "err", err)

			return errors.DeliverResult(err)
		}

		// force cache from of tx
		// TODO: Get chainID from config
		if _, err := types.Sender(types.NewEIP155Signer(big.NewInt(111)), tx); err != nil {
			app.logger.Debug("DeliverTx: Received invalid transaction", "tx", tx, "err", err)

			return errors.DeliverResult(err)
		}
		//resp, err := app.client.DeliverTxSync(txBytes)
		resp := app.EthApp.DeliverTx(tx)
		app.logger.Debug("ethermint DeliverTx response: %v\n", resp)

		return resp
	}

	app.logger.Info("DeliverTx: Received valid transaction", "tx", tx)

	ctx := stack.NewContext(
		app.GetChainID(),
		app.WorkingHeight(),
		app.Logger().With("call", "delivertx"),
	)

	//// fixme check if it's sendTx
	//switch tx.Unwrap().(type) {
	//case coin.SendTx:
	//	//return h.sendTx(ctx, store, t, cb)
	//	fmt.Println("transfer tx")
	//}

	res, err := app.handler.DeliverTx(ctx, app.Append(), tx)

	if err != nil {
		return errors.DeliverResult(err)
	}
	app.AddValChange(res.Diff)
	return res.ToABCI()
}

// CheckTx - ABCI - dispatches to the handler
func (app *BaseApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	app.logger.Debug("CheckTx")

	tx, err := sdk.LoadTx(txBytes)
	if err != nil {
		// try to decode with ethereum
		tx, err := decodeTx(txBytes)
		if err != nil {
			app.logger.Debug("CheckTx: Received invalid transaction", "tx", tx, "err", err)

			return errors.CheckResult(err)
		}

		//resp, err := app.client.CheckTxSync(txBytes)
		resp := app.EthApp.CheckTx(tx)
		app.logger.Debug("ethermint CheckTx response: %v\n", resp)

		if resp.IsErr() {
			return errors.CheckResult(goerr.New(resp.Error()))
		}

		return sdk.NewCheck(21000, "").ToABCI()
	}

	app.logger.Info("CheckTx: Received valid transaction", "tx", tx)

	ctx := stack.NewContext(
		app.GetChainID(),
		app.WorkingHeight(),
		app.Logger().With("call", "checktx"),
	)

	//ctx2, err := verifySignature(ctx, tx)
	//
	//// fixme check if it's sendTx
	//switch tx.Unwrap().(type) {
	//case coin.SendTx:
	//	//return h.sendTx(ctx, store, t, cb)
	//	fmt.Println("checkTx: transfer")
	//	return sdk.NewCheck(21000, "").ToABCI()
	//}

	res, err := app.handler.CheckTx(ctx, app.Check(), tx)

	if err != nil {
		return errors.CheckResult(err)
	}
	return res.ToABCI()
}

// BeginBlock - ABCI
func (app *BaseApp) BeginBlock(beginBlock abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	app.logger.Debug("BeginBlock")

	//resp, _ := app.client.BeginBlockSync(beginBlock)
	resp := app.EthApp.BeginBlock(beginBlock)
	app.logger.Debug("ethermint BeginBlock response: %v\n", resp)

	evidences := beginBlock.ByzantineValidators
	for _, evidence := range evidences {
		utils.ValidatorPubKeys = append(utils.ValidatorPubKeys, evidence.GetPubKey())
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
	ratioMap := stake.CalValidatorsStakeRatio(app.Append(), utils.ValidatorPubKeys)
	for k, v := range ratioMap {
		awardAmount := big.NewInt(0)
		intv := int64(1000 * v)
		awardAmount.Mul(blockAward, big.NewInt(intv))
		awardAmount.Div(awardAmount, big.NewInt(1000))
		utils.StateChangeQueue = append(utils.StateChangeQueue, utils.StateChangeObject{
			From: stake.DefaultHoldAccount.Address, To: []byte(k), Amount: awardAmount})
	}

	// todo send StateChangeQueue to VM

	return app.StoreApp.EndBlock(endBlock)
}

func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	app.logger.Debug("Commit")

	//resp, err := app.client.CommitSync()
	//if err != nil {
	//	panic(err)
	//}

	resp := app.EthApp.Commit()
	var hash = resp.Data
	app.logger.Debug("ethermint Commit response, %v, hash: %v\n", resp, hash.String())

	app.StoreApp.Commit()
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
		return fmt.Errorf("Unknown base option: %s", key)
	}

	log, err := app.handler.InitState(logger, state, module, key, value)
	if err != nil {
		logger.Error("Invalid genesis option", "err", err)
	} else {
		logger.Info(log)
	}
	return err
}

func (app *BaseApp) Info(req abci.RequestInfo) abci.ResponseInfo {
	//resp, err := app.client.InfoSync(res)
	//app.EthApp.Info(res)
	//if err != nil {
	//	panic(err)
	//}
	//
	//return *resp
	return app.EthApp.Info(req)
}

//func (app *BaseApp) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
//	fmt.Println("Query")
//
//	var d = data.Bytes(reqQuery.Data)
//	fmt.Println(d)
//	fmt.Println(d.MarshalJSON())
//	reqQuery.Data, _ = d.MarshalJSON()
//
//	resp, err := client.QuerySync(reqQuery)
//
//	if err != nil {
//		panic(err)
//	}
//
//	return *resp
//}

// rlp decode an ethereum transaction
func decodeTx(txBytes []byte) (*types.Transaction, error) {
	tx := new(types.Transaction)
	rlpStream := rlp.NewStream(bytes.NewBuffer(txBytes), 0)
	if err := tx.DecodeRLP(rlpStream); err != nil {
		return nil, err
	}
	return tx, nil
}
