package stake

import (
	"fmt"
	"strconv"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/tendermint/go-wire/data"
	"math/big"
)

// nolint
const stakingModuleName = "stake"

//_______________________________________________________________________

// DelegatedProofOfStake - interface to enforce delegation stake
type delegatedProofOfStake interface {
	declareCandidacy(TxDeclareCandidacy) error
	updateCandidacy(TxUpdateCandidacy) error
	withdrawCandidacy(TxWithdrawCandidacy) error
	verifyCandidacy(TxVerifyCandidacy) error
	delegate(TxDelegate) error
	withdraw(TxWithdraw) error
}

//_______________________________________________________________________

// InitState - set genesis parameters for staking
func InitState(key, value string, store state.SimpleDB) error {
	params := loadParams(store)
	switch key {
	case "reserve_requirement_ratio":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("input must be float, Error: %v", err.Error())
		}
		params.ReserveRequirementRatio = v
	case "max_vals":
		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("input must be integer, Error: %v", err.Error())
		}

		params.MaxVals = uint16(i)
	case "validator":
		setValidator(value)
	default:
		return errors.ErrUnknownKey(key)
	}

	saveParams(store, params)
	return nil
}

func setValidator(value string) error {
	var val genesisValidator
	err := data.FromJSON([]byte(value), &val)
	if err != nil {
		return fmt.Errorf("error reading validators")
	}

	if val.Address == common.HexToAddress("0000000000000000000000000000000000000000") {
		return ErrBadValidatorAddr()
	}

	// create and save the empty candidate
	bond := GetCandidateByAddress(val.Address)
	if bond != nil {
		return ErrCandidateExistsAddr()
	}

	shares := new(big.Int)
	shares.Mul(big.NewInt(val.Power), big.NewInt(1e18))
	candidate := NewCandidate(val.PubKey, val.Address, shares, val.Power, "Y", Description{})
	SaveCandidate(candidate)
	return nil
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
		store:    store,
		sender:   sender,
		params:   params,
		ethereum: ctx.Ethereum(),
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
	case TxDelegate:
		return res, checker.delegate(txInner)
	case TxWithdraw:
		return res, checker.withdraw(txInner)
	}

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
		store:    store,
		sender:   sender,
		params:   params,
		ethereum: ctx.Ethereum(),
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
	store    state.SimpleDB
	sender   common.Address
	params   Params
	ethereum *eth.Ethereum
}

var _ delegatedProofOfStake = check{} // enforce interface at compile time

func (c check) declareCandidacy(tx TxDeclareCandidacy) error {
	// check to see if the pubkey or address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate != nil && candidate.State == "Y" {
		return fmt.Errorf("address has been declared")
	}

	candidate = GetCandidateByPubKey(tx.PubKey.KeyString())
	if candidate != nil && candidate.State == "Y" {
		return fmt.Errorf("pubkey has been declared")
	}

	// todo check to see if the associated account has 10%(RRR, short for Reserve Requirement Ratio, configurable) of the max staked CMT amount

	return nil
}

func (c check) updateCandidacy(tx TxUpdateCandidacy) error {
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return fmt.Errorf("cannot edit non-exsits candidacy")
	}

	// todo If the max amount of CMTs is updated, the 10% of self-staking will be re-computed,
	// and the different will be charged or refunded from / into the new account address.

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

	// todo check to see if the request was initiated by a special account

	return nil
}

func (c check) delegate(tx TxDelegate) error {
	// check if the delegator has sufficient funds
	balance, err := commons.GetBalance(c.ethereum, c.sender)
	if err != nil {
		return err
	}

	amount := new(big.Int)
	_, ok := amount.SetString(tx.Amount, 10)
	if !ok {
		return ErrBadAmount()
	}

	if balance.Cmp(amount) < 0 {
		return ErrInsufficientFunds()
	}

	// todo check to see if the validator has reached its declared max amount CMTs to be staked.

	return nil
}

func (c check) withdraw(tx TxWithdraw) error {
	// todo check if has delegated

	return nil
}

//_____________________________________________________________________

type deliver struct {
	store    state.SimpleDB
	sender   common.Address
	params   Params
	ethereum *eth.Ethereum
}

var _ delegatedProofOfStake = deliver{} // enforce interface at compile time

// These functions assume everything has been authenticated,
// now we just perform action and save
func (d deliver) declareCandidacy(tx TxDeclareCandidacy) error {
	// create and save the empty candidate
	candidate := GetCandidateByAddress(d.sender)
	if candidate != nil && candidate.State == "Y" {
		return ErrCandidateExistsAddr()
	}

	if candidate == nil {
		candidate = NewCandidate(tx.PubKey, d.sender, big.NewInt(0), 0, "Y", tx.Description)
		SaveCandidate(candidate)
	} else {
		candidate.State = "Y"
		candidate.OwnerAddress = d.sender
		candidate.UpdatedAt = utils.GetNow()
		updateCandidate(candidate)
	}

	// todo delegate a part of the max staked CMT amount

	return nil
}

func (d deliver) updateCandidacy(tx TxUpdateCandidacy) error {

	// create and save the empty candidate
	candidate := GetCandidateByAddress(d.sender)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	candidate.OwnerAddress = tx.NewAddress
	candidate.UpdatedAt = utils.GetNow()
	updateCandidate(candidate)

	// todo if the max amount of CMTs is updated, check if the associated account has enough CMT amount(RRR)

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
	slots := GetSlotsByValidator(validatorAddress)
	for _, slot := range slots {
		slotId := slot.Id
		delegates := GetSlotDelegatesBySlot(slotId)
		for _, delegate := range delegates {
			err := commons.Transfer(d.params.HoldAccount, delegate.DelegatorAddress, delegate.Amount)
			if err != nil {
				return err
			}

			//delegate.Amount = 0
			//saveSlotDelegate(delegate)
			removeSlotDelegate(delegate)
		}
		slot.AvailableAmount = big.NewInt(0)
		updateSlot(slot)
	}
	//candidate.Shares = 0
	//candidate.UpdatedAt = utils.GetNow()
	//candidate.State = "N"
	//updateCandidate(candidate)
	removeCandidate(candidate)

	return nil
}

func (d deliver) verifyCandidacy(tx TxVerifyCandidacy) error {
	// todo verify candidacy

	return nil
}

func (d deliver) delegate(tx TxDelegate) error {
	// Get the slot
	slot := GetSlot(tx.SlotId)

	// Get the pubKey bond account
	candidate := GetCandidateByAddress(slot.ValidatorAddress)
	if candidate == nil {
		return ErrBondNotNominated()
	}

	amount := new(big.Int)
	_, ok := amount.SetString(tx.Amount, 10)
	if !ok {
		return ErrBadAmount()
	}

	if amount.Cmp(slot.AvailableAmount) > 0 {
		return ErrFullSlot()
	}

	// Move coins from the delegator account to the pubKey lock account
	err := commons.Transfer(d.sender, d.params.HoldAccount, amount)
	if err != nil {
		return err
	}

	// Get or create the delegate slot
	now := utils.GetNow()
	slotDelegate := GetSlotDelegate(d.sender, tx.SlotId)
	if slotDelegate == nil {
		slotDelegate = &SlotDelegate{DelegatorAddress: d.sender, SlotId: tx.SlotId, Amount: amount, CreatedAt: now, UpdatedAt: now}
		saveSlotDelegate(slotDelegate)
	} else {
		slotDelegate.Amount.Add(slotDelegate.Amount, amount)
		updateSlotDelegate(slotDelegate)
	}

	// Add shares to slot and candidate
	candidate.Shares.Add(candidate.Shares, amount)
	slot.AvailableAmount.Sub(slot.AvailableAmount, amount)
	delegateHistory := DelegateHistory{d.sender, tx.SlotId, amount, "accept", now}
	updateCandidate(candidate)
	updateSlot(slot)
	saveDelegateHistory(delegateHistory)

	return nil
}

func (d deliver) withdraw(tx TxWithdraw) error {
	// Get the slot
	slot := GetSlot(tx.SlotId)

	if slot == nil {
		return ErrBadSlot()
	}

	// get the slot delegate
	slotDelegate := GetSlotDelegate(d.sender, tx.SlotId)
	if slotDelegate == nil {
		return ErrBadSlotDelegate()
	}

	// get pubKey candidate
	candidate := GetCandidateByAddress(slot.ValidatorAddress)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	amount := new(big.Int)
	_, ok := amount.SetString(tx.Amount, 10)
	if !ok {
		return ErrBadAmount()
	}

	// subtract bond tokens from bond
	if slotDelegate.Amount.Cmp(amount) < 0 {
		return ErrInsufficientFunds()
	}
	slotDelegate.Amount.Sub(slotDelegate.Amount, amount)

	if slotDelegate.Amount.Cmp(big.NewInt(0)) == 0 {
		// remove the slot delegate
		removeSlotDelegate(slotDelegate)
	} else {
		updateSlotDelegate(slotDelegate)
	}

	// deduct shares from the candidate
	candidate.Shares.Sub(candidate.Shares, amount)
	if candidate.Shares.Cmp(big.NewInt(0)) == 0 {
		//candidate.State = "N"
		removeCandidate(candidate)
	}

	now := utils.GetNow()
	candidate.UpdatedAt = now
	updateCandidate(candidate)

	slot.AvailableAmount.Add(slot.AvailableAmount, amount)
	slot.UpdatedAt = now
	updateSlot(slot)

	delegateHistory := DelegateHistory{d.sender, tx.SlotId, amount, "withdraw", now}
	saveDelegateHistory(delegateHistory)

	// transfer coins back to account
	return commons.Transfer(d.params.HoldAccount, d.sender, amount)
}
