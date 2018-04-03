package nonce

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/CyberMiles/travis/types"
)

// nolint
const (
	NameNonce = "nonce"
	CostNonce = 10
)

// Verifies tx is not being replayed
func ReplayCheck(ctx *types.Context, store state.SimpleDB, tx *sdk.Tx) (res sdk.CheckResult, err error) {

	stx, err := checkIncrementNonceTx(ctx, store, *tx)
	if err != nil {
		return res, err
	}

	tx = &stx
	return res, err
}

// checkNonceTx varifies the nonce sequence, an increment sequence number
func checkIncrementNonceTx(ctx *types.Context, store state.SimpleDB, tx sdk.Tx) (sdk.Tx, error) {

	// make sure it is a the nonce Tx (Tx from this package)
	nonceTx, ok := tx.Unwrap().(Tx)
	if !ok {
		return tx, ErrNoNonce()
	}

	err := nonceTx.ValidateBasic()
	if err != nil {
		return tx, err
	}

	// check the nonce sequence number
	err = nonceTx.CheckIncrementSeq(ctx, store)
	if err != nil {
		return tx, err
	}
	return nonceTx.Tx, nil
}
