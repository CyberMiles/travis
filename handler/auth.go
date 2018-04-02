package handler

import (
	"github.com/cosmos/cosmos-sdk"
	"github.com/ethereum/go-ethereum/common"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/cosmos/cosmos-sdk/errors"
)

// SigPerm takes the binary address from PubKey.Address and makes it an Actor
func SigPerm(addr []byte) sdk.Actor {
	return sdk.NewActor(NameSigs, addr)
}

// Signable allows us to use txs.OneSig and txs.MultiSig (and others??)
type Signable interface {
	sdk.TxLayer
	Signers() (common.Address, error)
}

// verifies the signatures are correct
func VerifyTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx) error {
	sigs, tnext, err := getSigners(tx)
	if err != nil {
		return err
	}
	addSigners(ctx, sigs)
}


func addSigners(ctx sdk.Context, addr common.Address) sdk.Context {
	perms := make([]sdk.Actor, 1)
	perms[0] = SigPerm(addr.Bytes())
	return ctx.WithPermissions(perms...)
}

func getSigners(tx sdk.Tx) (common.Address, sdk.Tx, error) {
	stx, ok := tx.Unwrap().(Signable)
	if !ok {
		return common.Address{}, sdk.Tx{}, errors.ErrUnauthorized()
	}
	sig, err := stx.Signers()
	return sig, stx.Next(), err
}