package stake

import (
	"fmt"
	"strconv"

	"encoding/hex"
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

// Name is the name of the commons.
func Name() string {
	return stakingModuleName
}

//_______________________________________________________________________

// DelegatedProofOfStake - interface to enforce delegation stake
type delegatedProofOfStake interface {
	declareCandidacy(TxDeclareCandidacy) error
	editCandidacy(TxEditCandidacy) error
	withdrawCandidacy(TxWithdrawCandidacy) error
	proposeSlot(TxProposeSlot, []byte) error
	acceptSlot(TxAcceptSlot) error
	withdrawSlot(TxWithdrawSlot) error
	cancelSlot(TxCancelSlot) error
}

//_______________________________________________________________________

// InitState - set genesis parameters for staking
func InitState(key, value string, store state.SimpleDB) error {
	params := loadParams(store)
	switch key {
	case "allowed_bond_denom":
		params.AllowedBondDenom = value
	case "max_vals",
		"gas_bond",
		"gas_unbond":

		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("input must be integer, Error: %v", err.Error())
		}

		switch key {
		case "max_vals":
			params.MaxVals = uint16(i)
		}
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
	candidate := NewCandidate(val.PubKey, val.Address, shares, val.Power, "Y")
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

	// create the new checker object to
	checker := check{
		store:    store,
		sender:   sender,
		params:   params,
		ethereum: ctx.Ethereum(),
	}

	// return the fee for each tx type
	switch txInner := tx.Unwrap().(type) {
	case TxDeclareCandidacy:
		return res, checker.declareCandidacy(txInner)
	case TxEditCandidacy:
		return res, checker.editCandidacy(txInner)
	case TxWithdrawCandidacy:
		return res, checker.withdrawCandidacy(txInner)
	case TxProposeSlot:
		err := checker.proposeSlot(txInner, []byte{})
		return res, err
	case TxAcceptSlot:
		return res, checker.acceptSlot(txInner)
	case TxWithdrawSlot:
		return res, checker.withdrawSlot(txInner)
	case TxCancelSlot:
		return res, checker.cancelSlot(txInner)
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
	case TxEditCandidacy:
		return res, deliverer.editCandidacy(_tx)
	case TxWithdrawCandidacy:
		return res, deliverer.withdrawCandidacy(_tx)
	case TxProposeSlot:
		err := deliverer.proposeSlot(_tx, hash)
		res.Data = hash
		return res, err
	case TxAcceptSlot:
		return res, deliverer.acceptSlot(_tx)
	case TxWithdrawSlot:
		return res, deliverer.withdrawSlot(_tx)
	case TxCancelSlot:
		return res, deliverer.cancelSlot(_tx)
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

	return nil
}

func (c check) editCandidacy(tx TxEditCandidacy) error {
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return fmt.Errorf("cannot edit non-exsits candidacy")
	}

	return nil
}

func (c check) withdrawCandidacy(tx TxWithdrawCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(tx.Address)
	if candidate == nil {
		return fmt.Errorf("cannot withdrawCandidacy non-exsits candidacy")
	}

	return nil
}

func (c check) withdrawSlot(tx TxWithdrawSlot) error {
	slot := GetSlot(tx.SlotId)
	if slot == nil {
		return ErrBadSlot()
	}

	// check if have enough shares to unbond
	slotDelegate := GetSlotDelegate(c.sender, tx.SlotId)
	if slotDelegate == nil {
		return ErrBadSlotDelegate()
	}

	amount := new(big.Int)
	_, ok := amount.SetString(tx.Amount, 10)
	if !ok {
		return ErrBadAmount()
	}

	if slotDelegate.Amount.Cmp(amount) < 0 {
		return ErrInsufficientFunds()
	}
	return nil
}

func (c check) proposeSlot(tx TxProposeSlot, hash []byte) error {
	candidate := GetCandidateByAddress(tx.ValidatorAddress)
	if candidate == nil {
		return fmt.Errorf("cannot propose slot for non-existant validator address %v", tx.ValidatorAddress)
	}

	return nil
}

func (c check) acceptSlot(tx TxAcceptSlot) error {
	slot := GetSlot(tx.SlotId)
	if slot == nil {
		return ErrBadSlot()
	}

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
	return nil
}

func (c check) cancelSlot(tx TxCancelSlot) error {
	slot := GetSlot(tx.SlotId)
	if slot == nil {
		return ErrBadSlot()
	}

	if slot.State == "N" {
		return ErrCancelledSlot()
	}

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
		candidate = NewCandidate(tx.PubKey, d.sender, big.NewInt(0), 0, "Y")
		SaveCandidate(candidate)
	} else {
		candidate.State = "Y"
		candidate.OwnerAddress = d.sender
		candidate.UpdatedAt = utils.GetNow()
		updateCandidate(candidate)
	}

	return nil
}

func (d deliver) editCandidacy(tx TxEditCandidacy) error {

	// create and save the empty candidate
	candidate := GetCandidateByAddress(d.sender)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	candidate.OwnerAddress = tx.NewAddress
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

func (d deliver) acceptSlot(tx TxAcceptSlot) error {
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

func (d deliver) withdrawSlot(tx TxWithdrawSlot) error {
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

func (d deliver) proposeSlot(tx TxProposeSlot, hash []byte) error {
	amount := new(big.Int)
	_, ok := amount.SetString(tx.Amount, 10)
	if !ok {
		return ErrBadAmount()
	}

	now := utils.GetNow()
	slot := &Slot{
		Id:               hex.EncodeToString(hash),
		ValidatorAddress: tx.ValidatorAddress,
		TotalAmount:      amount,
		AvailableAmount:  amount,
		ProposedRoi:      tx.ProposedRoi,
		State:            "Y",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	saveSlot(slot)
	return nil
}

func (d deliver) cancelSlot(tx TxCancelSlot) error {
	slot := GetSlot(tx.SlotId)
	if slot == nil {
		return ErrBadSlot()
	}

	removeSlot(slot)

	//if slot.State == "N" {
	//	return ErrCancelledSlot()
	//}
	//
	//slot.AvailableAmount = 0
	//slot.State = "N"
	//slot.UpdatedAt = utils.GetNow()
	//updateSlot(slot)

	return nil
}
