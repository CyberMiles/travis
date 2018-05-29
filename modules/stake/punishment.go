package stake

import (
	"fmt"
	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/go-crypto"
	"math/big"
)

const (
	byzantine_slashing_ratio = 5     // deduction ratio %5 for byzantine validators
	absent_slashing_ratio    = 0.001 // deduction ratio for the absent validators
	max_slashes_limit        = 12
)

type Absence struct {
	count           int16
	lastBlockHeight int64
}

func (a Absence) Accumulate() {
	a.count++
	a.lastBlockHeight++
}

func (a Absence) ReachMaxSlashesLimit() bool {
	return a.count >= max_slashes_limit
}

type AbsentValidators struct {
	Validators map[crypto.PubKey]*Absence
}

func NewAbsentValidators() *AbsentValidators {
	return &AbsentValidators{Validators: make(map[crypto.PubKey]*Absence)}
}

func (av AbsentValidators) Add(pk crypto.PubKey, height int64) {
	absence := av.Validators[pk]
	if absence == nil {
		absence = &Absence{count: 1, lastBlockHeight: height}
	} else {
		absence.Accumulate()
	}
	av.Validators[pk] = absence
}

func (av AbsentValidators) Remove(pk crypto.PubKey) {
	delete(av.Validators, pk)
}

func (av AbsentValidators) Clear(currentBlockHeight int64) {
	for k, v := range av.Validators {
		if v.lastBlockHeight != currentBlockHeight {
			delete(av.Validators, k)
		}
	}
}

func PunishByzantineValidator(pubKey crypto.PubKey) (err error) {
	return punish(pubKey, byzantine_slashing_ratio, "Byzantine validator")
}

func PunishAbsentValidator(pubKey crypto.PubKey, absence *Absence) (err error) {
	err = punish(pubKey, absent_slashing_ratio, "Absent")
	if absence.ReachMaxSlashesLimit() {
		err = RemoveAbsentValidator(pubKey)
	}
	return
}

func punish(pubKey crypto.PubKey, ratio float64, reason string) (err error) {
	totalDeduction := new(big.Int)
	v := GetCandidateByPubKey(utils.PubKeyString(pubKey))
	if v == nil {
		return ErrNoCandidateForAddress()
	}

	// Get all of the delegators(includes the validator itself)
	delegations := GetDelegationsByPubKey(v.PubKey)
	for _, delegation := range delegations {
		tmp := new(big.Float)
		x := new(big.Float).SetInt(delegation.ParseDelegateAmount())
		tmp.Mul(x, big.NewFloat(ratio))
		slash := new(big.Int)
		tmp.Int(slash)
		punishDelegator(delegation, v.OwnerAddress, slash)
		totalDeduction.Add(totalDeduction, slash)
	}

	// Save punishment history
	punishHistory := &PunishHistory{PubKey: pubKey, SlashingRatio: ratio, SlashAmount: totalDeduction, Reason: reason, CreatedAt: utils.GetNow()}
	savePunishHistory(punishHistory)

	return
}

func punishDelegator(d *Delegation, validatorAddress common.Address, amount *big.Int) {
	fmt.Printf("punish delegator, address: %s, amount: %d\n", d.DelegatorAddress.String(), amount)

	commons.Transfer(d.DelegatorAddress, utils.MintAccount, amount)
	now := utils.GetNow()

	neg := new(big.Int).Neg(amount)
	d.AddSlashAmount(neg)
	d.UpdatedAt = now
	UpdateDelegation(d)

	// accumulate shares of the validator
	val := GetCandidateByAddress(validatorAddress)
	val.AddShares(neg)
	val.UpdatedAt = now
	updateCandidate(val)
}

func RemoveAbsentValidator(pubKey crypto.PubKey) (err error) {
	v := GetCandidateByPubKey(pubKey.KeyString())
	if v == nil {
		return ErrNoCandidateForAddress()
	}

	v.Active = "N"
	v.UpdatedAt = utils.GetNow()
	updateCandidate(v)

	// Save punishment history
	punishHistory := &PunishHistory{PubKey: pubKey, SlashingRatio: 0, SlashAmount: big.NewInt(0), Reason: "Absent for up to 12 consecutive blocks", CreatedAt: utils.GetNow()}
	savePunishHistory(punishHistory)
	return
}
