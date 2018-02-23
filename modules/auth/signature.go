package auth

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/ethereum/go-ethereum/common"
)

//nolint
const (
	NameSigs = "sigs"
)

// Signatures parses out go-crypto signatures and adds permissions to the
// context for use inside the application
type Signatures struct {
	stack.PassInitState
	stack.PassInitValidate
}

// Name of the module - fulfills Middleware interface
func (Signatures) Name() string {
	return NameSigs
}

var _ stack.Middleware = Signatures{}

// SigPerm takes the binary address from PubKey.Address and makes it an Actor
func SigPerm(addr []byte) sdk.Actor {
	return sdk.NewActor(NameSigs, addr)
}

// Signable allows us to use txs.OneSig and txs.MultiSig (and others??)
type Signable interface {
	sdk.TxLayer
	Signers() (common.Address, error)
}

// CheckTx verifies the signatures are correct - fulfills Middlware interface
func (Signatures) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Checker) (res sdk.CheckResult, err error) {
	sigs, tnext, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.CheckTx(ctx2, store, tnext)
}

// DeliverTx verifies the signatures are correct - fulfills Middlware interface
func (Signatures) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx, next sdk.Deliver) (res sdk.DeliverResult, err error) {
	sigs, tnext, err := getSigners(tx)
	if err != nil {
		return res, err
	}
	ctx2 := addSigners(ctx, sigs)
	return next.DeliverTx(ctx2, store, tnext)
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
