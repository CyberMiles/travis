package stake

import (
	"fmt"
	"strconv"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/errors"
	"github.com/CyberMiles/travis/sdk/state"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	ethstat "github.com/ethereum/go-ethereum/core/state"
	"math/big"
)

// nolint
const (
	stakingModuleName = "stake"
	foundationAddress = "0x7eff122b94897ea5b0e2a9abf47b86337fafebdc" // fixme move to config file
)

var (
	minStakedAmount, _ = new(big.Int).SetString("1000000000000000000000", 10)   // 1000: the minimum amount of CMTs a single CMT Cube device can hold and stake
	maxStakedAmount, _ = new(big.Int).SetString("100000000000000000000000", 10) // 100,000: the maximum amount of CMTs a single CMT Cube device can hold and stake
)

//_______________________________________________________________________

// DelegatedProofOfStake - interface to enforce delegation stake
type delegatedProofOfStake interface {
	declareCandidacy(TxDeclareCandidacy) error
	updateCandidacy(TxUpdateCandidacy) error
	withdrawCandidacy(TxWithdrawCandidacy) error
	verifyCandidacy(TxVerifyCandidacy) error
	activateCandidacy(TxActivateCandidacy) error
	delegate(TxDelegate) error
	withdraw(TxWithdraw) error
}

//_______________________________________________________________________

// InitState - set genesis parameters for staking
func InitState(key string, value interface{}, store state.SimpleDB) error {
	params := loadParams(store)
	switch key {
	case "self_staking_ratio":
		ratio, err := strconv.ParseFloat(value.(string), 64)
		if err != nil || ratio <= 0 || ratio >= 1 {
			return fmt.Errorf("input must be float, Error: %v", err.Error())
		}
		params.SelfStakingRatio = value.(string)
	case "max_vals":
		i, err := strconv.Atoi(value.(string))
		if err != nil {
			return fmt.Errorf("input must be integer, Error: %v", err.Error())
		}

		params.MaxVals = uint16(i)
	case "validator":
		setValidator(value.(types.GenesisValidator), store)
	default:
		return errors.ErrUnknownKey(key)
	}

	saveParams(store, params)
	return nil
}

func setValidator(val types.GenesisValidator, store state.SimpleDB) error {
	if val.Address == "0000000000000000000000000000000000000000" {
		return ErrBadValidatorAddr()
	}

	addr := common.HexToAddress(val.Address)

	// create and save the empty candidate
	bond := GetCandidateByAddress(addr)
	if bond != nil {
		return ErrCandidateExistsAddr()
	}

	params := loadParams(store)
	deliverer := deliver{
		store:  store,
		sender: addr,
		params: params,
	}

	tx := TxDeclareCandidacy{types.PubKeyString(val.PubKey), utils.ToWei(val.MaxAmount).String(), val.CompRate, Description{}}
	return deliverer.declareGenesisCandidacy(tx, val.Power)
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

	params := loadParams(store)
	checker := check{
		store:  store,
		sender: sender,
		params: params,
		state:  ctx.EthappState(),
	}

	switch txInner := tx.Unwrap().(type) {
	case TxDeclareCandidacy:
		return res, checker.declareCandidacy(txInner)
	case TxUpdateCandidacy:
		return res, checker.updateCandidacy(txInner)
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

	params := loadParams(store)
	deliverer := deliver{
		store:  store,
		sender: sender,
		params: params,
		state:  ctx.EthappState(),
	}

	// Run the transaction
	switch _tx := tx.Unwrap().(type) {
	case TxDeclareCandidacy:
		return res, deliverer.declareCandidacy(_tx)
	case TxUpdateCandidacy:
		return res, deliverer.updateCandidacy(_tx)
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
	params Params
	state  *ethstat.StateDB
}

var _ delegatedProofOfStake = check{} // enforce interface at compile time

func (c check) declareCandidacy(tx TxDeclareCandidacy) error {
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

	// check if the delegator has sufficient funds
	err := checkBalance(c.state, c.sender, rr)
	if err != nil {
		return err
	}

	return nil
}

func (c check) updateCandidacy(tx TxUpdateCandidacy) error {
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
			balance, err := commons.GetBalance(c.state, c.sender)
			if err != nil {
				return err
			}

			if balance.Cmp(rechargeAmount) < 0 {
				return ErrInsufficientFunds()
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
	candidate := GetCandidateByAddress(tx.ValidatorAddress)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	// check if the delegator has sufficient funds
	amount, ok := new(big.Int).SetString(tx.Amount, 10)
	if !ok || amount.Cmp(big.NewInt(0)) < 0 {
		return ErrBadAmount()
	}

	if amount.Cmp(minStakedAmount) < 0 || amount.Cmp(maxStakedAmount) > 0 {
		return ErrInvalidStakedAmount()
	}

	err := checkBalance(c.state, c.sender, amount)
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
	params Params
	state  *ethstat.StateDB
}

var _ delegatedProofOfStake = deliver{} // enforce interface at compile time

// These functions assume everything has been authenticated,
// now we just perform action and save
func (d deliver) declareCandidacy(tx TxDeclareCandidacy) error {
	// create and save the empty candidate
	pubKey, err := types.GetPubKey(tx.PubKey)
	if err != nil {
		return err
	}
	candidate := NewCandidate(pubKey, d.sender, "0", 0, tx.MaxAmount, tx.CompRate, tx.Description, "N", "Y")
	SaveCandidate(candidate)

	// delegate a part of the max staked CMT amount
	amount := tx.SelfStakingAmount(d.params.SelfStakingRatio)
	txDelegate := TxDelegate{ValidatorAddress: d.sender, Amount: amount.String()}
	return d.delegate(txDelegate)
}

func (d deliver) declareGenesisCandidacy(tx TxDeclareCandidacy, votingPower int64) error {
	// create and save the empty candidate
	pubKey, err := types.GetPubKey(tx.PubKey)
	if err != nil {
		return err
	}
	candidate := NewCandidate(pubKey, d.sender, "0", votingPower, tx.MaxAmount, tx.CompRate, tx.Description, "N", "Y")
	SaveCandidate(candidate)

	// delegate a part of the max staked CMT amount
	amount := new(big.Int).Mul(big.NewInt(votingPower), big.NewInt(1e18))
	txDelegate := TxDelegate{ValidatorAddress: d.sender, Amount: amount.String()}
	return d.delegate(txDelegate)
}

func (d deliver) updateCandidacy(tx TxUpdateCandidacy) error {
	// create and save the empty candidate
	candidate := GetCandidateByAddress(d.sender)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	// If the max amount of CMTs is updated, the 10% of self-staking will be re-computed,
	// and the different will be charged
	if tx.MaxAmount != "" {
		maxAmount := utils.ParseInt(tx.MaxAmount)

		if candidate.ParseMaxShares().Cmp(maxAmount) != 0 {
			rechargeAmount := getRechargeAmount(maxAmount, candidate, d.params.SelfStakingRatio)
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
		err := commons.Transfer(d.params.HoldAccount, delegation.DelegatorAddress, delegation.Shares())
		if err != nil {
			return err
		}
		//RemoveDelegation(delegation)
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

	delegation.AddWithdrawAmount(amount)
	if delegation.Shares().Cmp(big.NewInt(0)) == 0 {
		// todo remove or not?
		RemoveDelegation(delegation)
	} else {
		UpdateDelegation(delegation)
	}

	// deduct shares from the candidate
	neg := new(big.Int).Neg(amount)
	candidate.AddShares(neg)
	if candidate.Shares == "0" {
		//candidate.State = "N"
		removeCandidate(candidate)
	}

	now := utils.GetNow()
	candidate.UpdatedAt = now
	updateCandidate(candidate)

	delegateHistory := &DelegateHistory{0, d.sender, candidate.PubKey, amount, "withdraw", now}
	saveDelegateHistory(delegateHistory)

	// transfer coins back to account
	return commons.Transfer(d.params.HoldAccount, d.sender, amount)
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
	award := new(big.Int).Sub(delegation.Shares(), delegation.ParseDelegateAmount())
	needRechargeAmount.Sub(needRechargeAmount, award)
	return
}
