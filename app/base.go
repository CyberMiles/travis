package app

import (
	"fmt"
	"bytes"

	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk"
	abcicli "github.com/tendermint/abci/client"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/tendermint/go-wire/data"
	//"github.com/tendermint/go-wire"
	"github.com/CyberMiles/travis/modules/stake"
	//auth "github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/CyberMiles/travis/utils"
)

// BaseApp - The ABCI application
type BaseApp struct {
	*StoreApp
	handler sdk.Handler
	clock   sdk.Ticker
}

const (
	ETHERMINT_ADDR = "localhost:8848"
	BLOCK_AWARD = 10000000000000000000000
)

var (
	_ abci.Application = &BaseApp{}
	client, err = abcicli.NewClient(ETHERMINT_ADDR, "socket", true)
	handler = stake.NewHandler()
)

// NewBaseApp extends a StoreApp with a handler and a ticker,
// which it binds to the proper abci calls
func NewBaseApp(store *StoreApp, handler sdk.Handler, clock sdk.Ticker) *BaseApp {
	client.Start()

	return &BaseApp{
		StoreApp: store,
		handler:  handler,
		clock:    clock,
	}
}

// DeliverTx - ABCI - dispatches to the handler
func (app *BaseApp) DeliverTx(txBytes []byte) abci.ResponseDeliverTx {
	fmt.Println("DeliverTx")

	tx, err := sdk.LoadTx(txBytes)
	if err != nil {
		// try to decode with ethereum
		tx, err := decodeTx(txBytes)
		if err != nil {
			app.logger.Debug("DeliverTx: Received invalid transaction", "tx", tx, "err", err)

			return errors.DeliverResult(err)
		}

		resp, err := client.DeliverTxSync(txBytes)
		fmt.Printf("ethermint DeliverTx response: %v\n", resp)

		return abci.ResponseDeliverTx{Code: 0}
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
	fmt.Println("CheckTx")

	tx, err := sdk.LoadTx(txBytes)
	if err != nil {
		// try to decode with ethereum
		tx, err := decodeTx(txBytes)
		if err != nil {
			app.logger.Debug("CheckTx: Received invalid transaction", "tx", tx, "err", err)

			return errors.CheckResult(err)
		}

		resp, err := client.CheckTxSync(txBytes)
		fmt.Printf("ethermint CheckTx response: %v\n", resp)

		if err != nil {
			return errors.CheckResult(err)
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
	fmt.Println("BeginBlock")

	resp, _ := client.BeginBlockSync(beginBlock)
	fmt.Printf("ethermint BeginBlock response: %v\n", resp)

	evidences := beginBlock.ByzantineValidators
	for _, evidence := range evidences {
		utils.ValidatorPubKeys = append(utils.ValidatorPubKeys, evidence.GetPubKey())
	}

	return abci.ResponseBeginBlock{}
}

// EndBlock - ABCI - triggers Tick actions
func (app *BaseApp) EndBlock(endBlock abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	fmt.Println("EndBlock")

	resp, _ := client.EndBlockSync(endBlock)

	fmt.Printf("ethermint EndBlock response: %v\n", resp)

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
		utils.StateChangeQueue = append(utils.StateChangeQueue, utils.StateChangeObject{
			From: stake.DefaultHoldAccount.Address, To: []byte(k), Amount: int64(BLOCK_AWARD * v)})
	}

	// todo send StateChangeQueue to VM

	return app.StoreApp.EndBlock(endBlock)
}

func (app *BaseApp) Commit() (res abci.ResponseCommit) {
	fmt.Println("Commit")

	resp, err := client.CommitSync()

	if err != nil {
		panic(err)
	}

	var hash = resp.Data
	fmt.Printf("ethermint Commit response, %v, hash: %v\n", resp, hash.String())

	return abci.ResponseCommit{Data: resp.Data}
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

func (app *BaseApp) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	fmt.Println("Query")

	var d = data.Bytes(reqQuery.Data)
	fmt.Println(d)
	fmt.Println(d.MarshalJSON())
	reqQuery.Data, _ = d.MarshalJSON()

	resp, err := client.QuerySync(reqQuery)

	if err != nil {
		panic(err)
	}

	return *resp
}

// rlp decode an etherum transaction
func decodeTx(txBytes []byte) (*types.Transaction, error) {
	tx := new(types.Transaction)
	rlpStream := rlp.NewStream(bytes.NewBuffer(txBytes), 0)
	if err := tx.DecodeRLP(rlpStream); err != nil {
		return nil, err
	}
	return tx, nil
}


