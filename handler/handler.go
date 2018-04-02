package handler

import (
	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

type Handler struct {
}

func (h Handler) CheckTx(ctx Context, store state.SimpleDB, tx sdk.Tx) (res sdk.CheckResult, err error) {

}

func (h Handler) deliverTx(ctx Context, store state.SimpleDB, tx sdk.Tx) (res sdk.DeliverResult, err error) {

}