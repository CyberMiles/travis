package stake

import (
	"fmt"
	"strconv"

	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/CyberMiles/travis/modules/coin"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/tendermint/go-wire/data"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"encoding/hex"
)

// nolint
const stakingModuleName = "stake"

// Name is the name of the modules.
func Name() string {
	return stakingModuleName
}

//_______________________________________________________________________

// DelegatedProofOfStake - interface to enforce delegation stake
type delegatedProofOfStake interface {
	declareCandidacy(TxDeclareCandidacy) error
	editCandidacy(TxEditCandidacy) error
	withdraw(TxWithdraw) error
	proposeSlot(TxProposeSlot) ([]byte, error)
	acceptSlot(TxAcceptSlot) error
	withdrawSlot(TxWithdrawSlot) error
	cancelSlot(TxCancelSlot) error
}

type coinSend interface {
	transferFn(sender, receiver sdk.Actor, coins coin.Coins) error
}

//_______________________________________________________________________

// Handler - the transaction processing handler
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
	return stakingModuleName
}

// AssertDispatcher - placeholder for stack.Dispatchable
func (Handler) AssertDispatcher() {}

// InitState - set genesis parameters for staking
func (h Handler) InitState(l log.Logger, store state.SimpleDB,
	module, key, value string, cb sdk.InitStater) (log string, err error) {
	return "", h.initState(module, key, value, store)
}

// separated for testing
func (Handler) initState(module, key, value string, store state.SimpleDB) error {
	if module != stakingModuleName {
		return errors.ErrUnknownModule(module)
	}

	params := loadParams(store)
	switch key {
	case "allowed_bond_denom":
		params.AllowedBondDenom = value
	case "max_vals",
		"gas_bond",
		"gas_unbond":

		// TODO: enforce non-negative integers in input
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

	// create and save the empty candidate
	bond := GetCandidateByAddress(val.Address)
	if bond != nil {
		return ErrCandidateExistsAddr()
	}

	candidate := NewCandidate(val.PubKey, val.Address, val.Power, val.Power, "Y")
	SaveCandidate(candidate)

	return nil
}

// CheckTx checks if the tx is properly structured
func (h Handler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, dispatch sdk.Checker) (res sdk.CheckResult, err error) {

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
		store:  store,
		sender: sender,
		params: params,
		transfer: coinChecker{
			store:    store,
			dispatch: dispatch,
			ctx:      ctx,
		}.transferFn,
	}

	// return the fee for each tx type
	switch txInner := tx.Unwrap().(type) {
	case TxDeclareCandidacy:
		return sdk.NewCheck(params.GasDeclareCandidacy, ""),
			checker.declareCandidacy(txInner)
	case TxEditCandidacy:
		return sdk.NewCheck(params.GasEditCandidacy, ""),
			checker.editCandidacy(txInner)
	case TxWithdraw:
		return sdk.NewCheck(params.GasWithdraw, ""),
			checker.withdraw(txInner)
	case TxProposeSlot:
		_, err := checker.proposeSlot(txInner)
		return sdk.NewCheck(params.GasProposeSlot, ""), err
	case TxAcceptSlot:
		return sdk.NewCheck(params.GasAcceptSlot, ""),
			checker.acceptSlot(txInner)
	case TxWithdrawSlot:
		return sdk.NewCheck(params.GasWithdrawSlot, ""),
			checker.withdrawSlot(txInner)
	case TxCancelSlot:
		return sdk.NewCheck(params.GasCancelSlot, ""),
			checker.cancelSlot(txInner)
	}

	return res, errors.ErrUnknownTxType(tx)
}

// DeliverTx executes the tx if valid
func (h Handler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, dispatch sdk.Deliver) (res sdk.DeliverResult, err error) {

	_, err = h.CheckTx(ctx, store, tx, nil)
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
		transfer: coinSender{
			store:    store,
			dispatch: dispatch,
			ctx:      ctx,
		}.transferFn,
	}

	// Run the transaction
	switch _tx := tx.Unwrap().(type) {
	case TxDeclareCandidacy:
		res.GasUsed = params.GasDeclareCandidacy
		return res, deliverer.declareCandidacy(_tx)
	case TxEditCandidacy:
		res.GasUsed = params.GasEditCandidacy
		return res, deliverer.editCandidacy(_tx)
	case TxWithdraw:
		res.GasUsed = params.GasWithdraw
		return res, deliverer.withdraw(_tx)
	case TxProposeSlot:
		res.GasUsed = params.GasProposeSlot
		id, err := deliverer.proposeSlot(_tx)
		res.Data = []byte(id)
		return res, err
	case TxAcceptSlot:
		res.GasUsed = params.GasAcceptSlot
		return res, deliverer.acceptSlot(_tx)
	case TxWithdrawSlot:
		//context with hold account permissions
		params := loadParams(store)
		res.GasUsed = params.GasWithdrawSlot
		ctx2 := ctx.WithPermissions(params.HoldAccount)
		deliverer.transfer = coinSender{
			store:    store,
			dispatch: dispatch,
			ctx:      ctx2,
		}.transferFn
		return res, deliverer.withdrawSlot(_tx)
	case TxCancelSlot:
		res.GasUsed = params.GasCancelSlot
		return res, deliverer.cancelSlot(_tx)
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

//_______________________________________________________________________

type coinChecker struct {
	store    state.SimpleDB
	dispatch sdk.Checker
	ctx      sdk.Context
}

var _ coinSend = coinSender{} // enforce interface at compile time

func (c coinChecker) transferFn(sender, receiver sdk.Actor, coins coin.Coins) error {
	send := coin.NewSendOneTx(sender, receiver, coins)

	// If the deduction fails (too high), abort the command
	_, err := c.dispatch.CheckTx(c.ctx, c.store, send)
	return err
}


type coinSender struct {
	store    state.SimpleDB
	dispatch sdk.Deliver
	ctx      sdk.Context
}

var _ coinSend = coinSender{} // enforce interface at compile time

func (c coinSender) transferFn(sender, receiver sdk.Actor, coins coin.Coins) error {
	send := coin.NewSendOneTx(sender, receiver, coins)

	// If the deduction fails (too high), abort the command
	_, err := c.dispatch.DeliverTx(c.ctx, c.store, send)
	return err
}

//_____________________________________________________________________

type check struct {
	store  state.SimpleDB
	sender sdk.Actor
	params   Params
	transfer transferFn
}

var _ delegatedProofOfStake = check{} // enforce interface at compile time

func (c check) declareCandidacy(tx TxDeclareCandidacy) error {
	// check to see if the pubkey or address has been registered before
	candidate := GetCandidateByAddress(common.BytesToAddress(c.sender.Address))
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
	candidate := GetCandidateByAddress(common.BytesToAddress(c.sender.Address))
	if candidate == nil {
		return fmt.Errorf("cannot edit non-exsits candidacy")
	}

	return nil
}

func (c check) withdraw(tx TxWithdraw) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(tx.Address)
	if candidate == nil {
		return fmt.Errorf("cannot withdraw pubkey which is not declared"+
			" PubKey %v already registered with %v candidate address",
			candidate.PubKey, candidate.OwnerAddress.String())
	}

	return nil
}

func (c check) withdrawSlot(tx TxWithdrawSlot) error {
	slot := GetSlot(tx.SlotId)
	if slot == nil {
		return ErrBadSlot()
	}

	// check if have enough shares to unbond
	delegatorAddress := common.BytesToAddress(c.sender.Address)
	slotDelegate := GetSlotDelegate(delegatorAddress, tx.SlotId)
	if slotDelegate == nil {
		return ErrBadSlotDelegate()
	}

	if slotDelegate.Amount < tx.Amount {
		return ErrInsufficientFunds()
	}
	return nil
}

func (c check) proposeSlot(tx TxProposeSlot) ([]byte, error) {
	candidate := GetCandidateByAddress(tx.ValidatorAddress)
	if candidate == nil {
		return nil, fmt.Errorf("cannot propose slot for non-existant validator address %v", tx.ValidatorAddress)
	}

	return nil, nil
}

func (c check) acceptSlot(tx TxAcceptSlot) error {
	slot := GetSlot(tx.SlotId)
	if slot == nil {
		return ErrBadSlot()
	}

	// Move coins from the delegator account to the pubKey lock account
	err := c.transfer(c.sender, c.params.HoldAccount, coin.Coins{{"cmt", tx.Amount}})
	if err != nil {
		return err
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
	sender   sdk.Actor
	params   Params
	transfer transferFn
}

type transferFn func(sender, receiver sdk.Actor, coins coin.Coins) error

var _ delegatedProofOfStake = deliver{} // enforce interface at compile time

// These functions assume everything has been authenticated,
// now we just perform action and save
func (d deliver) declareCandidacy(tx TxDeclareCandidacy) error {

	// create and save the empty candidate
	ownerAddress := common.BytesToAddress(d.sender.Address)
	candidate := GetCandidateByAddress(ownerAddress)
	if candidate != nil && candidate.State == "Y" {
		return ErrCandidateExistsAddr()
	}

	if candidate == nil {
		candidate = NewCandidate(tx.PubKey, ownerAddress, 0, 0, "Y")
		SaveCandidate(candidate)
	} else {
		candidate.State = "Y"
		candidate.OwnerAddress = ownerAddress
		candidate.UpdatedAt = utils.GetNow()
		updateCandidate(candidate)
	}

	return nil
}

func (d deliver) editCandidacy(tx TxEditCandidacy) error {

	// create and save the empty candidate
	ownerAddress := common.BytesToAddress(d.sender.Address)
	candidate := GetCandidateByAddress(ownerAddress)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	candidate.OwnerAddress = tx.NewAddress
	candidate.UpdatedAt = utils.GetNow()
	updateCandidate(candidate)

	return nil
}

func (d deliver) withdraw(tx TxWithdraw) error {

	// create and save the empty candidate
	validatorAddress := common.BytesToAddress(d.sender.Address)
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
			delegator := sdk.Actor{"", stakingModuleName, delegate.DelegatorAddress.Bytes()}
			d.transfer(d.params.HoldAccount, delegator,
				coin.Coins{{d.params.AllowedBondDenom, delegate.Amount}})
			delegate.Amount = 0
			saveSlotDelegate(*delegate)
		}
		slot.AvailableAmount = 0
		updateSlot(slot)
	}
	candidate.Shares = 0
	candidate.VotingPower = 0
	candidate.UpdatedAt = utils.GetNow()
	candidate.State = "N"
	updateCandidate(candidate)

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

	if tx.Amount > slot.AvailableAmount {
		return ErrFullSlot()
	}

	// Move coins from the delegator account to the pubKey lock account
	err := d.transfer(d.sender, d.params.HoldAccount, coin.Coins{{"cmt", tx.Amount}})
	if err != nil {
		return err
	}

	// Get or create the delegate slot
	delegatorAddress := common.BytesToAddress(d.sender.Address)
	slotDelegate := GetSlotDelegate(delegatorAddress, tx.SlotId)
	if slotDelegate == nil {
		slotDelegate = NewSlotDelegate(delegatorAddress, tx.SlotId, 0)
	}

	// Add shares to slot and candidate
	slotDelegate.Amount += tx.Amount
	candidate.Shares += uint64(tx.Amount)
	slot.AvailableAmount -= tx.Amount

	delegateHistory := DelegateHistory{delegatorAddress, tx.SlotId, tx.Amount, "accept"}

	updateCandidate(candidate)
	updateSlot(slot)
	saveSlotDelegate(*slotDelegate)
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
	delegatorAddress := common.BytesToAddress(d.sender.Address)
	slotDelegate := GetSlotDelegate(delegatorAddress, tx.SlotId)
	if slotDelegate == nil {
		return ErrBadSlotDelegate()
	}

	// get pubKey candidate
	candidate := GetCandidateByAddress(slot.ValidatorAddress)
	if candidate == nil {
		return ErrNoCandidateForAddress()
	}

	// subtract bond tokens from bond
	if slotDelegate.Amount < tx.Amount {
		return ErrInsufficientFunds()
	}
	slotDelegate.Amount -= tx.Amount

	if slotDelegate.Amount == 0 {
		// remove the slot delegate
		removeSlotDelegate(*slotDelegate)
	} else {
		saveSlotDelegate(*slotDelegate)
	}

	// deduct shares from the candidate
	candidate.Shares -= uint64(tx.Amount)
	if candidate.Shares == 0 {
		candidate.State = "N"
	}
	updateCandidate(candidate)

	slot.AvailableAmount -= tx.Amount
	updateSlot(slot)

	// transfer coins back to account
	returnCoins := tx.Amount
	return d.transfer(d.params.HoldAccount, d.sender,
		coin.Coins{{d.params.AllowedBondDenom, returnCoins}})
}

func (d deliver) proposeSlot(tx TxProposeSlot) ([]byte, error) {
	//hash := merkle.SimpleHashFromBinary(tx)
	//hexHash := hex.EncodeToString(hash)
	uuid := utils.GetUUID()
	hexStr := hex.EncodeToString(uuid)
	slot := NewSlot(hexStr, tx.ValidatorAddress, tx.Amount, tx.Amount, tx.ProposedRoi, "Y")
	saveSlot(slot)

	return uuid, nil
}

func (d deliver) cancelSlot(tx TxCancelSlot) error {
	slot := GetSlot(tx.SlotId)
	if slot == nil {
		return ErrBadSlot()
	}

	if slot.State == "N" {
		return ErrCancelledSlot()
	}

	slot.AvailableAmount = 0
	slot.State = "N"
	slot.UpdatedAt = utils.GetNow()
	updateSlot(slot)

	return nil
}