package governance

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/state"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"github.com/CyberMiles/travis/version"
	"github.com/ethereum/go-ethereum/common"
	ethState "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm/eni"
	"net/rpc"

)

// nolint
const governanceModuleName = "governance"

var OTAInstance = eni.NewOTAInstance()

var cancelDownload = make(map[string]bool)

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
	app_state := ctx.EthappState()

	switch txInner := tx.Unwrap().(type) {
	case TxTransferFundPropose:
		if txInner.Proposer == nil || !bytes.Equal(txInner.Proposer.Bytes(), sender.Bytes()) {
			return sdk.NewCheck(0, ""), ErrMissingSignature()
		}
		validators := stake.GetCandidates().Validators()
		if validators == nil || validators.Len() == 0 {
			return sdk.NewCheck(0, ""), ErrInvalidValidator()
		}
		for i, v := range validators {
			if v.OwnerAddress == txInner.Proposer.String() {
				break
			}
			if i+1 == len(validators) {
				return sdk.NewCheck(0, ""), ErrInvalidValidator()
			}
		}

		amount := big.NewInt(0)
		amount.SetString(txInner.Amount, 10)
		if amount.Cmp(big.NewInt(0)) <= 0 {
			return sdk.NewCheck(0, ""), ErrInvalidParameter()
		}

		balance, err := commons.GetBalance(app_state, *txInner.From)
		if err != nil {
			return sdk.NewCheck(0, ""), ErrInvalidParameter()
		}

		if balance.Cmp(amount) < 0 {
			return sdk.NewCheck(0, ""), ErrInsufficientBalance()
		}

		if txInner.ExpireTimestamp != nil && txInner.ExpireBlockHeight != nil {
			return sdk.NewCheck(0, ""), ErrExceedsExpiration()
		}

		if txInner.ExpireTimestamp != nil && ctx.BlockTime() > *txInner.ExpireTimestamp {
			return sdk.NewCheck(0, ""), ErrInvalidExpireTimestamp()
		}

		if txInner.ExpireBlockHeight != nil && ctx.BlockHeight() >= *txInner.ExpireBlockHeight {
			return sdk.NewCheck(0, ""), ErrInvalidExpireBlockHeight()
		}

		// Transfer gasFee
		gasFee, err := checkGasFee(app_state, sender, utils.GetParams().TransferFundProposalGas)
		if err != nil {
			return sdk.NewCheck(0, ""), err
		}
		app_state.SubBalance(*txInner.From, amount)
		app_state.SubBalance(sender, gasFee.Int)

	case TxChangeParamPropose:
		if txInner.Proposer == nil || !bytes.Equal(txInner.Proposer.Bytes(), sender.Bytes()) {
			return sdk.NewCheck(0, ""), ErrMissingSignature()
		}
		validators := stake.GetCandidates().Validators()
		if validators == nil || validators.Len() == 0 {
			return sdk.NewCheck(0, ""), ErrInvalidValidator()
		}
		for i, v := range validators {
			if v.OwnerAddress == txInner.Proposer.String() {
				break
			}
			if i+1 == len(validators) {
				return sdk.NewCheck(0, ""), ErrInvalidValidator()
			}
		}

		if txInner.ExpireTimestamp != nil && txInner.ExpireBlockHeight != nil {
			return sdk.NewCheck(0, ""), ErrExceedsExpiration()
		}

		if txInner.ExpireTimestamp != nil && ctx.BlockTime() > *txInner.ExpireTimestamp {
			return sdk.NewCheck(0, ""), ErrInvalidExpireTimestamp()
		}

		if txInner.ExpireBlockHeight != nil && ctx.BlockHeight() >= *txInner.ExpireBlockHeight {
			return sdk.NewCheck(0, ""), ErrInvalidExpireBlockHeight()
		}

		if !utils.CheckParamType(txInner.Name, txInner.Value) {
			return sdk.NewCheck(0, ""), ErrInvalidParameter()
		}

		// Transfer gasFee
		gasFee, err := checkGasFee(app_state, sender, utils.GetParams().ChangeParamsProposalGas)
		if err != nil {
			return sdk.NewCheck(0, ""), err
		}
		app_state.SubBalance(sender, gasFee.Int)
	case TxDeployLibEniPropose:
		if txInner.Proposer == nil || !bytes.Equal(txInner.Proposer.Bytes(), sender.Bytes()) {
			return sdk.NewCheck(0, ""), ErrMissingSignature()
		}
		validators := stake.GetCandidates().Validators()
		if validators == nil || validators.Len() == 0 {
			return sdk.NewCheck(0, ""), ErrInvalidValidator()
		}
		for i, v := range validators {
			if v.OwnerAddress == txInner.Proposer.String() {
				break
			}
			if i+1 == len(validators) {
				return sdk.NewCheck(0, ""), ErrInvalidValidator()
			}
		}

		if strings.Trim(txInner.Name, " ") == "" || strings.Trim(txInner.Version, " ") == "" {
			return sdk.NewCheck(0, ""), ErrInsufficientParameters()
		}

		if txInner.ExpireTimestamp != nil && txInner.ExpireBlockHeight != nil {
			return sdk.NewCheck(0, ""), ErrExceedsExpiration()
		}

		if txInner.ExpireTimestamp != nil && ctx.BlockTime() > *txInner.ExpireTimestamp {
			return sdk.NewCheck(0, ""), ErrInvalidExpireTimestamp()
		}

		if txInner.ExpireBlockHeight != nil && ctx.BlockHeight() >= *txInner.ExpireBlockHeight {
			return sdk.NewCheck(0, ""), ErrInvalidExpireBlockHeight()
		}

		otaInfo := eni.OTAInfo{
			LibName: txInner.Name,
			Version: txInner.Version,
		}
		if valid, _ := OTAInstance.IsValidNewLib(otaInfo); !valid {
			return sdk.NewCheck(0, ""), ErrInvalidNewLib()
		}

		if HasUndeployedProposal(txInner.Name) {
			return sdk.NewCheck(0, ""), ErrOngoingLibFound()
		}

		var fileurlJson map[string][]string

		if err = json.Unmarshal([]byte(txInner.Fileurl), &fileurlJson); err != nil {
			return sdk.NewCheck(0, ""), ErrInvalidFileurlJson()
		}

		if _, ok := fileurlJson[utils.GOOSDIST]; !ok {
			return sdk.NewCheck(0, ""), ErrNoFileurl()
		}

		var md5Json map[string]string

		if err = json.Unmarshal([]byte(txInner.Md5), &md5Json); err != nil {
			return sdk.NewCheck(0, ""), ErrInvalidMd5Json()
		}

		if _, ok := md5Json[utils.GOOSDIST]; !ok {
			return sdk.NewCheck(0, ""), ErrNoMd5()
		}

		// Transfer gasFee
		gasFee, err := checkGasFee(app_state, sender, utils.GetParams().DeployLibEniProposalGas)
		if err != nil {
			return sdk.NewCheck(0, ""), err
		}
		app_state.SubBalance(sender, gasFee.Int)
	case TxRetireProgramPropose:
		if txInner.Proposer == nil || !bytes.Equal(txInner.Proposer.Bytes(), sender.Bytes()) {
			return sdk.NewCheck(0, ""), ErrMissingSignature()
		}
		validators := stake.GetCandidates().Validators()
		if validators == nil || validators.Len() == 0 {
			return sdk.NewCheck(0, ""), ErrInvalidValidator()
		}
		for i, v := range validators {
			if v.OwnerAddress == txInner.Proposer.String() {
				break
			}
			if i+1 == len(validators) {
				return sdk.NewCheck(0, ""), ErrInvalidValidator()
			}
		}

		rp := GetRetiringProposal(version.Version)
		if rp != nil {
			return sdk.NewCheck(0, ""), ErrOngoingRetiringFound()
		}

		if txInner.PreservedValidators == "" {
			return sdk.NewCheck(0, ""), ErrInvalidParameter()
		}

		if txInner.ExpireBlockHeight != nil && ctx.BlockHeight() >= *txInner.ExpireBlockHeight {
			return sdk.NewCheck(0, ""), ErrInvalidExpireBlockHeight()
		}

		// Transfer gasFee
		gasFee, err := checkGasFee(app_state, sender, utils.GetParams().RetireProgramProposalGas)
		if err != nil {
			return sdk.NewCheck(0, ""), err
		}
		app_state.SubBalance(sender, gasFee.Int)
	case TxUpgradeProgramPropose:
		if txInner.Proposer == nil || !bytes.Equal(txInner.Proposer.Bytes(), sender.Bytes()) {
			return sdk.NewCheck(0, ""), ErrMissingSignature()
		}
		validators := stake.GetCandidates().Validators()
		if validators == nil || validators.Len() == 0 {
			return sdk.NewCheck(0, ""), ErrInvalidValidator()
		}
		for i, v := range validators {
			if v.OwnerAddress == txInner.Proposer.String() {
				break
			}
			if i+1 == len(validators) {
				return sdk.NewCheck(0, ""), ErrInvalidValidator()
			}
		}

		if txInner.ExpireBlockHeight != nil && ctx.BlockHeight() >= *txInner.ExpireBlockHeight {
			return sdk.NewCheck(0, ""), ErrInvalidExpireBlockHeight()
		}

		var fileurlJson map[string][]string

		if err = json.Unmarshal([]byte(txInner.Fileurl), &fileurlJson); err != nil {
			return sdk.NewCheck(0, ""), ErrInvalidFileurlJson()
		}

		if _, ok := fileurlJson[utils.GOOSDIST]; !ok {
			return sdk.NewCheck(0, ""), ErrNoFileurl()
		}

		var md5Json map[string]string

		if err = json.Unmarshal([]byte(txInner.Md5), &md5Json); err != nil {
			return sdk.NewCheck(0, ""), ErrInvalidMd5Json()
		}

		if _, ok := md5Json[utils.GOOSDIST]; !ok {
			return sdk.NewCheck(0, ""), ErrNoMd5()
		}

		// Transfer gasFee
		gasFee, err := checkGasFee(app_state, sender, utils.GetParams().UpgradeProgramProposalGas)
		if err != nil {
			return sdk.NewCheck(0, ""), err
		}
		app_state.SubBalance(sender, gasFee.Int)
	case TxVote:
		if !bytes.Equal(txInner.Voter.Bytes(), sender.Bytes()) {
			return sdk.NewCheck(0, ""), ErrMissingSignature()
		}

		validators := stake.GetCandidates().Validators()
		if validators == nil || validators.Len() == 0 {
			return sdk.NewCheck(0, ""), ErrInvalidValidator()
		}
		for i, v := range validators {
			if v.OwnerAddress == txInner.Voter.String() {
				break
			}
			if i+1 == len(validators) {
				return sdk.NewCheck(0, ""), ErrInvalidValidator()
			}
		}

		proposal := GetProposalById(txInner.ProposalId)
		if proposal == nil {
			return sdk.NewCheck(0, ""), ErrInvalidParameter()
		}

		if proposal.ExpireBlockHeight > 0 && ctx.BlockHeight() >= proposal.ExpireBlockHeight - 2 {
			return sdk.NewCheck(0, ""), ErrExpirationTooClose()
		}

		if proposal.ResultBlockHeight != 0 {
			if proposal.Result == "Approved" {
				return sdk.NewCheck(0, ""), ErrApprovedProposal()
			} else {
				return sdk.NewCheck(0, ""), ErrRejectedProposal()
			}
		}
	}

	return
}

// DeliverTx executes the tx if valid
func DeliverTx(ctx types.Context, store state.SimpleDB,
	tx sdk.Tx, hash []byte) (res sdk.DeliverResult, err error) {

	res.GasFee = big.NewInt(0)

	app_state := ctx.EthappState()

	switch txInner := tx.Unwrap().(type) {
	case TxTransferFundPropose:
		expireBlockHeight := ctx.BlockHeight() + int64(utils.GetParams().ProposalExpirePeriod)
		var expireTimestamp int64
		if txInner.ExpireTimestamp != nil {
			expireTimestamp = *txInner.ExpireTimestamp
			expireBlockHeight = 0
		} else if txInner.ExpireBlockHeight != nil {
			expireBlockHeight = *txInner.ExpireBlockHeight
		}
		hashJson, _ := json.Marshal(hash)
		pp := NewTransferFundProposal(
			string(hashJson[1:len(hashJson)-1]),
			txInner.Proposer,
			ctx.BlockHeight(),
			txInner.From,
			txInner.To,
			txInner.Amount,
			txInner.Reason,
			expireTimestamp,
			expireBlockHeight,
		)

		balance, err := commons.GetBalance(app_state, *txInner.From)
		if err != nil {
			return res, ErrInvalidParameter()
		}

		amount, _ := sdk.NewIntFromString(txInner.Amount)
		if balance.LT(amount) {
			return res, ErrInsufficientBalance()
		}

		SaveProposal(pp)
		commons.TransferWithReactor(*pp.Detail["from"].(*common.Address), utils.GovHoldAccount, amount, ProposalReactor{pp.Id, ctx.BlockHeight(), ""})

		// Check gasFee  -- start
		// get the sender
		sender, err := getTxSender(ctx)
		if err != nil {
			return res, err
		}
		params := utils.GetParams()
		gasUsed := params.TransferFundProposalGas

		if gasFee, err := checkGasFee(app_state, sender, gasUsed); err != nil {
			return res, err
		} else {
			res.GasFee = gasFee.Int
			res.GasUsed = int64(gasUsed)
			// transfer gasFee
			commons.Transfer(sender, utils.HoldAccount, gasFee)
		}
		// Check gasFee  -- end

		utils.PendingProposal.Add(pp.Id, pp.ExpireTimestamp, pp.ExpireBlockHeight)

		res.Data = hash

	case TxChangeParamPropose:
		expireBlockHeight := ctx.BlockHeight() + int64(utils.GetParams().ProposalExpirePeriod)
		var expireTimestamp int64
		if txInner.ExpireTimestamp != nil {
			expireTimestamp = *txInner.ExpireTimestamp
			expireBlockHeight = 0
		} else if txInner.ExpireBlockHeight != nil {
			expireBlockHeight = *txInner.ExpireBlockHeight
		}
		hashJson, _ := json.Marshal(hash)
		cp := NewChangeParamProposal(
			string(hashJson[1:len(hashJson)-1]),
			txInner.Proposer,
			ctx.BlockHeight(),
			txInner.Name,
			txInner.Value,
			txInner.Reason,
			expireTimestamp,
			expireBlockHeight,
		)
		SaveProposal(cp)

		// Check gasFee  -- start
		// get the sender
		sender, err := getTxSender(ctx)
		if err != nil {
			return res, err
		}
		params := utils.GetParams()
		gasUsed := params.ChangeParamsProposalGas

		if gasFee, err := checkGasFee(app_state, sender, gasUsed); err != nil {
			return res, err
		} else {
			res.GasFee = gasFee.Int
			res.GasUsed = int64(gasUsed)
			// transfer gasFee
			commons.Transfer(sender, utils.HoldAccount, gasFee)
		}
		// Check gasFee  -- end

		utils.PendingProposal.Add(cp.Id, cp.ExpireTimestamp, cp.ExpireBlockHeight)

		res.Data = hash

	case TxDeployLibEniPropose:
		expireBlockHeight := ctx.BlockHeight() + int64(utils.GetParams().ProposalExpirePeriod)
		var expireTimestamp int64
		if txInner.ExpireTimestamp != nil {
			expireTimestamp = *txInner.ExpireTimestamp
			expireBlockHeight = 0
		} else if txInner.ExpireBlockHeight != nil {
			expireBlockHeight = *txInner.ExpireBlockHeight
		}
		hashJson, _ := json.Marshal(hash)
		dp := NewDeployLibEniProposal(
			string(hashJson[1:len(hashJson)-1]),
			txInner.Proposer,
			ctx.BlockHeight(),
			txInner.Name,
			txInner.Version,
			txInner.Fileurl,
			txInner.Md5,
			txInner.Reason,
			"init",
			expireTimestamp,
			expireBlockHeight,
		)
		SaveProposal(dp)

		// Check gasFee  -- start
		// get the sender
		sender, err := getTxSender(ctx)
		if err != nil {
			return res, err
		}
		params := utils.GetParams()
		gasUsed := params.DeployLibEniProposalGas

		if gasFee, err := checkGasFee(app_state, sender, gasUsed); err != nil {
			return res, err
		} else {
			res.GasFee = gasFee.Int
			res.GasUsed = int64(gasUsed)
			// transfer gasFee
			commons.Transfer(sender, utils.HoldAccount, gasFee)
		}
		// Check gasFee  -- end

		utils.PendingProposal.Add(dp.Id, dp.ExpireTimestamp, dp.ExpireBlockHeight)

		res.Data = hash

		DownloadLibEni(dp)

	case TxRetireProgramPropose:
		expireBlockHeight := ctx.BlockHeight() + int64(utils.GetParams().ProposalExpirePeriod)
		if txInner.ExpireBlockHeight != nil {
			expireBlockHeight = *txInner.ExpireBlockHeight
		}
		hashJson, _ := json.Marshal(hash)
		cp := NewRetireProgramProposal(
			string(hashJson[1:len(hashJson)-1]),
			txInner.Proposer,
			ctx.BlockHeight(),
			version.Version,
			txInner.PreservedValidators,
			txInner.Reason,
			expireBlockHeight,
		)
		SaveProposal(cp)

		// Check gasFee  -- start
		// get the sender
		sender, err := getTxSender(ctx)
		if err != nil {
			return res, err
		}
		params := utils.GetParams()
		gasUsed := params.RetireProgramProposalGas

		if gasFee, err := checkGasFee(app_state, sender, gasUsed); err != nil {
			return res, err
		} else {
			res.GasFee = gasFee.Int
			res.GasUsed = int64(gasUsed)
			// transfer gasFee
			commons.Transfer(sender, utils.HoldAccount, gasFee)
		}
		// Check gasFee  -- end

		// check ahead one block
		utils.PendingProposal.Add(cp.Id, cp.ExpireTimestamp, cp.ExpireBlockHeight - 1)

		res.Data = hash
	case TxUpgradeProgramPropose:
		expireBlockHeight := ctx.BlockHeight() + int64(utils.GetParams().ProposalExpirePeriod)
		if txInner.ExpireBlockHeight != nil {
			expireBlockHeight = *txInner.ExpireBlockHeight
		}
		hashJson, _ := json.Marshal(hash)
		cp := NewUpgradeProgramProposal(
			string(hashJson[1:len(hashJson)-1]),
			txInner.Proposer,
			ctx.BlockHeight(),
			version.Version,
			txInner.Name,
			txInner.Version,
			txInner.Fileurl,
			txInner.Md5,
			txInner.Reason,
			expireBlockHeight,
		)
		SaveProposal(cp)

		// Check gasFee  -- start
		// get the sender
		sender, err := getTxSender(ctx)
		if err != nil {
			return res, err
		}
		params := utils.GetParams()
		gasUsed := params.UpgradeProgramProposalGas

		if gasFee, err := checkGasFee(app_state, sender, gasUsed); err != nil {
			return res, err
		} else {
			res.GasFee = gasFee.Int
			res.GasUsed = int64(gasUsed)
			// transfer gasFee
			commons.Transfer(sender, utils.HoldAccount, gasFee)
		}
		// Check gasFee  -- end

		utils.PendingProposal.Add(cp.Id, cp.ExpireTimestamp, cp.ExpireBlockHeight)
		res.Data = hash

		DownloadProgramCmd(cp)

	case TxVote:
		var vote *Vote
		if vote = GetVoteByPidAndVoter(txInner.ProposalId, txInner.Voter.String()); vote != nil {
			vote.Answer = txInner.Answer
			vote.BlockHeight = ctx.BlockHeight()
			UpdateVote(vote)
		} else {
			vote = NewVote(
				txInner.ProposalId,
				txInner.Voter,
				ctx.BlockHeight(),
				txInner.Answer,
			)
			SaveVote(vote)
		}

		proposal := GetProposalById(txInner.ProposalId)

		checkResult := CheckProposal(txInner.ProposalId, &txInner.Voter)

		switch proposal.Type {
		case TRANSFER_FUND_PROPOSAL:
			amount, _ := sdk.NewIntFromString(proposal.Detail["amount"].(string))
			switch checkResult {
			case "approved":
				// as succeeded proposal only need to add balance to receiver,
				// so the transfer should always be successful
				// but we still use the reactor to keep the compatible with the old strategy
				commons.TransferWithReactor(utils.GovHoldAccount, *proposal.Detail["to"].(*common.Address), amount, ProposalReactor{proposal.Id, ctx.BlockHeight(), "Approved"})
			case "rejected":
				// as succeeded proposal only need to refund balance to sender,
				// so the transfer should always be successful
				// but we still use the reactor to keep the compatible with the old strategy
				commons.TransferWithReactor(utils.GovHoldAccount, *proposal.Detail["from"].(*common.Address), amount, ProposalReactor{proposal.Id, ctx.BlockHeight(), "Rejected"})
			}
			if checkResult == "approved" || checkResult == "rejected" {
				utils.PendingProposal.Del(proposal.Id)
			}
		case CHANGE_PARAM_PROPOSAL:
			switch checkResult {
			case "approved":
				utils.SetParam(proposal.Detail["name"].(string), proposal.Detail["value"].(string))
				ProposalReactor{proposal.Id, ctx.BlockHeight(), "Approved"}.React("success", "")
			case "rejected":
				ProposalReactor{proposal.Id, ctx.BlockHeight(), "Rejected"}.React("success", "")
			}
			if checkResult == "approved" || checkResult == "rejected" {
				utils.PendingProposal.Del(proposal.Id)
			}
		case DEPLOY_LIBENI_PROPOSAL:
			switch checkResult {
			case "approved":
				ProposalReactor{proposal.Id, ctx.BlockHeight(), "Approved"}.React("success", "")
			case "rejected":
				if proposal.Detail["status"] != "ready" {
					CancelDownload(proposal, false)
				}
				utils.PendingProposal.Del(proposal.Id)
				ProposalReactor{proposal.Id, ctx.BlockHeight(), "Rejected"}.React("success", "")
			}
		case RETIRE_PROGRAM_PROPOSAL:
			switch checkResult {
			case "approved":
				ProposalReactor{proposal.Id, ctx.BlockHeight(), "Approved"}.React("success", "")
			case "rejected":
				utils.PendingProposal.Del(proposal.Id)
				ProposalReactor{proposal.Id, ctx.BlockHeight(), "Rejected"}.React("success", "")
			}
		case UPGRADE_PROGRAM_PROPOSAL:
			switch checkResult {
			case "approved":
				ProposalReactor{proposal.Id, ctx.BlockHeight(), "Approved"}.React("success", "")
			case "rejected":
				utils.PendingProposal.Del(proposal.Id)
				ProposalReactor{proposal.Id, ctx.BlockHeight(), "Rejected"}.React("success", "")
			}
		}
	}

	return
}

func CheckProposal(pid string, voter *common.Address) string {
	votes := GetVotesByPid(pid)
	validators := stake.GetCandidates().Validators()

	if validators == nil || validators.Len() == 0 {
		return "no validator"
	}

	approvedPower := big.NewInt(0)
	rejectedPower := big.NewInt(0)
	allPower := big.NewInt(0)
	voterPower := big.NewInt(0)
	for _, va := range validators {
		for _, vo := range votes {
			// should check voter is still valid validator first
			if vo.Voter.String() == va.OwnerAddress {
				if strings.Compare(vo.Answer, "Y") == 0 {
					approvedPower.Add(approvedPower, big.NewInt(va.VotingPower))
				}
				if strings.Compare(vo.Answer, "N") == 0 {
					rejectedPower.Add(rejectedPower, big.NewInt(va.VotingPower))
				}
			}
		}
		if voter != nil && voter.String() == va.OwnerAddress {
			voterPower = big.NewInt(va.VotingPower)
		}
		allPower.Add(allPower, big.NewInt(va.VotingPower))
	}

	allPower.Mul(allPower, big.NewInt(2))
	three := big.NewInt(3)
	voterPower.Mul(voterPower, three)
	approvedPower.Mul(approvedPower, three)
	rejectedPower.Mul(rejectedPower, three)

	if approvedPower.Cmp(allPower) >= 0 {
		// To avoid repeated commit, let's recheck with count of voters - voter
		if voter == nil || approvedPower.Sub(approvedPower, voterPower).Cmp(allPower) < 0 {
			return "approved"
		}
	} else if rejectedPower.Cmp(allPower) >= 0 {
		// To avoid repeated commit, let's recheck with count of voters - voter
		if voter == nil || rejectedPower.Sub(rejectedPower, voterPower).Cmp(allPower) < 0 {
			return "rejected"
		}
	}
	return "not determined"
}

type ProposalReactor struct {
	ProposalId  string
	BlockHeight int64
	Result      string
}

func (pr ProposalReactor) React(result, msg string) {
	if result == "success" {
		// If the default result is not set, then do nothing
		if pr.Result == "" {
			return
		}
		result = pr.Result
	}
	UpdateProposalResult(pr.ProposalId, result, msg, pr.BlockHeight)
}

// get the sender from the ctx and ensure it matches the tx pubkey
func getTxSender(ctx types.Context) (sender common.Address, err error) {
	senders := ctx.GetSigners()
	if len(senders) != 1 {
		return sender, ErrMissingSignature()
	}
	return senders[0], nil
}

func checkGasFee(state *ethState.StateDB, address common.Address, gas uint64) (sdk.Int, error) {
	balance, err := commons.GetBalance(state, address)
	if err != nil {
		return sdk.Int{}, ErrInvalidParameter()
	}

	gasFee := utils.CalGasFee(gas, utils.GetParams().GasPrice)

	if balance.LT(gasFee) {
		return sdk.Int{}, ErrInsufficientBalance()
	}

	return gasFee, nil
}

func getOTAInfo(p *Proposal) *eni.OTAInfo {
	var fileurlJson map[string][]string
	if err := json.Unmarshal([]byte(p.Detail["fileurl"].(string)), &fileurlJson); err != nil {
		return nil
	}

	var md5Json map[string]string
	if err := json.Unmarshal([]byte(p.Detail["md5"].(string)), &md5Json); err != nil {
		return nil
	}

	fileurl, ok := fileurlJson[utils.GOOSDIST]
	if !ok {
		return nil
	}

	md5, ok := md5Json[utils.GOOSDIST]
	if !ok {
		return nil
	}

	return &eni.OTAInfo{
		p.Detail["name"].(string),
		p.Detail["version"].(string),
		fileurl,
		md5,
	}
}

func DownloadLibEni(p *Proposal) {
	oi := getOTAInfo(p)
	if oi == nil {
		return
	}

	result := make(chan bool)

	go func() {
		if r := <-result; r {
			if r, ok := cancelDownload[p.Id]; ok {
				delete(cancelDownload, p.Id)
				if r {
					RegisterLibEni(p)
					UpdateDeployLibEniStatus(p.Id, "deployed")
				} else {
					UpdateDeployLibEniStatus(p.Id, "ready")
				}
			} else {
				UpdateDeployLibEniStatus(p.Id, "ready")
			}
		} else {
			if r, ok := cancelDownload[p.Id]; ok {
				delete(cancelDownload, p.Id)
				if r {
					UpdateDeployLibEniStatus(p.Id, "collapsed") // failed, but proposal has been approved
				} else {
					UpdateDeployLibEniStatus(p.Id, "failed")
				}
			} else {
				UpdateDeployLibEniStatus(p.Id, "failed")
			}
		}
	}()

	go func() {
		for {
			if _, ok := cancelDownload[p.Id]; ok {
				result <- false
				break
			}
			if err := OTAInstance.DownloadInfo(*oi); err == nil {
				result <- true
				break
			}
			if _, ok := cancelDownload[p.Id]; ok {
				result <- false
				break
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

func CancelDownload(p *Proposal, bpanic bool) {
	cancelDownload[p.Id] = bpanic
}

func RegisterLibEni(p *Proposal) {
	oi := getOTAInfo(p)
	if oi == nil {
		return
	}
	OTAInstance.Register(*oi)
}

func DestroyLibEni(p *Proposal) {
	oi := getOTAInfo(p)
	if oi == nil {
		return
	}
	OTAInstance.Destroy(*oi)
}

// KilProgramCmd kill the process from internal
func KillProgramCmd(p *Proposal) error {
	info := &types.CmdInfo{}
	reply := &types.MonitorResponse{}
	err := callRpc("Monitor.Kill", info, reply)
	if err != nil {
		//log.Fatal("call monitor rpc error:", err)
		return err
	}
	return nil
}

// DownloadProgramCmd download new program version
func DownloadProgramCmd(p *Proposal) error {
	oi := getOTAInfo(p)
	if oi == nil {
		return errors.New("unknown error")
	}
	info := &types.CmdInfo{Name: oi.LibName, Version: oi.Version, DownloadURLs:oi.Url, MD5: oi.Checksum}
	reply := &types.MonitorResponse{}
	err := callRpc("Monitor.Download", info, reply)
	if err != nil {
		//log.Fatal("call monitor rpc error:", err)
		return err
	}
	return nil
}

// UpgradeProgramCmd upgrade new program version
func UpgradeProgramCmd(p *Proposal) error {
	oi := getOTAInfo(p)
	if oi == nil {
		return errors.New("unknown error")
	}
	info := &types.CmdInfo{Name: oi.LibName, Version: oi.Version, DownloadURLs:oi.Url, MD5: oi.Checksum}
	reply := &types.MonitorResponse{}
	err := callRpc("Monitor.Upgrade", info, reply)
	if err != nil {
		return err
	}
	return nil
}

func callRpc(serviceMethod string, info *types.CmdInfo, reply *types.MonitorResponse) error {
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:26650")
	if err != nil {
		return err
	}
	err = client.Call(serviceMethod, info, reply)
	if err != nil {
		return err
	}
	return nil
}
