package app

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	tmLog "github.com/tendermint/tendermint/libs/log"

	"github.com/CyberMiles/travis/api"
	"github.com/CyberMiles/travis/errors"
	"github.com/CyberMiles/travis/utils"
	emtTypes "github.com/CyberMiles/travis/vm/types"
)

type FromTo struct {
	from common.Address
	to   common.Address
}

// EthermintApplication implements an ABCI application
// #stable - 0.4.0
type EthermintApplication struct {

	// backend handles the ethereum state machine
	// and wrangles other services started by an ethereum node (eg. tx pool)
	backend *api.Backend // backend ethereum struct

	checkTxState *state.StateDB

	// an ethereum rpc client we can forward queries to
	rpcClient *rpc.Client

	// strategy for validator compensation
	strategy *emtTypes.Strategy

	logger tmLog.Logger

	lowPriceCheckTransactions   map[FromTo]struct{}
	lowPriceDeliverTransactions map[FromTo]struct{}
}

// NewEthermintApplication creates a fully initialised instance of EthermintApplication
// #stable - 0.4.0
func NewEthermintApplication(backend *api.Backend,
	client *rpc.Client, strategy *emtTypes.Strategy) (*EthermintApplication, error) {

	state := backend.ManagedState()
	if state == nil {
		panic("Error getting latest state")
	}

	app := &EthermintApplication{
		backend:                     backend,
		rpcClient:                   client,
		checkTxState:                state.StateDB,
		strategy:                    strategy,
		lowPriceCheckTransactions:   make(map[FromTo]struct{}),
		lowPriceDeliverTransactions: make(map[FromTo]struct{}),
	}

	if err := app.backend.InitEthState(app.Receiver()); err != nil {
		return nil, err
	}

	return app, nil
}

// SetLogger sets the logger for the ethermint application
// #unstable
func (app *EthermintApplication) SetLogger(log tmLog.Logger) {
	app.logger = log
}

var bigZero = big.NewInt(0)

// maxTransactionSize is 32KB in order to prevent DOS attacks
const maxTransactionSize = 32768

// Info returns information about the last height and app_hash to the tendermint engine
// #stable - 0.4.0

func (app *EthermintApplication) Info(req abciTypes.RequestInfo) abciTypes.ResponseInfo {
	blockchain := app.backend.Ethereum().BlockChain()
	currentBlock := blockchain.CurrentBlock()
	height := currentBlock.Number()
	hash := currentBlock.Hash()

	app.logger.Debug("Info", "height", height) // nolint: errcheck

	// This check determines whether it is the first time ethermint gets started.
	// If it is the first time, then we have to respond with an empty hash, since
	// that is what tendermint expects.
	if height.Cmp(bigZero) == 0 {
		return abciTypes.ResponseInfo{
			Data:             "ABCIEthereum",
			LastBlockHeight:  height.Int64(),
			LastBlockAppHash: []byte{},
		}
	}

	return abciTypes.ResponseInfo{
		Data:             "ABCIEthereum",
		LastBlockHeight:  height.Int64(),
		LastBlockAppHash: hash[:],
	}
}

// SetOption sets a configuration option
// #stable - 0.4.0
func (app *EthermintApplication) SetOption(req abciTypes.RequestSetOption) abciTypes.ResponseSetOption {

	app.logger.Debug("SetOption", "key", req.GetKey(), "value", req.GetValue()) // nolint: errcheck
	return abciTypes.ResponseSetOption{}
}

// InitChain initializes the validator set
// #stable - 0.4.0
func (app *EthermintApplication) InitChain(req abciTypes.RequestInitChain) abciTypes.ResponseInitChain {

	app.logger.Debug("InitChain") // nolint: errcheck
	app.SetValidators(req.GetValidators())
	return abciTypes.ResponseInitChain{}
}

// CheckTx checks a transaction is valid but does not mutate the state
// #stable - 0.4.0
func (app *EthermintApplication) CheckTx(tx *ethTypes.Transaction) abciTypes.ResponseCheckTx {
	app.logger.Debug("CheckTx: Received valid transaction", "tx", tx) // nolint: errcheck

	return app.validateTx(tx)
}

// DeliverTx executes a transaction against the latest state
// #stable - 0.4.0
func (app *EthermintApplication) DeliverTx(tx *ethTypes.Transaction) abciTypes.ResponseDeliverTx {
	app.logger.Debug("DeliverTx: Received valid transaction", "tx", tx) // nolint: errcheck

	networkId := big.NewInt(int64(app.backend.Ethereum().NetVersion()))
	signer := ethTypes.NewEIP155Signer(networkId)
	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		// TODO: Add errors.CodeTypeInvalidSignature ?
		return abciTypes.ResponseDeliverTx{
			Code: errors.CodeTypeInternalErr,
			Log:  err.Error()}
	}
	if code, errLog := app.lowPriceTxCheck(from, tx, app.lowPriceDeliverTransactions); code != abciTypes.CodeTypeOK {
		return abciTypes.ResponseDeliverTx{Code: code, Log: errLog}
	}

	res := app.backend.DeliverTx(tx)
	if res.IsErr() {
		// nolint: errcheck
		app.logger.Error("DeliverTx: Error delivering tx to ethereum backend", "tx", tx,
			"err", res.String())
		return res
	}
	app.CollectTx(tx)

	return abciTypes.ResponseDeliverTx{
		Code: abciTypes.CodeTypeOK,
	}
}

func (app *EthermintApplication) DeliverTxState() *state.StateDB {
	return app.backend.DeliverTxState()
}

// BeginBlock starts a new Ethereum block
// #stable - 0.4.0
func (app *EthermintApplication) BeginBlock(beginBlock abciTypes.RequestBeginBlock) abciTypes.ResponseBeginBlock {

	app.logger.Debug("BeginBlock") // nolint: errcheck

	// update the eth header with the tendermint header
	app.backend.UpdateHeaderWithTimeInfo(beginBlock.GetHeader(), beginBlock.GetHash())
	return abciTypes.ResponseBeginBlock{}
}

// EndBlock accumulates rewards for the validators and updates them
// #stable - 0.4.0
func (app *EthermintApplication) EndBlock(endBlock abciTypes.RequestEndBlock) abciTypes.ResponseEndBlock {

	app.logger.Debug("EndBlock", "height", endBlock.GetHeight()) // nolint: errcheck
	app.backend.AccumulateRewards(app.backend.Ethereum().BlockChain().Config(), app.strategy)

	app.backend.EndBlock()

	return app.GetUpdatedValidators()
}

// Commit commits the block and returns a hash of the current state
// #stable - 0.4.0
func (app *EthermintApplication) Commit() (abciTypes.ResponseCommit, error) {
	app.logger.Debug("Commit") // nolint: errcheck
	blockHash, err := app.backend.Commit(app.Receiver())
	if err != nil {
		// nolint: errcheck
		app.logger.Error("Error getting latest ethereum state", "err", err)
		return abciTypes.ResponseCommit{
			Data: blockHash[:],
		}, err
	}

	state, err := app.backend.ResetState()
	if err != nil {
		app.logger.Error("Error getting latest state", "err", err) // nolint: errcheck
		return abciTypes.ResponseCommit{
			Data: blockHash[:],
		}, err
	}
	app.checkTxState = state.StateDB

	app.lowPriceCheckTransactions = make(map[FromTo]struct{})
	app.lowPriceDeliverTransactions = make(map[FromTo]struct{})

	return abciTypes.ResponseCommit{
		Data: blockHash[:],
	}, nil
}

// Query queries the state of the EthermintApplication
// #stable - 0.4.0
func (app *EthermintApplication) Query(query abciTypes.RequestQuery) abciTypes.ResponseQuery {
	app.logger.Debug("Query") // nolint: errcheck
	var in jsonRequest
	if err := json.Unmarshal(query.Data, &in); err != nil {
		return abciTypes.ResponseQuery{Code: errors.CodeTypeInternalErr,
			Log: err.Error()}
	}
	var result interface{}
	if err := app.rpcClient.Call(&result, in.Method, in.Params...); err != nil {
		return abciTypes.ResponseQuery{Code: errors.CodeTypeInternalErr,
			Log: err.Error()}
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return abciTypes.ResponseQuery{Code: errors.CodeTypeInternalErr,
			Log: err.Error()}
	}
	return abciTypes.ResponseQuery{Code: abciTypes.CodeTypeOK, Value: bytes}
}

//-------------------------------------------------------

// validateTx checks the validity of a tx against the blockchain's current state.
// it duplicates the logic in ethereum's tx_pool
func (app *EthermintApplication) validateTx(tx *ethTypes.Transaction) abciTypes.ResponseCheckTx {

	currentState, from, nonce, resp := app.basicCheck(tx)
	if resp.Code != abciTypes.CodeTypeOK {
		return resp
	}

	// Transactor should have enough funds to cover the costs
	currentBalance := currentState.GetBalance(from)

	// cost == V + GP * GL
	if currentBalance.Cmp(tx.Cost()) < 0 &&
		(tx.To() == nil || len(tx.Data()) == 0 || currentState.GetBalance(*tx.To()).Cmp(tx.Cost()) < 0) {
		return abciTypes.ResponseCheckTx{
			// TODO: Add errors.CodeTypeInsufficientFunds ?
			Code: errors.CodeTypeBaseInvalidInput,
			Log: fmt.Sprintf(
				"Current balance: %s, tx cost: %s",
				currentBalance, tx.Cost())}
	}

	intrGas, err := core.IntrinsicGas(tx.Data(), tx.To() == nil, true) // homestead == true
	if err != nil {
		return abciTypes.ResponseCheckTx{
			Code: errors.CodeTypeBaseInvalidInput,
			Log:  err.Error()}
	}
	if tx.Gas() < intrGas {
		return abciTypes.ResponseCheckTx{
			Code: errors.CodeTypeBaseInvalidInput,
			Log:  core.ErrIntrinsicGas.Error()}
	}

	if code, errLog := app.lowPriceTxCheck(from, tx, app.lowPriceCheckTransactions); code != abciTypes.CodeTypeOK {
		return abciTypes.ResponseCheckTx{Code: code, Log: errLog}
	}

	// Update ether balances
	// amount + gasprice * gaslimit
	currentState.SubBalance(from, tx.Cost())
	// tx.To() returns a pointer to a common address. It returns nil
	// if it is a contract creation transaction.
	if to := tx.To(); to != nil {
		currentState.AddBalance(*to, tx.Value())
	}
	currentState.SetNonce(from, nonce+1)

	return abciTypes.ResponseCheckTx{Code: abciTypes.CodeTypeOK}
}

func (app *EthermintApplication) lowPriceTxCheck(from common.Address, tx *ethTypes.Transaction, lowPriceTxs map[FromTo]struct{}) (uint32, string) {
	// Iterate over all transactions to check if the gas price is too low for the
	// non-first transaction with the same from/to address
	// Todo performance maybe
	var to common.Address
	if tx.To() != nil {
		to = *tx.To()
	}
	ft := FromTo{from: from, to: to}

	if tx.GasPrice().Cmp(new(big.Int).SetUint64(utils.GetParams().GasPrice)) < 0 {
		if _, ok := lowPriceTxs[ft]; ok {
			return errors.CodeLowGasPriceErr, "The gas price is too low for transaction"
		}
		if tx.Gas() > utils.GetParams().LowPriceTxGasLimit {
			return errors.CodeHighGasLimitErr, "The gas limit is too high for low price transaction"
		}
		if len(lowPriceTxs) > utils.GetParams().LowPriceTxSlotsCap {
			return errors.CodeLowPriceTxCapErr, "The capacity of one block is reached for low price transactions"
		}
		lowPriceTxs[ft] = struct{}{}
	}

	return abciTypes.CodeTypeOK, ""
}
