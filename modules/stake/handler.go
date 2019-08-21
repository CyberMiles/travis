package stake

import (
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	ethstat "github.com/ethereum/go-ethereum/core/state"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/errors"
	"github.com/CyberMiles/travis/sdk/state"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
)

// DelegatedProofOfStake - interface to enforce delegation stake
type delegatedProofOfStake interface {
	declareCandidacy(TxDeclareCandidacy, sdk.Int) error
	updateCandidacy(TxUpdateCandidacy, sdk.Int) error
	withdrawCandidacy(TxWithdrawCandidacy) error
	verifyCandidacy(TxVerifyCandidacy) error
	activateCandidacy(TxActivateCandidacy) error
	deactivateCandidacy(TxDeactivateCandidacy) error
	delegate(TxDelegate) error
	withdraw(TxWithdraw) error
	setCompRate(TxSetCompRate, sdk.Int) error
	updateCandidateAccount(TxUpdateCandidacyAccount, sdk.Int) (int64, error)
	acceptCandidateAccountUpdateRequest(TxAcceptCandidacyAccountUpdate, sdk.Int) error
}

func SetGenesisValidator(val types.GenesisValidator, store state.SimpleDB) error {
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
		ctx:    types.NewContext("", 0, 0, nil),
	}

	desc := Description{
		Name:     val.Name,
		Website:  val.Website,
		Location: val.Location,
		Email:    val.Email,
		Profile:  val.Profile,
	}

	tx := TxDeclareCandidacy{types.PubKeyString(val.PubKey), utils.ToWei(val.MaxAmount).String(), val.CompRate, desc}
	return deliverer.declareGenesisCandidacy(tx, val)
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
		ctx:    ctx,
	}

	switch txInner := tx.Unwrap().(type) {
	case TxDeclareCandidacy:
		gasFee := utils.CalGasFee(params.DeclareCandidacyGas, params.GasPrice)
		return res, checker.declareCandidacy(txInner, gasFee)
	case TxUpdateCandidacy:
		gasFee := utils.CalGasFee(params.UpdateCandidacyGas, params.GasPrice)
		return res, checker.updateCandidacy(txInner, gasFee)
	case TxWithdrawCandidacy:
		return res, checker.withdrawCandidacy(txInner)
	case TxVerifyCandidacy:
		return res, checker.verifyCandidacy(txInner)
	case TxActivateCandidacy:
		return res, checker.activateCandidacy(txInner)
	case TxDeactivateCandidacy:
		return res, checker.deactivateCandidacy(txInner)
	case TxDelegate:
		return res, checker.delegate(txInner)
	case TxWithdraw:
		return res, checker.withdraw(txInner)
	case TxSetCompRate:
		gasFee := utils.CalGasFee(params.SetCompRateGas, params.GasPrice)
		return res, checker.setCompRate(txInner, gasFee)
	case TxUpdateCandidacyAccount:
		gasFee := utils.CalGasFee(params.UpdateCandidateAccountGas, params.GasPrice)
		_, err := checker.updateCandidateAccount(txInner, gasFee)
		return res, err
	case TxAcceptCandidacyAccountUpdate:
		gasFee := utils.CalGasFee(params.AcceptCandidateAccountUpdateRequestGas, params.GasPrice)
		return res, checker.acceptCandidateAccountUpdateRequest(txInner, gasFee)
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

	params := utils.GetParams()
	deliverer := deliver{
		store:  store,
		sender: sender,
		params: params,
		ctx:    ctx,
	}
	res.GasFee = big.NewInt(0)

	// Run the transaction
	switch txInner := tx.Unwrap().(type) {
	case TxDeclareCandidacy:
		gasFee := utils.CalGasFee(params.DeclareCandidacyGas, params.GasPrice)
		err := deliverer.declareCandidacy(txInner, gasFee)
		if err == nil {
			res.GasUsed = int64(params.DeclareCandidacyGas)
			res.GasFee = gasFee.Int
		}
		return res, err
	case TxUpdateCandidacy:
		gasFee := utils.CalGasFee(params.UpdateCandidacyGas, params.GasPrice)
		err := deliverer.updateCandidacy(txInner, gasFee)
		if err == nil {
			res.GasUsed = int64(params.UpdateCandidacyGas)
			res.GasFee = gasFee.Int
		}
		return res, err
	case TxWithdrawCandidacy:
		return res, deliverer.withdrawCandidacy(txInner)
	case TxVerifyCandidacy:
		return res, deliverer.verifyCandidacy(txInner)
	case TxActivateCandidacy:
		return res, deliverer.activateCandidacy(txInner)
	case TxDeactivateCandidacy:
		return res, deliverer.deactivateCandidacy(txInner)
	case TxDelegate:
		return res, deliverer.delegate(txInner)
	case TxWithdraw:
		return res, deliverer.withdraw(txInner)
	case TxSetCompRate:
		gasFee := utils.CalGasFee(params.SetCompRateGas, params.GasPrice)
		err := deliverer.setCompRate(txInner, gasFee)
		if err == nil {
			res.GasUsed = int64(params.SetCompRateGas)
			res.GasFee = gasFee.Int
		}
		return res, err
	case TxUpdateCandidacyAccount:
		gasFee := utils.CalGasFee(params.UpdateCandidateAccountGas, params.GasPrice)
		id, err := deliverer.updateCandidateAccount(txInner, gasFee)
		if err == nil {
			res.GasUsed = int64(params.UpdateCandidateAccountGas)
			res.GasFee = gasFee.Int
		}
		res.Data = []byte(strconv.Itoa(int(id)))
		return res, err
	case TxAcceptCandidacyAccountUpdate:
		gasFee := utils.CalGasFee(params.AcceptCandidateAccountUpdateRequestGas, params.GasPrice)
		err := deliverer.acceptCandidateAccountUpdateRequest(txInner, gasFee)
		if err == nil {
			res.GasUsed = int64(params.AcceptCandidateAccountUpdateRequestGas)
			res.GasFee = gasFee.Int
		}
		return res, err
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
	ctx    types.Context
}

var _ delegatedProofOfStake = check{} // enforce interface at compile time

func (c check) declareCandidacy(tx TxDeclareCandidacy, gasFee sdk.Int) error {
	pk, err := types.GetPubKey(tx.PubKey)
	if err != nil {
		return err
	}

	// check to see if the pubkey or address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate != nil {
		return ErrAddressAlreadyDeclared()
	}

	candidate = GetCandidateByPubKey(pk)
	if candidate != nil {
		return ErrPubKeyAleadyDeclared()
	}

	// check to see if the associated account has 10%(ssr, short for self-staking ratio, configurable) of the max staked CMT amount
	maxAmount, ok := sdk.NewIntFromString(tx.MaxAmount)
	if !ok || maxAmount.LTE(sdk.ZeroInt) {
		return ErrBadAmount()
	}

	ss := tx.SelfStakingAmount(c.params.SelfStakingRatio)
	totalCost := ss.Add(gasFee)

	// check if the delegator has sufficient funds
	if err := checkBalance(c.ctx.EthappState(), c.sender, totalCost); err != nil {
		return err
	}

	// Check to see if the compensation rate is between 0 and 1
	if tx.CompRate.IsNil() || tx.CompRate.LTE(sdk.ZeroRat) || tx.CompRate.GTE(sdk.OneRat) {
		return ErrBadCompRate()
	}

	return nil
}

func (c check) updateCandidacy(tx TxUpdateCandidacy, gasFee sdk.Int) error {
	if !utils.IsBlank(tx.PubKey) {
		_, err := types.GetPubKey(tx.PubKey)
		if err != nil {
			return err
		}
	}

	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	if sdk.ZeroInt.Equal(candidate.ParseShares()) {
		return ErrCandidateAlreadyWithdrew()
	}

	totalCost := gasFee
	// If the max amount of CMTs is updated, the 10% of self-staking will be re-computed,
	// and the different will be charged
	if tx.MaxAmount != "" {
		maxAmount, ok := sdk.NewIntFromString(tx.MaxAmount)
		if !ok || maxAmount.LTE(sdk.ZeroInt) {
			return ErrBadAmount()
		}

		if maxAmount.GT(candidate.ParseMaxShares()) {
			rechargeAmount := getRechargeAmount(maxAmount, candidate, c.params.SelfStakingRatio)
			totalCost = totalCost.Add(rechargeAmount)
		}
	}

	// check if the delegator has sufficient funds
	if err := checkBalance(c.ctx.EthappState(), c.sender, totalCost); err != nil {
		return err
	}

	if !tx.CompRate.IsNil() && (tx.CompRate.LT(sdk.ZeroRat) || tx.CompRate.GTE(sdk.OneRat)) {
		return ErrBadCompRate()
	}

	if !utils.IsBlank(tx.PubKey) {
		pk, err := types.GetPubKey(tx.PubKey)
		if err != nil {
			return err
		}

		candidate = GetCandidateByPubKey(pk)
		if candidate != nil {
			return ErrPubKeyAleadyDeclared()
		}
	}

	return nil
}

func (c check) withdrawCandidacy(tx TxWithdrawCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	return nil
}

func (c check) verifyCandidacy(tx TxVerifyCandidacy) error {
	// check to see if the candidate address to be verified has been registered before
	candidate := GetCandidateByAddress(tx.CandidateAddress)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	if sdk.ZeroInt.Equal(candidate.ParseShares()) {
		return ErrCandidateAlreadyWithdrew()
	}

	// check to see if the request was initiated by a special account
	if c.sender != common.HexToAddress(utils.GetParams().FoundationAddress) {
		return ErrVerificationDisallowed()
	}

	return nil
}

func (c check) activateCandidacy(tx TxActivateCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	if candidate.Active == "Y" {
		return ErrCandidateAlreadyActivated()
	}

	if candidate.ParseShares().Equal(sdk.ZeroInt) {
		return ErrCandidateAlreadyWithdrew()
	}

	return nil
}

func (c check) deactivateCandidacy(tx TxDeactivateCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	if candidate.Active == "N" {
		return ErrCandidateAlreadyDeactivated()
	}

	if sdk.ZeroInt.Equal(candidate.ParseShares()) {
		return ErrCandidateAlreadyWithdrew()
	}

	return nil
}

func (c check) delegate(tx TxDelegate) error {
	if tx.Sig != "" {
		err := VerifyCubeSignature(c.sender, c.ctx.GetNonce(), tx.CubeBatch, tx.Sig)
		if err != nil {
			return err
		}
	}

	candidate := GetCandidateByAddress(tx.ValidatorAddress)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	if sdk.ZeroInt.Equal(candidate.ParseShares()) {
		return ErrCandidateAlreadyWithdrew()
	}

	// check if the delegator has sufficient funds
	amount, ok := sdk.NewIntFromString(tx.Amount)
	if !ok || amount.LTE(sdk.ZeroInt) {
		return ErrBadAmount()
	}

	err := checkBalance(c.ctx.EthappState(), c.sender, amount)
	if err != nil {
		return err
	}

	d := GetDelegation(c.sender, candidate.Id)
	if d != nil && d.CompletelyWithdraw == "Y" {
		return ErrDelegatorHasPendingWithdrawal()
	}

	// check to see if the simpleValidator has reached its declared max amount CMTs to be staked.
	if candidate.ParseShares().Add(amount).GT(candidate.ParseMaxShares()) {
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

	amount, ok := sdk.NewIntFromString(tx.Amount)
	if !ok || amount.LTE(sdk.ZeroInt) {
		return ErrBadAmount()
	}

	d := GetDelegation(c.sender, candidate.Id)
	if d == nil {
		return ErrDelegationNotExists()
	}

	if d.CompletelyWithdraw == "Y" {
		return ErrDelegatorHasPendingWithdrawal()
	}

	if amount.GT(d.Shares()) {
		return ErrInvalidWithdrawalAmount()
	}

	return nil
}

func (c check) setCompRate(tx TxSetCompRate, gasFee sdk.Int) error {
	// Check to see if the compensation rate is between 0 and 1
	if tx.CompRate.IsNil() || tx.CompRate.LTE(sdk.ZeroRat) || tx.CompRate.GTE(sdk.OneRat) {
		return ErrBadCompRate()
	}

	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	if sdk.ZeroInt.Equal(candidate.ParseShares()) {
		return ErrCandidateAlreadyWithdrew()
	}

	d := GetDelegation(tx.DelegatorAddress, candidate.Id)
	if d == nil {
		return ErrDelegationNotExists()
	}

	if tx.CompRate.GT(candidate.CompRate) {
		return ErrBadCompRate()
	}

	// check if the delegator has sufficient funds
	if err := checkBalance(c.ctx.EthappState(), c.sender, gasFee); err != nil {
		return err
	}

	return nil
}

func (c check) updateCandidateAccount(tx TxUpdateCandidacyAccount, gasFee sdk.Int) (int64, error) {
	candidate := GetCandidateByAddress(c.sender)
	if candidate == nil {
		return 0, ErrBadRequest()
	}

	if sdk.ZeroInt.Equal(candidate.ParseShares()) {
		return 0, ErrCandidateAlreadyWithdrew()
	}

	tmp := GetCandidateByAddress(tx.NewCandidateAddress)
	if tmp != nil {
		return 0, ErrBadRequest()
	}

	// check if the new address has been delegated the candidate
	d := GetDelegation(tx.NewCandidateAddress, candidate.Id)
	if d != nil {
		return 0, ErrBadRequest()
	}

	// check if the new address has been used
	exists := getCandidateAccountUpdateRequestByToAddress(tx.NewCandidateAddress)
	if len(exists) > 0 {
		return 0, ErrBadRequest()
	}

	// check if the address has been changed
	ownerAddress := common.HexToAddress(candidate.OwnerAddress)
	if utils.IsEmptyAddress(tx.NewCandidateAddress) || tx.NewCandidateAddress == ownerAddress {
		return 0, ErrBadRequest()
	}

	// check if the candidate has sufficient funds
	if err := checkBalance(c.ctx.EthappState(), c.sender, gasFee); err != nil {
		return 0, err
	}

	return 0, nil
}

func (c check) acceptCandidateAccountUpdateRequest(tx TxAcceptCandidacyAccountUpdate, gasFee sdk.Int) error {
	req := getCandidateAccountUpdateRequestById(tx.AccountUpdateRequestId)
	if req == nil {
		return ErrBadRequest()
	}

	tmp := GetCandidateByAddress(req.ToAddress)
	if tmp != nil {
		return ErrBadRequest()
	}

	// check if the new address has been delegated the candidate
	d := GetDelegation(req.ToAddress, req.CandidateId)
	if d != nil {
		return ErrBadRequest()
	}

	if req.ToAddress != c.sender || req.State != "PENDING" {
		return ErrBadRequest()
	}

	delegation := GetDelegation(req.FromAddress, req.CandidateId)
	totalCost := delegation.Shares().Add(gasFee)

	// check if the new account has sufficient funds
	if err := checkBalance(c.ctx.EthappState(), req.ToAddress, totalCost); err != nil {
		return err
	}

	// check if the candidate has some pending withdrawal requests
	reqs := GetUnstakeRequestsByDelegator(req.FromAddress)
	if len(reqs) > 0 {
		return ErrCandidateHasPendingUnstakeRequests()
	}

	return nil
}

//_____________________________________________________________________

type deliver struct {
	store  state.SimpleDB
	sender common.Address
	params *utils.Params
	ctx    types.Context
}

var _ delegatedProofOfStake = deliver{} // enforce interface at compile time

// These functions assume everything has been authenticated,
// now we just perform action and save
func (d deliver) declareCandidacy(tx TxDeclareCandidacy, gasFee sdk.Int) error {
	// create and save the empty candidate
	pubKey, err := types.GetPubKey(tx.PubKey)
	if err != nil {
		return err
	}

	now := d.ctx.BlockTime()
	candidate := &Candidate{
		PubKey:       pubKey,
		OwnerAddress: d.sender.String(),
		Shares:       "0",
		VotingPower:  0,
		MaxShares:    tx.MaxAmount,
		CompRate:     tx.CompRate,
		CreatedAt:    now,
		Description:  tx.Description,
		Verified:     "N",
		Active:       "Y",
		BlockHeight:  d.ctx.BlockHeight(),
		State:        "Candidate",
	}
	// delegate a part of the max staked CMT amount
	amount := tx.SelfStakingAmount(d.params.SelfStakingRatio)
	totalCost := amount.Add(gasFee)

	// check if the delegator has sufficient funds
	if err := checkBalance(d.ctx.EthappState(), d.sender, totalCost); err != nil {
		return err
	}

	// only charge gas fee here
	//commons.Transfer(d.sender, utils.HoldAccount, gasFee)
	d.ctx.EthappState().SubBalance(d.sender, gasFee.Int)
	d.ctx.EthappState().AddBalance(utils.HoldAccount, gasFee.Int)
	SaveCandidate(candidate)

	txDelegate := TxDelegate{ValidatorAddress: d.sender, Amount: amount.String(), Source: utils.Cube}
	d.delegate(txDelegate)

	candidate = GetCandidateByPubKey(pubKey) // candidate object was modified by the delegation operation.
	//cds := &CandidateDailyStake{CandidateId: candidate.Id, Amount: candidate.Shares, BlockHeight: d.ctx.BlockHeight()}
	//SaveCandidateDailyStake(cds)
	candidate.PendingVotingPower = candidate.CalcVotingPower(d.ctx.BlockHeight())
	updateCandidate(candidate)
	return nil
}

func (d deliver) declareGenesisCandidacy(tx TxDeclareCandidacy, val types.GenesisValidator) error {
	// create and save the empty candidate
	pubKey, err := types.GetPubKey(tx.PubKey)
	if err != nil {
		return err
	}

	power, _ := strconv.ParseInt(val.Power, 10, 64)
	candidate := &Candidate{
		PubKey:       pubKey,
		OwnerAddress: d.sender.String(),
		Shares:       "0",
		VotingPower:  power,
		MaxShares:    tx.MaxAmount,
		CompRate:     tx.CompRate,
		CreatedAt:    0,
		Description:  tx.Description,
		Verified:     "N",
		Active:       "Y",
		BlockHeight:  d.ctx.BlockHeight(),
		State:        "Validator",
	}
	SaveCandidate(candidate)

	// delegate a part of the max staked CMT amount
	amount := sdk.NewInt(val.Shares).Mul(sdk.E18Int).String()
	txDelegate := TxDelegate{ValidatorAddress: d.sender, Amount: amount, Source: utils.Cube}
	d.delegate(txDelegate)

	candidate = GetCandidateByPubKey(pubKey) // candidate object was modified by the delegation operation.
	//cds := &CandidateDailyStake{CandidateId: candidate.Id, Amount: candidate.Shares, BlockHeight: d.ctx.BlockHeight()}
	//SaveCandidateDailyStake(cds)
	candidate.PendingVotingPower = candidate.CalcVotingPower(d.ctx.BlockHeight())
	updateCandidate(candidate)
	return nil
}

func (d deliver) updateCandidacy(tx TxUpdateCandidacy, gasFee sdk.Int) error {
	// create and save the empty candidate
	candidate := GetCandidateByAddress(d.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	totalCost := gasFee
	// If the max amount of CMTs is updated, the 10% of self-staking will be re-computed,
	// and the different will be charged
	if tx.MaxAmount != "" {
		maxAmount, _ := sdk.NewIntFromString(tx.MaxAmount)

		if !candidate.ParseMaxShares().Equal(maxAmount) {
			rechargeAmount := getRechargeAmount(maxAmount, candidate, d.params.SelfStakingRatio)

			if rechargeAmount.Cmp(big.NewInt(0)) > 0 {
				// charge
				totalCost = totalCost.Add(rechargeAmount)
				candidate.AddShares(rechargeAmount)

				// update delegation
				delegation := GetDelegation(d.sender, candidate.Id)
				delegation.AddDelegateAmount(rechargeAmount)
				UpdateDelegation(delegation)

				delegateHistory := &DelegateHistory{DelegatorAddress: d.sender, CandidateId: candidate.Id, Amount: rechargeAmount, OpCode: "recharge", BlockHeight: d.ctx.BlockHeight()}
				saveDelegateHistory(delegateHistory)
			}

			candidate.MaxShares = maxAmount.String()
		}
	}

	// If other information was updated, set the verified status to false
	if len(tx.Description.Name) > 0 {
		candidate.Verified = "N"
		candidate.Description.Name = tx.Description.Name
	}
	if len(tx.Description.Email) > 0 {
		candidate.Verified = "N"
		candidate.Description.Email = tx.Description.Email
	}
	if len(tx.Description.Website) > 0 {
		candidate.Verified = "N"
		candidate.Description.Website = tx.Description.Website
	}
	if len(tx.Description.Location) > 0 {
		candidate.Verified = "N"
		candidate.Description.Location = tx.Description.Location
	}
	if len(tx.Description.Profile) > 0 {
		candidate.Verified = "N"
		candidate.Description.Profile = tx.Description.Profile
	}

	// check if the delegator has sufficient funds
	if err := checkBalance(d.ctx.EthappState(), d.sender, totalCost); err != nil {
		return err
	}

	if !utils.IsBlank(tx.PubKey) {
		newPk, _ := types.GetPubKey(tx.PubKey)

		// save the previous pubkey which will be used to update validator set
		var updates PubKeyUpdates
		tuple := PubKeyUpdate{candidate.PubKey, newPk, candidate.VotingPower}
		b := d.store.Get(utils.PubKeyUpdatesKey)
		if b == nil {
			updates = PubKeyUpdates{tuple}
		} else {
			json.Unmarshal(b, &updates)
			updates = append(updates, tuple)
		}
		b, err := json.Marshal(updates)
		if err != nil {
			panic(err)
		}

		d.store.Set(utils.PubKeyUpdatesKey, b)
	}

	if !tx.CompRate.IsNil() && !sdk.ZeroRat.Equal(tx.CompRate) {
		candidate.CompRate = tx.CompRate
	}

	//commons.Transfer(d.sender, utils.HoldAccount, totalCost)
	d.ctx.EthappState().SubBalance(d.sender, totalCost.Int)
	d.ctx.EthappState().AddBalance(utils.HoldAccount, totalCost.Int)
	updateCandidate(candidate)
	return nil
}

func (d deliver) withdrawCandidacy(tx TxWithdrawCandidacy) error {
	// create and save the empty candidate
	validatorAddress := d.sender
	candidate := GetCandidateByAddress(validatorAddress)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	// All staked tokens will be distributed back to delegator addresses.
	// Self-staked CMTs will be refunded back to the validator address.
	delegations := GetDelegationsByCandidate(candidate.Id, "Y")
	for _, delegation := range delegations {
		if delegation.CompletelyWithdraw == "Y" {
			continue
		}

		txWithdraw := TxWithdraw{ValidatorAddress: validatorAddress, Amount: delegation.Shares().String()}
		d.doWithdraw(delegation, delegation.Shares(), candidate, txWithdraw)
	}

	candidate.Shares = "0"
	candidate.NumOfDelegators = 0
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
	updateCandidate(candidate)
	return nil
}

func (d deliver) activateCandidacy(tx TxActivateCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(d.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	candidate.Active = "Y"
	updateCandidate(candidate)
	return nil
}

func (d deliver) deactivateCandidacy(tx TxDeactivateCandidacy) error {
	// check to see if the address has been registered before
	candidate := GetCandidateByAddress(d.sender)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	candidate.Active = "N"
	updateCandidate(candidate)
	return nil
}

func (d deliver) delegate(tx TxDelegate) error {
	// Get the pubKey bond account
	candidate := GetCandidateByAddress(tx.ValidatorAddress)

	delegateAmount, ok := sdk.NewIntFromString(tx.Amount)
	if !ok || delegateAmount.LT(sdk.ZeroInt) {
		return ErrBadAmount()
	}

	//err := checkBalance(d.ctx.EthappState(), d.sender, delegateAmount)
	//if err != nil {
	//	return err
	//}

	// Move coins from the delegator account to the pubKey lock account
	if d.ctx.BlockHeight() == 0 && d.ctx.EthappState() == nil {
		// call from declareGenesisCandidacy
		err := commons.Transfer(d.sender, utils.HoldAccount, delegateAmount)
		if err != nil {
			return err
		}
	} else {
		// normal tx
		d.ctx.EthappState().SubBalance(d.sender, delegateAmount.Int)
		d.ctx.EthappState().AddBalance(utils.HoldAccount, delegateAmount.Int)
	}

	var source string
	if tx.Source == utils.Cube {
		source = utils.Cube
	} else if tx.Sig == "" {
		source = utils.CmtWallet
	} else {
		source = utils.Cube
	}

	// create or update delegation
	now := d.ctx.BlockTime()
	delegation := GetDelegation(d.sender, candidate.Id)
	if delegation == nil {
		delegation = &Delegation{
			DelegatorAddress:      d.sender,
			PubKey:                candidate.PubKey,
			CandidateId:           candidate.Id,
			DelegateAmount:        tx.Amount,
			AwardAmount:           "0",
			WithdrawAmount:        "0",
			PendingWithdrawAmount: "0",
			SlashAmount:           "0",
			State:                 "Y",
			CompRate:              candidate.CompRate,
			BlockHeight:           d.ctx.BlockHeight(),
			CreatedAt:             now,
			Source:                source,
			CompletelyWithdraw:    "N",
		}
		SaveDelegation(delegation)
	} else {
		delegation.AddDelegateAmount(delegateAmount)
		delegation.State = "Y"
		delegation.Source = source
		UpdateDelegation(delegation)
	}

	// Add delegateAmount to candidate
	candidate.AddShares(delegateAmount)
	candidate.NumOfDelegators = GetNumOfDelegatorsByCandidate(candidate.Id)
	updateCandidate(candidate)

	delegateHistory := &DelegateHistory{DelegatorAddress: d.sender, CandidateId: candidate.Id, Amount: delegateAmount, OpCode: "delegate", BlockHeight: d.ctx.BlockHeight()}
	saveDelegateHistory(delegateHistory)
	return nil
}

func (d deliver) withdraw(tx TxWithdraw) error {
	// get pubKey candidate
	candidate := GetCandidateByAddress(tx.ValidatorAddress)
	if candidate == nil {
		return ErrBadValidatorAddr()
	}

	amount, ok := sdk.NewIntFromString(tx.Amount)
	if !ok {
		return ErrInvalidWithdrawalAmount()
	}

	delegation := GetDelegation(d.sender, candidate.Id)

	// candidates can't withdraw the reserved reservation fund
	if d.sender.String() == candidate.OwnerAddress {
		remained := delegation.Shares().Sub(amount)
		if remained.LT(candidate.SelfStakingAmount(d.params.SelfStakingRatio)) {
			return ErrCandidateWithdrawalDisallowed()
		}
	}

	d.doWithdraw(delegation, amount, candidate, tx)

	// deduct shares from the candidate
	candidate.AddShares(amount.Neg())
	updateCandidate(candidate)
	return nil
}

func (d deliver) doWithdraw(delegation *Delegation, amount sdk.Int, candidate *Candidate, tx TxWithdraw) {
	//delegation.ReduceAverageStakingDate(amount)
	delegation.AddPendingWithdrawAmount(amount)
	if tx.CompletelyWithdraw {
		delegation.CompletelyWithdraw = "Y"
	} else {
		delegation.CompletelyWithdraw = "N"
	}
	UpdateDelegation(delegation)

	// record the unstaking requests which will be processed in 7 days
	performedBlockHeight := d.ctx.BlockHeight() + int64(utils.GetParams().UnstakeWaitingPeriod)
	unstakeRequest := &UnstakeRequest{
		DelegatorAddress:     delegation.DelegatorAddress,
		InitiatedBlockHeight: d.ctx.BlockHeight(),
		CandidateId:          candidate.Id,
		PerformedBlockHeight: performedBlockHeight,
		Amount:               amount.String(),
		State:                "PENDING",
	}
	saveUnstakeRequest(unstakeRequest)

	delegateHistory := &DelegateHistory{DelegatorAddress: delegation.DelegatorAddress, CandidateId: candidate.Id, Amount: amount, OpCode: "withdraw", BlockHeight: d.ctx.BlockHeight()}
	saveDelegateHistory(delegateHistory)
	return
}

func (d deliver) setCompRate(tx TxSetCompRate, gasFee sdk.Int) error {
	candidate := GetCandidateByAddress(d.sender)
	delegation := GetDelegation(tx.DelegatorAddress, candidate.Id)
	if delegation == nil {
		return ErrDelegationNotExists()
	}

	// check if the candidate has sufficient funds
	if err := checkBalance(d.ctx.EthappState(), d.sender, gasFee); err != nil {
		return err
	}
	// only charge gas fee here
	//commons.Transfer(d.sender, utils.HoldAccount, gasFee)
	d.ctx.EthappState().SubBalance(d.sender, gasFee.Int)
	d.ctx.EthappState().AddBalance(utils.HoldAccount, gasFee.Int)

	delegation.CompRate = tx.CompRate
	UpdateDelegation(delegation)
	return nil
}

func (d deliver) updateCandidateAccount(tx TxUpdateCandidacyAccount, gasFee sdk.Int) (int64, error) {
	// check if the delegator has sufficient funds
	if err := checkBalance(d.ctx.EthappState(), d.sender, gasFee); err != nil {
		return 0, err
	}

	// only charge gas fee here
	//commons.Transfer(d.sender, utils.HoldAccount, gasFee)
	d.ctx.EthappState().SubBalance(d.sender, gasFee.Int)
	d.ctx.EthappState().AddBalance(utils.HoldAccount, gasFee.Int)

	candidate := GetCandidateByAddress(d.sender)
	req := &CandidateAccountUpdateRequest{
		CandidateId: candidate.Id,
		FromAddress: d.sender, ToAddress: tx.NewCandidateAddress,
		CreatedBlockHeight: d.ctx.BlockHeight(),
		State:              "PENDING",
	}
	id := saveCandidateAccountUpdateRequest(req)
	return id, nil
}

func (d deliver) acceptCandidateAccountUpdateRequest(tx TxAcceptCandidacyAccountUpdate, gasFee sdk.Int) error {
	req := getCandidateAccountUpdateRequestById(tx.AccountUpdateRequestId)
	if req == nil {
		return ErrBadRequest()
	}

	if req.ToAddress != d.sender || req.State != "PENDING" {
		return ErrBadRequest()
	}

	candidate := GetCandidateById(req.CandidateId)
	if candidate == nil {
		return ErrBadRequest()
	}

	// check if the candidate has sufficient funds
	if err := checkBalance(d.ctx.EthappState(), d.sender, gasFee); err != nil {
		return err
	}

	candidate.OwnerAddress = req.ToAddress.String()
	updateCandidate(candidate)

	// update the candidate's self-delegation
	delegation := GetDelegation(req.FromAddress, candidate.Id)
	if delegation == nil {
		return ErrBadRequest()
	}

	delegation.DelegatorAddress = req.ToAddress
	UpdateDelegation(delegation)

	// return coins to the original account
	//commons.Transfer(utils.HoldAccount, req.FromAddress, delegation.Shares())
	d.ctx.EthappState().SubBalance(utils.HoldAccount, delegation.Shares().Int)
	d.ctx.EthappState().AddBalance(req.FromAddress, delegation.Shares().Int)

	// lock coins from the new account
	//commons.Transfer(req.ToAddress, utils.HoldAccount, delegation.Shares().Add(gasFee))
	d.ctx.EthappState().SubBalance(req.ToAddress, delegation.Shares().Add(gasFee).Int)
	d.ctx.EthappState().AddBalance(utils.HoldAccount, delegation.Shares().Add(gasFee).Int)

	// mark the request as completed
	req.State = "COMPLETED"
	req.AcceptedBlockHeight = d.ctx.BlockHeight()
	updateCandidateAccountUpdateRequest(req)

	return nil
}

func HandlePendingUnstakeRequests(height int64) error {
	reqs := GetUnstakeRequests(height)
	for _, req := range reqs {
		amount, _ := sdk.NewIntFromString(req.Amount)
		candidate := GetCandidateById(req.CandidateId)
		if candidate == nil {
			continue
		}

		delegation := GetDelegation(req.DelegatorAddress, candidate.Id)
		if delegation == nil {
			continue
		}

		if delegation.CompletelyWithdraw == "Y" {
			shares := delegation.Shares()
			amount = amount.Add(shares)
			delegation.AddWithdrawAmount(amount)
			delegation.AddPendingWithdrawAmount(shares)
			delegation.AddPendingWithdrawAmount(amount.Neg())
			delegation.CompletelyWithdraw = "N"
		} else {
			delegation.AddWithdrawAmount(amount)
			delegation.AddPendingWithdrawAmount(amount.Neg())
		}

		UpdateDelegation(delegation)

		minStakingAmount := sdk.NewInt(utils.GetParams().MinStakingAmount).Mul(sdk.E18Int)
		if delegation.ProfitableShares().LT(minStakingAmount) {
			RemoveDelegation(delegation.Id)
			candidate.NumOfDelegators = GetNumOfDelegatorsByCandidate(candidate.Id)
			updateCandidate(candidate)
		}

		req.State = "COMPLETED"
		updateUnstakeRequest(req)

		// transfer coins back to account
		commons.Transfer(utils.HoldAccount, req.DelegatorAddress, amount)
	}

	return nil
}

func checkBalance(state *ethstat.StateDB, addr common.Address, amount sdk.Int) error {
	balance, err := commons.GetBalance(state, addr)
	if err != nil {
		return err
	}

	if balance.LT(amount) {
		return ErrInsufficientFunds()
	}

	return nil
}

func getRechargeAmount(maxAmount sdk.Int, candidate *Candidate, ssr sdk.Rat) (res sdk.Int) {
	tmp := maxAmount.MulRat(ssr)
	d := GetDelegation(common.HexToAddress(candidate.OwnerAddress), candidate.Id)
	res = tmp.Sub(d.Shares())
	return
}
