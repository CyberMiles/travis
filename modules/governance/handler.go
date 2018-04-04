package governance

import (
	"math/big"
	"bytes"
	"encoding/hex"

	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/CyberMiles/travis/utils"
)

// nolint
const governanceModuleName = "governance"

// Name is the name of the modules.
func Name() string {
	return governanceModuleName
}

type Handler struct {
	stack.PassInitValidate
}

var _ stack.Dispatchable = Handler{} // enforce interface at compile time

// NewHandler returns a new Handler with the default Params
func NewHandler() Handler {
	return Handler{}
}

// Name - return stake namespace
func (Handler) Name() string {
	return governanceModuleName
}

func (h Handler) InitState(l log.Logger, store state.SimpleDB,
	module, key, value string, cb sdk.InitStater) (log string, err error) {
	return
}

// AssertDispatcher - placeholder for stack.Dispatchable
func (Handler) AssertDispatcher() {}

func (h Handler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, dispatch sdk.Checker) (res sdk.CheckResult, err error) {

	err = tx.ValidateBasic()
	if err != nil {
		return
	}

	// get the sender
	sender, err := getTxSender(ctx)
	if err != nil {
		return
	}

	switch txInner := tx.Unwrap().(type) {
	case TxPropose:
		if !bytes.Equal(txInner.Proposer.Bytes(), sender.Address.Bytes()) {
			return sdk.NewCheck(0,  ""), ErrMissingSignature()
		}
	case TxVote:
		if !bytes.Equal(txInner.Voter.Bytes(), sender.Address.Bytes()) {
			return sdk.NewCheck(0,  ""), ErrMissingSignature()
		}
		if proposal := GetProposalById(txInner.ProposalId); proposal == nil {
			return sdk.NewCheck(0, ""), ErrInvalidParamerter()
		}
		if vote := GetVoteByPidAndVoter(txInner.ProposalId, txInner.Voter.String()); vote != nil {
			return sdk.NewCheck(0, ""), ErrRepeatedVote()
		}
	}

	return
}

// DeliverTx executes the tx if valid
func (h Handler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, dispatch sdk.Deliver) (res sdk.DeliverResult, err error) {

	_, err = h.CheckTx(ctx, store, tx, nil)
	if err != nil {
		return
	}

	switch txInner := tx.Unwrap().(type) {
	case TxPropose:
		pid := utils.GetUUID()
		amount := new(big.Int)
		amount.SetString(txInner.Amount, 10)

		pp := NewProposal(
			hex.EncodeToString(pid),
			txInner.Proposer,
			uint64(ctx.BlockHeight()),
			txInner.From,
			txInner.To,
			amount,
			txInner.Reason,
		)

		SaveProposal(pp)

	case TxVote:
		vote := NewVote(
			txInner.ProposalId,
			txInner.Voter,
			uint64(ctx.BlockHeight()),
			txInner.Answer,
		)
		SaveVote(vote)
	}

	return
}

// get the sender from the ctx and ensure it matches the tx pubkey
func getTxSender(ctx sdk.Context) (sender sdk.Actor, err error) {
	senders := ctx.GetPermissions("", auth.NameSigs)
	if len(senders) != 1 {
		return sender, ErrMissingSignature()
	}
	return senders[0], nil
}
