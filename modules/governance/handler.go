package governance

import (
	"bytes"
	"encoding/json"
	"math/big"
	"strings"
	"time"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/state"
	"github.com/ethereum/go-ethereum/common"
	ethState "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm/eni"
)

// nolint
const governanceModuleName = "governance"

var OTAInstance = eni.NewOTAInstance()

var downloading = make(map[string] chan<- bool)

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
		if !bytes.Equal(txInner.Proposer.Bytes(), sender.Bytes()) {
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
			if i + 1 == len(validators) {
				return sdk.NewCheck(0, ""), ErrInvalidValidator()
			}
		}

		balance, err := commons.GetBalance(app_state, *txInner.From)
		if err != nil {
			return sdk.NewCheck(0, ""), ErrInvalidParameter()
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

		amount := big.NewInt(0)
		amount.SetString(txInner.Amount, 10)
		if balance.Cmp(amount) < 0 {
			return sdk.NewCheck(0, ""), ErrInsufficientBalance()
		}

		// Transfer gasFee
		if _, err := checkGasFee(app_state, sender, utils.GetParams().TransferFundProposal); err != nil {
			return sdk.NewCheck(0, ""), err
		}

		utils.TravisTxAddrs = append(utils.TravisTxAddrs, txInner.From)
	case TxChangeParamPropose:
		if !bytes.Equal(txInner.Proposer.Bytes(), sender.Bytes()) {
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
			if i + 1 == len(validators) {
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

		// Transfer gasFee
		if _, err := checkGasFee(app_state, sender, utils.GetParams().ChangeParamsProposal); err != nil {
			return sdk.NewCheck(0, ""), err
		}

		if ! utils.CheckParamType(txInner.Name, txInner.Value) {
			return sdk.NewCheck(0, ""), ErrInvalidParameter()
		}
	case TxDeployLibEniPropose:
		if !bytes.Equal(txInner.Proposer.Bytes(), sender.Bytes()) {
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
			if i + 1 == len(validators) {
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

		otaInfo := eni.OTAInfo {
			LibName: txInner.Name,
			Version: txInner.Version,
		}
		if valid, _ := OTAInstance.IsValidNewLib(otaInfo); !valid {
			return sdk.NewCheck(0, ""), ErrInvalidNewLib()
		}

		if HasUndeployedProposal(txInner.Name) {
			return sdk.NewCheck(0, ""), ErrOngoingLibFound()
		}

		// Transfer gasFee
		if _, err := checkGasFee(app_state, sender, utils.GetParams().DeployLibEniProposal); err != nil {
			return sdk.NewCheck(0, ""), err
		}

		var fileurlJson map[string]string

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
			if i + 1 == len(validators) {
				return sdk.NewCheck(0, ""), ErrInvalidValidator()
			}
		}

		proposal := GetProposalById(txInner.ProposalId)
		if proposal == nil {
			return sdk.NewCheck(0, ""), ErrInvalidParameter()
		}
		if proposal.ResultBlockHeight != 0 {
			if proposal.Result == "Approved" {
				return sdk.NewCheck(0, ""), ErrApprovedProposal()
			} else {
				return sdk.NewCheck(0, ""), ErrRejectedProposal()
			}
		}
		if proposal.Type == TRANSFER_FUND_PROPOSAL {
			utils.TravisTxAddrs = append(utils.TravisTxAddrs, proposal.Detail["to"].(*common.Address))
		}
	}

	return
}

// DeliverTx executes the tx if valid
func DeliverTx(ctx types.Context, store state.SimpleDB,
	tx sdk.Tx, hash []byte) (res sdk.DeliverResult, err error) {

	res.GasFee = big.NewInt(0)

	_, err = CheckTx(ctx, store, tx)
	if err != nil {
		return
	}

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
		hashJson, _ :=	json.Marshal(hash)
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

		amount := big.NewInt(0)
		amount.SetString(txInner.Amount, 10)
		if balance.Cmp(amount) < 0 {
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
		gasUsed := params.TransferFundProposal

		if gasFee, err := checkGasFee(app_state, sender, gasUsed); err != nil {
			return res, err
		} else {
			res.GasFee = gasFee
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
		gasUsed := params.ChangeParamsProposal

		if gasFee, err := checkGasFee(app_state, sender, gasUsed); err != nil {
			return res, err
		} else {
			res.GasFee = gasFee
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
		gasUsed := params.DeployLibEniProposal

		if gasFee, err := checkGasFee(app_state, sender, gasUsed); err != nil {
			return res, err
		} else {
			res.GasFee = gasFee
			res.GasUsed = int64(gasUsed)
			// transfer gasFee
			commons.Transfer(sender, utils.HoldAccount, gasFee)
		}
		// Check gasFee  -- end

		utils.PendingProposal.Add(dp.Id, dp.ExpireTimestamp, dp.ExpireBlockHeight)

		res.Data = hash

		result := make(chan bool)

		go func() {
			if r := <- result; r {
				UpdateDeployLibEniStatus(dp.Id, "ready")
			} else {
				UpdateDeployLibEniStatus(dp.Id, "unready")
			}
		}()

		go DownloadLibEni(dp, result, true)

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
			amount := new(big.Int)
			amount.SetString(proposal.Detail["amount"].(string), 10)
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
				CancelDownload(proposal)
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

func checkGasFee(state *ethState.StateDB, address common.Address, gas uint64) (*big.Int,  error) {
	balance, err := commons.GetBalance(state, address)
	if err != nil {
		return nil, ErrInvalidParameter()
	}

	gasFee := utils.CalGasFee(gas, utils.GetParams().GasPrice)

	if balance.Cmp(gasFee) < 0 {
		return nil, ErrInsufficientBalance()
	}

	return gasFee, nil
}

func getOTAInfo(p *Proposal) *eni.OTAInfo {
	var fileurlJson map[string]string
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

	return &eni.OTAInfo {
		p.Detail["name"].(string),
		p.Detail["version"].(string),
		fileurl,
		md5,
	}
}

func DownloadLibEni(p *Proposal, result chan<- bool, retry bool) {
	oi := getOTAInfo(p)
	if oi == nil {
		result <- false
		return
	}

	if _, ok := downloading[oi.LibName]; !ok {
		if retry {
			downloading[oi.LibName] = nil
		} else  {
			downloading[oi.LibName] = result
		}
		for {
			if err := OTAInstance.Download(*oi); err == nil {
				if err = OTAInstance.Verify(*oi); err == nil {
					if r := downloading[oi.LibName]; r != nil {
						r <- true
					}
					if retry {
						result <- true
					}
					break
				}
			}
			if r := downloading[oi.LibName]; r != nil {
				if retry {
					result <- false
				}
				r <- false
				break;
			}
			time.Sleep(10 * time.Minute)
		}
		delete(downloading, oi.LibName)
	} else {
		downloading[oi.LibName] = result
	}
}

func CancelDownload(p *Proposal) {
	oi := getOTAInfo(p)
	if oi == nil {
		return
	}

	if _, ok := downloading[oi.LibName]; ok {
		t := make(chan bool)
		go func() {
			<- t
			DestroyLibEni(p)
		}()
		downloading[oi.LibName] = t
	} else {
		DestroyLibEni(p)
	}
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
