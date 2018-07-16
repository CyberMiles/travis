package stake

import (
	"fmt"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/errors"
	"github.com/CyberMiles/travis/sdk/state"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	ethstat "github.com/ethereum/go-ethereum/core/state"
	"math"
	"math/big"
	"strconv"
)

// nolint
const (
	stakingModuleName = "stake"
	foundationAddress = "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc" // fixme move to config file
)

//var (
//	minStakedAmount, _ = new(big.Int).SetString("1000000000000000000000", 10)   // 1000: the minimum amount of CMTs a single CMT Cube device can hold and stake
//	maxStakedAmount, _ = new(big.Int).SetString("100000000000000000000000", 10) // 100,000: the maximum amount of CMTs a single CMT Cube device can hold and stake
//)

//_______________________________________________________________________

// DelegatedProofOfStake - interface to enforce delegation stake
type delegatedProofOfStake interface {
	declareCandidacy(TxDeclareCandidacy, *big.Int) error
	updateCandidacy(TxUpdateCandidacy, *big.Int) error
	withdrawCandidacy(TxWithdrawCandidacy) error
	verifyCandidacy(TxVerifyCandidacy) error
	activateCandidacy(TxActivateCandidacy) error
	delegate(TxDelegate) error
	withdraw(TxWithdraw) error
}

//_______________________________________________________________________

// InitState - set genesis parameters for staking
func InitState(key string, value interface{}, store state.SimpleDB) error {
	return nil
}

func SetValidator(val types.GenesisValidator, store state.SimpleDB) error {
	if val.Address == "0000000000000000000000000000000000000000" {
		return ErrBadValidatorAddr()
	}

	addr := common.HexToAddress(val.Address)

	// create and save the empty candidate
	bond := GetCandidateByAddress(addr)
	if bond != nil {
		return ErrCandidateExistsAddr()
	}

	params := utils.GetParams()
	deliverer := deliver{
		store:  store,
		sender: addr,
		params: params,
	}

	desc := Description{
		Name:     val.Name,
		Website:  val.Website,
		Location: val.Location,
		Email:    val.Email,
		Profile:  val.Profile,
	}

	tx := TxDeclareCandidacy{types.PubKeyString(val.PubKey), utils.ToWei(val.MaxAmount).String(), val.CompRate, desc}
	power, _ := strconv.ParseInt(val.Power, 10, 64)
	return deliverer.declareGenesisCandidacy(tx, power)
}

// CheckTx checks if the tx is properly structured
func CheckTx(ctx types.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.CheckResult, err error) {
	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	// get the sender
	sender, err := getTxSender(ctx)
	if err != nil {
		return res, err
	}

	params := utils.GetParams()
	checker := check{
		store:  store,
		sender: sender,
		params: params,
		state:  ctx.EthappState(),
		nonce:  ctx.GetNonce(),
	}

	switch txInner := tx.Unwrap().(type) {
	case TxDeclareCandidacy:
		gasFee := utils.CalGasFee(params.DeclareCandidacy, params.GasPrice)
		return res, checker.declareCandidacy(txInner, gasFee)
	case TxUpdateCandidacy:
		gasFee := utils.CalGasFee(params.UpdateCandidacy, params.GasPrice)
		return res, checker.updateCandidacy(txInner, gasFee)
	case TxWithdrawCandidacy:
		return res, checker.withdrawCandidacy(txInner)
	case TxVerifyCandidacy:
		return res, checker.verifyCandidacy(txInner)
	case TxActivateCandidacy:
		return res, checker.activateCandidacy(txInner)
	case TxDelegate:
		return res, checker.delegate(txInner)
	case TxWithdraw:
		return res, checker.withdraw(txInner)
	}

	utils.TravisTxAddrs = append(utils.TravisTxAddrs, &sender)
	return res, errors.ErrUnknownTxType(tx)
}

// DeliverTx executes the tx if valid
func DeliverTx(ctx types.Context, store state.SimpleDB, tx sdk.Tx, hash []byte) (res sdk.DeliverResult, err error) {
	_, err = CheckTx(ctx, store, tx)
	if err != nil {
		return
	}

	sender, err := getTxSender(ctx)
	if err != nil {
		return
	}

	params := utils.GetParams()
	deliverer := deliver{
		store:  store,
		sender: sender,
		params: params,
		state:  ctx.EthappState(),
		height: ctx.BlockHeight(),
	}
	res.GasFee = big.NewInt(0)
	// Run the transaction
	switch _tx := tx.Unwrap().(type) {
	case TxDeclareCandidacy:
		gasFee := utils.CalGasFee(utils.GetParams().DeclareCandidacy, utils.GetParams().GasPrice)
		err := deliverer.declareCandidacy(_tx, gasFee)
		if err == nil {
			res.GasUsed = int64(utils.GetParams().DeclareCandidacy)
			res.GasFee = gasFee
		}
		return res, err
	case TxUpdateCandidacy:
		gasFee := utils.CalGasFee(utils.GetParams().UpdateCandidacy, utils.GetParams().GasPrice)
		err := deliverer.updateCandidacy(_tx, gasFee)
		if err == nil {
			res.GasUsed = int64(utils.GetParams().UpdateCandidacy)
			res.GasFee = gasFee
		}
		return res, err
	case TxWithdrawCandidacy:
		return res, deliverer.withdrawCandidacy(_tx)
	case TxVerifyCandidacy:
		return res, deliverer.verifyCandidacy(_tx)
	case TxActivateCandidacy:
		return res, deliverer.activateCandidacy(_tx)
	case TxDelegate:
		return res, deliverer.delegate(_tx)
	case TxWithdraw:
		return res, deliverer.withdraw(_tx)
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

//_______________________________________________________________________

type check struct {
	store  state.SimpleDB
	sender common.Address
	params *utils.Params
	state  *ethstat.StateDB
	nonce  uint64
}

var _ delegatedProofOfStake = check{} // enforce interface at compile time

func (c check) declareCandidacy(tx TxDeclareCandidacy, gasFee *big.Int) error {
	// check to see if the pubkey or address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate != nil {
		return fmt.Errorf("address has been declared")
	}

	candidate = GetCandidateByPubKey(tx.PubKey)
	if candidate != nil {
		return fmt.Errorf("pubkey has been declared")
	}

	// check to see if the associated account has 10%(ssr, short for self-staking ratio, configurable) of the max staked CMT amount
	maxAmount, ok := new(big.Int).SetString(tx.MaxAmount, 10)
	if !ok || maxAmount.Cmp(big.NewInt(0)) < 0 {
		return ErrBadAmount()
	}

	rr := tx.SelfStakingAmount(c.params.SelfStakingRatio)

	totalCost := new(big.Int).Add(rr, gasFee)

	// check if the delegator has sufficient funds
	err := checkBalance(c.state, c.sender, totalCost)
	if err != nil {
		return err
	}

	return nil
}

func (c check) updateCandidacy(tx TxUpdateCandidacy, gasFee *big.Int) error {
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return fmt.Errorf("cannot edit non-exsits candidacy")
	}

	// If the max amount of CMTs is updated, the 10% of self-staking will be re-computed,
	// and the different will be charged
	if tx.MaxAmount != "" {
		maxAmount, ok := new(big.Int).SetString(tx.MaxAmount, 10)
		if !ok || maxAmount.Cmp(big.NewInt(0)) < 0 {
			return ErrBadAmount()
		}

		if maxAmount.Cmp(candidate.ParseMaxShares()) > 0 {
			rechargeAmount := getRechargeAmount(maxAmount, candidate, c.params.SelfStakingRatio)

			totalCost := new(big.Int).Add(rechargeAmount, gasFee)

			// check if the delegator has sufficient funds
			err := checkBalance(c.state, c.sender, totalCost)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c check) withdrawCandidacy(tx TxWithdrawCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return fmt.Errorf("cannot withdraw non-exsits candidacy")
	}

	return nil
}

func (c check) verifyCandidacy(tx TxVerifyCandidacy) error {
	// check to see if the candidate address to be verified has been registered before
	candidate := GetCandidateByAddress(tx.CandidateAddress)
	if candidate == nil {
		return fmt.Errorf("cannot verify non-exsits candidacy")
	}

	// check to see if the request was initiated by a special account
	if c.sender != common.HexToAddress(foundationAddress) {
		return ErrVerificationDisallowed()
	}

	return nil
}

func (c check) activateCandidacy(tx TxActivateCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return fmt.Errorf("cannot activate non-exsits candidacy")
	}

	if candidate.Active == "Y" {
		return fmt.Errorf("already activated")
	}

	return nil
}

func (c check) delegate(tx TxDelegate) error {
	err := VerifyCubeSignature(c.sender, c.nonce, tx.CubeBatch, tx.Sig)
	if err != nil {
		return err
	}

	candidate := GetCandidateByAddress(tx.ValidatorAddress)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	// check if the delegator has sufficient funds
	amount, ok := new(big.Int).SetString(tx.Amount, 10)
	if !ok || amount.Cmp(big.NewInt(0)) < 0 {
		return ErrBadAmount()
	}

	err = checkBalance(c.state, c.sender, amount)
	if err != nil {
		return err
	}

	// check to see if the validator has reached its declared max amount CMTs to be staked.
	x := new(big.Int)
	x.Add(candidate.ParseShares(), amount)
	if x.Cmp(candidate.ParseMaxShares()) > 0 {
		return ErrReachMaxAmount()
	}

	return nil
}

func (c check) withdraw(tx TxWithdraw) error {
	// check if has delegated
	candidate := GetCandidateByAddress(tx.ValidatorAddress)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	amount, ok := new(big.Int).SetString(tx.Amount, 10)
	if !ok || amount.Cmp(big.NewInt(0)) < 0 {
		return ErrBadAmount()
	}

	d := GetDelegation(c.sender, candidate.PubKey)
	if d == nil {
		return ErrDelegationNotExists()
	}

	if amount.Cmp(d.Shares()) > 0 {
		return ErrInvalidWithdrawalAmount()
	}

	return nil
}

//_____________________________________________________________________

type deliver struct {
	store  state.SimpleDB
	sender common.Address
	params *utils.Params
	state  *ethstat.StateDB
	height int64
}

var _ delegatedProofOfStake = deliver{} // enforce interface at compile time

// These functions assume everything has been authenticated,
// now we just perform action and save
func (d deliver) declareCandidacy(tx TxDeclareCandidacy, gasFee *big.Int) error {
	// create and save the empty candidate
	pubKey, err := types.GetPubKey(tx.PubKey)
	if err != nil {
		return err
	}

	now := utils.GetNow()
	candidate := &Candidate{
		PubKey:       pubKey,
		OwnerAddress: d.sender.String(),
		Shares:       "0",
		VotingPower:  0,
		MaxShares:    tx.MaxAmount,
		CompRate:     tx.CompRate,
		CreatedAt:    now,
		UpdatedAt:    now,
		Description:  tx.Description,
		Verified:     "N",
		Active:       "Y",
		BlockHeight:  d.height,
	}
	SaveCandidate(candidate)

	// delegate a part of the max staked CMT amount
	amount := tx.SelfStakingAmount(d.params.SelfStakingRatio)
	totalCost := big.NewInt(0).Add(amount, gasFee)
	// check if the delegator has sufficient funds
	if err := checkBalance(d.state, d.sender, totalCost); err != nil {
		return err
	} else {
		commons.Transfer(d.sender, utils.HoldAccount, gasFee)
	}

	txDelegate := TxDelegate{ValidatorAddress: d.sender, Amount: amount.String()}
	return d.delegate(txDelegate)
}

func (d deliver) declareGenesisCandidacy(tx TxDeclareCandidacy, votingPower int64) error {
	// create and save the empty candidate
	pubKey, err := types.GetPubKey(tx.PubKey)
	if err != nil {
		return err
	}

	rp := int64(math.Sqrt(float64(votingPower * 10000)))
	now := utils.GetNow()
	candidate := &Candidate{
		PubKey:       pubKey,
		OwnerAddress: d.sender.String(),
		Shares:       "0",
		VotingPower:  votingPower,
		RankingPower: rp,
		MaxShares:    tx.MaxAmount,
		CompRate:     tx.CompRate,
		CreatedAt:    now,
		UpdatedAt:    now,
		Description:  tx.Description,
		Verified:     "N",
		Active:       "Y",
		BlockHeight:  1,
	}
	SaveCandidate(candidate)

	// delegate a part of the max staked CMT amount
	amount := new(big.Int).Mul(big.NewInt(votingPower), big.NewInt(1e18))
	txDelegate := TxDelegate{ValidatorAddress: d.sender, Amount: amount.String()}
	return d.delegate(txDelegate)
}

func (d deliver) updateCandidacy(tx TxUpdateCandidacy, gasFee *big.Int) error {
	// create and save the empty candidate
	candidate := GetCandidateByAddress(d.sender)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	var rechargeAmount = big.NewInt(0)
	// If the max amount of CMTs is updated, the 10% of self-staking will be re-computed,
	// and the different will be charged
	if tx.MaxAmount != "" {
		maxAmount := utils.ParseInt(tx.MaxAmount)

		if candidate.ParseMaxShares().Cmp(maxAmount) != 0 {
			rechargeAmount = getRechargeAmount(maxAmount, candidate, d.params.SelfStakingRatio)
			rechargeAmountAbs := new(big.Int)
			rechargeAmountAbs.Abs(rechargeAmount)

			if rechargeAmount.Cmp(big.NewInt(0)) > 0 {
				// charge
				commons.Transfer(d.sender, utils.HoldAccount, rechargeAmountAbs)
				candidate.AddShares(rechargeAmount)

				// update delegation
				delegation := GetDelegation(d.sender, candidate.PubKey)
				delegation.AddDelegateAmount(rechargeAmount)
				delegation.UpdatedAt = utils.GetNow()
				UpdateDelegation(delegation)
			}

			candidate.MaxShares = maxAmount.String()
		}
	}

	// If other information was updated, set the verified status to false
	if candidate.Description != tx.Description {
		candidate.Verified = "N"
		candidate.Description = tx.Description
	}

	totalCost := big.NewInt(0).Add(rechargeAmount, gasFee)
	// check if the delegator has sufficient funds
	if err := checkBalance(d.state, d.sender, totalCost); err != nil {
		return err
	} else {
		commons.Transfer(d.sender, utils.HoldAccount, gasFee)
	}

	candidate.UpdatedAt = utils.GetNow()
	updateCandidate(candidate)
	return nil
}

func (d deliver) withdrawCandidacy(tx TxWithdrawCandidacy) error {
	// create and save the empty candidate
	validatorAddress := d.sender
	candidate := GetCandidateByAddress(validatorAddress)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	// All staked tokens will be distributed back to delegator addresses.
	// Self-staked CMTs will be refunded back to the validator address.
	delegations := GetDelegationsByPubKey(candidate.PubKey)
	for _, delegation := range delegations {
		txWithdraw := TxWithdraw{ValidatorAddress: validatorAddress, Amount: delegation.Shares().String()}
		d.doWithdraw(delegation, delegation.Shares(), candidate, txWithdraw)
	}

	//removeCandidate(candidate)
	candidate.Shares = "0"
	candidate.UpdatedAt = utils.GetNow()
	updateCandidate(candidate)
	return nil
}

func (d deliver) verifyCandidacy(tx TxVerifyCandidacy) error {
	// verify candidacy
	candidate := GetCandidateByAddress(tx.CandidateAddress)
	if tx.Verified {
		candidate.Verified = "Y"
	} else {
		candidate.Verified = "N"
	}
	candidate.UpdatedAt = utils.GetNow()
	updateCandidate(candidate)
	return nil
}

func (d deliver) activateCandidacy(tx TxActivateCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(d.sender)
	if candidate == nil {
		return fmt.Errorf("cannot activate non-exsits candidacy")
	}

	candidate.Active = "Y"
	candidate.UpdatedAt = utils.GetNow()
	updateCandidate(candidate)
	return nil
}

func (d deliver) delegate(tx TxDelegate) error {
	// Get the pubKey bond account
	candidate := GetCandidateByAddress(tx.ValidatorAddress)

	delegateAmount, ok := new(big.Int).SetString(tx.Amount, 0)
	if !ok || delegateAmount.Cmp(big.NewInt(0)) < 0 {
		return ErrBadAmount()
	}

	// Move coins from the delegator account to the pubKey lock account
	err := commons.Transfer(d.sender, d.params.HoldAccount, delegateAmount)
	if err != nil {
		return err
	}

	// create or update delegation
	now := utils.GetNow()
	delegation := GetDelegation(d.sender, candidate.PubKey)
	if delegation == nil {
		delegation = &Delegation{
			DelegatorAddress: d.sender,
			PubKey:           candidate.PubKey,
			DelegateAmount:   tx.Amount,
			AwardAmount:      "0",
			WithdrawAmount:   "0",
			SlashAmount:      "0",
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		SaveDelegation(delegation)
	} else {
		delegation.AddDelegateAmount(delegateAmount)
		delegation.UpdatedAt = now
		UpdateDelegation(delegation)
	}

	// Add delegateAmount to candidate
	candidate.AddShares(delegateAmount)
	delegateHistory := &DelegateHistory{0, d.sender, candidate.PubKey, delegateAmount, "delegate", now}
	updateCandidate(candidate)
	saveDelegateHistory(delegateHistory)
	return nil
}

func (d deliver) withdraw(tx TxWithdraw) error {
	// get pubKey candidate
	candidate := GetCandidateByAddress(tx.ValidatorAddress)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	amount, ok := new(big.Int).SetString(tx.Amount, 10)
	if !ok {
		return ErrInvalidWithdrawalAmount()
	}

	delegation := GetDelegation(d.sender, candidate.PubKey)

	// candidates can't withdraw the reserved reservation fund
	if d.sender.String() == candidate.OwnerAddress {
		remained := new(big.Int)
		remained.Sub(delegation.Shares(), amount)
		if remained.Cmp(candidate.SelfStakingAmount(d.params.SelfStakingRatio)) < 0 {
			return ErrCandidateWithdrawalDisallowed()
		}
	}

	d.doWithdraw(delegation, amount, candidate, tx)

	// deduct shares from the candidate
	neg := new(big.Int).Neg(amount)
	candidate.AddShares(neg)
	now := utils.GetNow()
	candidate.UpdatedAt = now
	updateCandidate(candidate)

	delegateHistory := &DelegateHistory{0, d.sender, candidate.PubKey, amount, "withdraw", now}
	saveDelegateHistory(delegateHistory)

	return nil
}

func (d deliver) doWithdraw(delegation *Delegation, amount *big.Int, candidate *Candidate, tx TxWithdraw) {
	// update delegation withdraw amount
	delegation.AddWithdrawAmount(amount)
	UpdateDelegation(delegation)
	now := utils.GetNow()

	// record unstake requests, waiting 7 days
	performedBlockHeight := d.height + int64(utils.GetParams().UnstakeWaitingPeriod)
	// just for test
	//performedBlockHeight := d.height + 4
	unstakeRequest := &UnstakeRequest{"", delegation.DelegatorAddress, candidate.PubKey, d.height, performedBlockHeight, amount.String(), "PENDING", now, now}
	unstakeRequest.Id = common.Bytes2Hex(unstakeRequest.GenId())
	saveUnstakeRequest(unstakeRequest)

	return
}

func HandlePendingUnstakeRequests(height int64, store state.SimpleDB) error {
	params := utils.GetParams()
	reqs := GetUnstakeRequests(height)
	for _, req := range reqs {
		// get pubKey candidate
		candidate := GetCandidateByPubKey(types.PubKeyString(req.PubKey))
		if candidate == nil {
			continue
		}

		if candidate.Shares == "0" {
			//candidate.State = "N"
			removeCandidate(candidate)
		}

		delegation := GetDelegation(req.DelegatorAddress, candidate.PubKey)
		if delegation == nil {
			continue
		}

		if delegation.Shares().Cmp(big.NewInt(0)) == 0 {
			RemoveDelegation(delegation)
		}

		req.State = "COMPLETED"
		req.UpdatedAt = utils.GetNow()
		updateUnstakeRequest(req)

		// transfer coins back to account
		amount, _ := new(big.Int).SetString(req.Amount, 10)
		commons.Transfer(params.HoldAccount, req.DelegatorAddress, amount)
	}

	return nil
}

func checkBalance(state *ethstat.StateDB, addr common.Address, amount *big.Int) error {
	balance, err := commons.GetBalance(state, addr)
	if err != nil {
		return err
	}

	if balance.Cmp(amount) < 0 {
		return ErrInsufficientFunds()
	}

	return nil
}

func getRechargeAmount(maxAmount *big.Int, candidate *Candidate, ratio string) (needRechargeAmount *big.Int) {
	needRechargeAmount = new(big.Int)
	diff := new(big.Int).Sub(maxAmount, candidate.ParseMaxShares())
	x := new(big.Float).SetInt(diff)
	z := new(big.Float)
	r, _ := new(big.Float).SetString(ratio)
	z.Mul(x, r)
	z.Int(needRechargeAmount)

	delegation := GetDelegation(common.HexToAddress(candidate.OwnerAddress), candidate.PubKey)
	needRechargeAmount.Sub(needRechargeAmount, delegation.ParseAwardAmount())
	return
}
