package governance

import (
	"math/big"
	"bytes"
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/CyberMiles/travis/commons"
)

// nolint
const governanceModuleName = "governance"

// Name is the name of the modules.
func Name() string {
	return governanceModuleName
}

func InitState(module, key, value string,store state.SimpleDB) error {
	return nil
}

func CheckTx(ctx types.Context, store state.SimpleDB,
	tx sdk.Tx) (res sdk.CheckResult, err error) {

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
		if !bytes.Equal(txInner.Proposer.Bytes(), sender.Bytes()) {
			return sdk.NewCheck(0,  ""), ErrMissingSignature()
		}
		candidate := stake.GetCandidateByAddress(txInner.Proposer)
		if candidate == nil || candidate.State != "Y" || candidate.VotingPower == 0 {
			return sdk.NewCheck(0, ""), ErrInvalidValidator()
		}
	case TxVote:
		if !bytes.Equal(txInner.Voter.Bytes(), sender.Bytes()) {
			return sdk.NewCheck(0,  ""), ErrMissingSignature()
		}
		validator := stake.GetCandidateByAddress(txInner.Voter)
		if validator == nil || validator.State != "Y" {
			return sdk.NewCheck(0, ""), ErrInvalidValidator()
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
func DeliverTx(ctx types.Context, store state.SimpleDB,
	tx sdk.Tx, hash []byte) (res sdk.DeliverResult, err error) {

	_, err = CheckTx(ctx, store, tx)
	if err != nil {
		return
	}

	switch txInner := tx.Unwrap().(type) {
	case TxPropose:
		amount := new(big.Int)
		amount.SetString(txInner.Amount, 10)

		pp := NewProposal(
			hex.EncodeToString(hash),
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

		votes := GetVotesByPid(txInner.ProposalId)
		validators := stake.GetCandidates().Validators()

		if validators == nil || validators.Len() == 0 {
			return
		}

		if len(votes) * 3 < len(validators) * 2 {
			return
		}

		var c int
		for _, vo := range votes {
			for _, va := range validators {
				if bytes.Equal(vo.Voter.Bytes(), va.OwnerAddress.Bytes()) {
					c++
					continue
				}
			}
		}

		if c * 3 >= len(validators) * 2 {
			proposal := GetProposalById(txInner.ProposalId)
			commons.Transfer(proposal.From, proposal.To, proposal.Amount)
		}
	}

	return
}

// get the sender from the ctx and ensure it matches the tx pubkey
func getTxSender(ctx types.Context) (sender common.Address, err error) {
	senders := ctx.GetSigners()
	if len(senders) != 1 {
		return sender, ErrMissingSignature()
	}
	return senders[0], nil
}
