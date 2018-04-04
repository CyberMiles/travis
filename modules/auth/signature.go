package auth

import (
	"github.com/cosmos/cosmos-sdk"
	"github.com/ethereum/go-ethereum/common"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/CyberMiles/travis/types"
)

// Signable allows us to use txs.OneSig and txs.MultiSig (and others??)
type Signable interface {
	sdk.TxLayer
	Signers() (common.Address, error)
}

// verifies the signatures are correct
func VerifyTx(ctx *types.Context, tx sdk.Tx) (res sdk.CheckResult, stx sdk.Tx, err error) {
	signers, stx, err := getSigners(tx)
	if err != nil {
		return res, sdk.Tx{}, err
	}
	addSigners(ctx, signers)
	return
}


func addSigners(ctx *types.Context, addr common.Address) {
	ctx.WithSigners(addr)
}

func getSigners(tx sdk.Tx) (common.Address, sdk.Tx, error) {
	stx, ok := tx.Unwrap().(Signable)
	if !ok {
		return common.Address{}, sdk.Tx{}, errors.ErrUnauthorized()
	}
	sig, err := stx.Signers()
	return sig, stx.Next(), err
}