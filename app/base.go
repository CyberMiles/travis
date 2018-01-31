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
)

// BaseApp - The ABCI application
type BaseApp struct {
	*StoreApp
	handler sdk.Handler
	clock   sdk.Ticker
}

var _ abci.Application = &BaseApp{}

const ETHERMINT_ADDR = "localhost:8848"

// NewBaseApp extends a StoreApp with a handler and a ticker,
// which it binds to the proper abci calls
func NewBaseApp(store *StoreApp, handler sdk.Handler, clock sdk.Ticker) *BaseApp {
	return &BaseApp{
		StoreApp: store,
		handler:  handler,
		clock:    clock,
	}
}

// DeliverTx - ABCI - dispatches to the handler
func (app *BaseApp) DeliverTx(txBytes []byte) (res abci.ResponseDeliverTx) {
	fmt.Println("DeliverTx")

	var client, err = abcicli.NewClient(ETHERMINT_ADDR, "socket", true)
	client.Start()

	resp, err := client.DeliverTxSync(txBytes)

	fmt.Printf("ethermint DeliverTx response: %v\n", resp)

	if err != nil {
		return errors.DeliverResult(err)
	}

	res.Code = 0

	return res
}

// CheckTx - ABCI - dispatches to the handler
func (app *BaseApp) CheckTx(txBytes []byte) abci.ResponseCheckTx {
	fmt.Println("CheckTx")

	tx, err := decodeTx(txBytes)
	if err != nil {
		// nolint: errcheck
		app.logger.Debug("CheckTx: Received invalid transaction", "tx", tx, "err", err)
		return abci.ResponseCheckTx{
			Code: errors.CodeTypeInternalErr,
			Log:  err.Error(),
		}
	}

	app.logger.Info("CheckTx: Received valid transaction", "tx", tx) // nolint: errcheck

	client, err := abcicli.NewClient(ETHERMINT_ADDR, "socket", true)
	client.Start()

	resp, err := client.CheckTxSync(txBytes)

	fmt.Printf("ethermint CheckTx response: %v\n", resp)

	if err != nil {
		return errors.CheckResult(err)
	}

	return sdk.NewCheck(21000, "").ToABCI()
}

// BeginBlock - ABCI
func (app *BaseApp) BeginBlock(beginBlock abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	fmt.Println("BeginBlock")

	var client, _ = abcicli.NewClient(ETHERMINT_ADDR, "socket", true)
	client.Start()

	resp, _ := client.BeginBlockSync(beginBlock)

	fmt.Printf("ethermint BeginBlock response: %v\n", resp)

	return abci.ResponseBeginBlock{}
}

// EndBlock - ABCI - triggers Tick actions
func (app *BaseApp) EndBlock(endBlock abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	fmt.Println("EndBlock")

	var client, _ = abcicli.NewClient(ETHERMINT_ADDR, "socket", true)
	client.Start()

	resp, _ := client.EndBlockSync(endBlock)

	fmt.Printf("ethermint EndBlock response: %v\n", resp)

	return abci.ResponseEndBlock{}
}

// InitState - used to setup state (was SetOption)
// to be used by InitChain later
//
func (app *BaseApp) InitState(module, key, value string) error {
	return nil
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