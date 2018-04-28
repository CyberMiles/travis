package governance

import (
	"bytes"
	"strings"
	"encoding/hex"
	"math/big"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/ethereum/go-ethereum/common"
)

// default proposal expiration = block count of 7 days
const defaultProposalExpire uint64 = 7 * 24 * 60 * 60 / 10

// nolint
const governanceModuleName = "governance"

// Name is the name of the modules.
func Name() string {
	return governanceModuleName
}

func InitState(module, key, value string, store state.SimpleDB) error {
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
			return sdk.NewCheck(0, ""), ErrMissingSignature()
		}
		candidate := stake.GetCandidateByAddress(*txInner.Proposer)
		if candidate == nil || candidate.VotingPower == 0 {
			return sdk.NewCheck(0, ""), ErrInvalidValidator()
		}

		ethereum := ctx.Ethereum()
		balance, err := commons.GetBalance(ethereum, *txInner.From)
		if err != nil {
			return sdk.NewCheck(0, ""), ErrInvalidParamerter()
		}

		amount := big.NewInt(0)
		amount.SetString(txInner.Amount, 10)
		if balance.Cmp(amount) < 0 {
			return sdk.NewCheck(0, ""), ErrInsufficientBalance()
		}
		utils.TravisTxAddrs = append(utils.TravisTxAddrs, txInner.From)
	case TxVote:
		if !bytes.Equal(txInner.Voter.Bytes(), sender.Bytes()) {
			return sdk.NewCheck(0, ""), ErrMissingSignature()
		}
		validator := stake.GetCandidateByAddress(txInner.Voter)
		if validator == nil {
			return sdk.NewCheck(0, ""), ErrInvalidValidator()
		}

		proposal := GetProposalById(txInner.ProposalId)
		if proposal == nil {
			return sdk.NewCheck(0, ""), ErrInvalidParamerter()
		}
		if proposal.ResultBlockHeight != 0 {
			if proposal.Result == "Approved" {
				return sdk.NewCheck(0, ""), ErrApprovedProposal()
			} else {
				return sdk.NewCheck(0, ""), ErrRejectedProposal()
			}
		}
		if vote := GetVoteByPidAndVoter(txInner.ProposalId, txInner.Voter.String()); vote != nil {
			return sdk.NewCheck(0, ""), ErrRepeatedVote()
		}
		utils.TravisTxAddrs = append(utils.TravisTxAddrs, proposal.To)
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
		expire := defaultProposalExpire
		if txInner.Expire != 0 {
			expire = txInner.Expire
		}
		pp := NewProposal(
			hex.EncodeToString(hash),
			txInner.Proposer,
			uint64(ctx.BlockHeight()),
			txInner.From,
			txInner.To,
			txInner.Amount,
			txInner.Reason,
			uint64(ctx.BlockHeight()) + expire,
		)

		SaveProposal(pp)
		amount := big.NewInt(0)
		amount.SetString(pp.Amount, 10)
		commons.TransferWithReactor(*pp.From, utils.EmptyAddress, amount, ProposalReactor{pp.Id, uint64(ctx.BlockHeight()), ""})

		utils.PendingProposal.Add(pp.Id, pp.ExpireBlockHeight)

		res.Data = hash

	case TxVote:
		vote := NewVote(
			txInner.ProposalId,
			txInner.Voter,
			uint64(ctx.BlockHeight()),
			txInner.Answer,
		)
		SaveVote(vote)

		proposal := GetProposalById(txInner.ProposalId)
		amount := new(big.Int)
		amount.SetString(proposal.Amount, 10)

		switch CheckProposal(txInner.ProposalId) {
		case "approved":
			// as succeeded proposal only need to add balance to receiver,
			// so the transfer should always be successful
			// but we still use the reactor to keep the compatible with the old strategy
			commons.TransferWithReactor(utils.EmptyAddress, *proposal.To, amount, ProposalReactor{proposal.Id, uint64(ctx.BlockHeight()), "Approved"})
			utils.PendingProposal.Del(proposal.Id)
		case "rejected":
			// as succeeded proposal only need to refund balance to sender,
			// so the transfer should always be successful
			// but we still use the reactor to keep the compatible with the old strategy
			commons.TransferWithReactor(utils.EmptyAddress, *proposal.From, amount, ProposalReactor{proposal.Id, uint64(ctx.BlockHeight()), "Rejected"})
			utils.PendingProposal.Del(proposal.Id)
		}
	}

	return
}

func CheckProposal(pid string) string {
	votes := GetVotesByPid(pid)
	validators := stake.GetCandidates().Validators()

	if validators == nil || validators.Len() == 0 {
		return "no validator"
	}

	if len(votes) * 3 < len(validators) * 2 {
		return "not enough vote"
	}

	var approvedCount, rejectedCount int
	for _, vo := range votes {
		for _, va := range validators {
			// should check voter is still valid validator first
			if bytes.Equal(vo.Voter.Bytes(), va.OwnerAddress.Bytes()) {
				if strings.Compare(vo.Answer, "Y") == 0 {
					approvedCount++
				}
				if strings.Compare(vo.Answer, "N") == 0 {
					rejectedCount++
				}
				continue
			}
		}
	}

	if approvedCount * 3 >= len(validators) * 2 {
		// To avoid repeated commit, let's recheck with count of voters - 1
		if (approvedCount - 1) * 3 < len(validators) * 2 {
			return "approved"
		}
	} else if rejectedCount * 3 >= len(validators) * 2 {
		// To avoid repeated commit, let's recheck with count of voters - 1
		if (rejectedCount - 1) * 3 < len(validators) * 2 {
			return "rejected"
		}
	}
	return "not determined"
}

type ProposalReactor struct {
	ProposalId string
	BlockHeight uint64
	Result string
}

func (pr ProposalReactor) React(result, msg string) {
	now := utils.GetNow()
	if result == "success" {
		// If the default result is not set, then do nothing
		if pr.Result == "" {
			return
		}
		result = pr.Result
	}
	UpdateProposalResult(pr.ProposalId, result, msg, pr.BlockHeight, now)
}

// get the sender from the ctx and ensure it matches the tx pubkey
func getTxSender(ctx types.Context) (sender common.Address, err error) {
	senders := ctx.GetSigners()
	if len(senders) != 1 {
		return sender, ErrMissingSignature()
	}
	return senders[0], nil
}
