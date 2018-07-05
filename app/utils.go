package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	abciTypes "github.com/tendermint/tendermint/abci/types"

	"github.com/CyberMiles/travis/errors"
	"github.com/CyberMiles/travis/utils"
)

// format of query data
type jsonRequest struct {
	Method string          `json:"method"`
	ID     json.RawMessage `json:"id,omitempty"`
	Params []interface{}   `json:"params,omitempty"`
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

//-------------------------------------------------------
// convenience methods for validators

// Receiver returns the receiving address based on the selected strategy
// #unstable
func (app *EthermintApplication) Receiver() common.Address {
	if app.strategy != nil {
		return app.strategy.Receiver()
	}
	return utils.HoldAccount
}

// SetValidators sets new validators on the strategy
// #unstable
func (app *EthermintApplication) SetValidators(validators []abciTypes.Validator) {
	if app.strategy != nil {
		app.strategy.SetValidators(validators)
	}
}

// GetUpdatedValidators returns an updated validator set from the strategy
// #unstable
func (app *EthermintApplication) GetUpdatedValidators() abciTypes.ResponseEndBlock {
	if app.strategy != nil {
		return abciTypes.ResponseEndBlock{ValidatorUpdates: app.strategy.GetUpdatedValidators()}
	}
	return abciTypes.ResponseEndBlock{}
}

// CollectTx invokes CollectTx on the strategy
// #unstable
func (app *EthermintApplication) CollectTx(tx *types.Transaction) {
	if app.strategy != nil {
		app.strategy.CollectTx(tx)
	}
}

func (app *EthermintApplication) basicCheck(tx *ethTypes.Transaction) (*state.StateDB, common.Address, uint64, abciTypes.ResponseCheckTx) {

	// Heuristic limit, reject transactions over 32KB to prevent DOS attacks
	if tx.Size() > maxTransactionSize {
		return nil, common.Address{}, 0,
			abciTypes.ResponseCheckTx{
				Code: errors.CodeTypeInternalErr,
				Log:  core.ErrOversizedData.Error()}
	}

	// tx.ChainID() must > 0
	if tx.ChainId().Cmp(big.NewInt(0)) <= 0 {
		return nil, common.Address{}, 0,
			abciTypes.ResponseCheckTx{
				Code: errors.CodeTypeInternalErr,
				Log:  types.ErrInvalidChainId.Error()}
	}

	networkId := big.NewInt(int64(app.backend.Ethereum().NetVersion()))
	signer := ethTypes.NewEIP155Signer(networkId)

	// Make sure the transaction is signed properly
	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		// TODO: Add errors.CodeTypeInvalidSignature ?
		return nil, common.Address{}, 0,
			abciTypes.ResponseCheckTx{
				Code: errors.CodeTypeInternalErr,
				Log:  err.Error()}
	}

	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Value().Sign() < 0 {
		return nil, common.Address{}, 0,
			abciTypes.ResponseCheckTx{
				Code: errors.CodeTypeBaseInvalidInput,
				Log:  core.ErrNegativeValue.Error()}
	}

	currentState := app.checkTxState

	// Make sure the account exist - cant send from non-existing account.
	if !currentState.Exist(from) {
		return nil, common.Address{}, 0,
			abciTypes.ResponseCheckTx{
				Code: errors.CodeTypeUnknownAddress,
				Log:  core.ErrInvalidSender.Error()}
	}

	// Check the transaction doesn't exceed the current block limit gas.
	gasLimit := app.backend.GasLimit()
	if gasLimit.Cmp(tx.Gas()) < 0 {
		return nil, common.Address{}, 0,
			abciTypes.ResponseCheckTx{
				Code: errors.CodeTypeInternalErr,
				Log:  core.ErrGasLimitReached.Error()}
	}

	nonce := currentState.GetNonce(from)
	if _, ok := utils.NonceCheckedTx[tx.Hash()]; !ok {
		// Check if nonce is not strictly increasing
		// if not then recheck with feeding failed count
		if nonce != tx.Nonce() {
			if c, ok := app.checkFailedCount[from]; ok {
				if nonce+c != tx.Nonce() {
					return nil, common.Address{}, 0,
						abciTypes.ResponseCheckTx{
							Code: errors.CodeTypeBadNonce,
							Log: fmt.Sprintf(
								"Nonce not strictly increasing. Expected %d Got %d",
								nonce, tx.Nonce())}
				}
			} else {
				return nil, common.Address{}, 0,
					abciTypes.ResponseCheckTx{
						Code: errors.CodeTypeBadNonce,
						Log: fmt.Sprintf(
							"Nonce not strictly increasing. Expected %d Got %d",
							nonce, tx.Nonce())}
			}
		}
	}

	return currentState, from, nonce, abciTypes.ResponseCheckTx{Code: abciTypes.CodeTypeOK}
}
