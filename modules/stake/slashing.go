package stake

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
)

type Absence struct {
	count           int16
	lastBlockHeight int64
}

func (a *Absence) Accumulate() {
	a.count++
	a.lastBlockHeight++
}

func (a Absence) GetCount() int16 {
	return a.count
}

type AbsentValidators struct {
	Validators map[types.PubKey]*Absence
}

func NewAbsentValidators() *AbsentValidators {
	return &AbsentValidators{Validators: make(map[types.PubKey]*Absence)}
}

func (av AbsentValidators) Add(pk types.PubKey, height int64) {
	absence := av.Validators[pk]
	if absence == nil {
		absence = &Absence{count: 1, lastBlockHeight: height}
	} else {
		absence.Accumulate()
	}
	av.Validators[pk] = absence
}

func (av AbsentValidators) Remove(pk types.PubKey) {
	delete(av.Validators, pk)
}

func (av AbsentValidators) Clear(currentBlockHeight int64) {
	for k, v := range av.Validators {
		if v.lastBlockHeight != currentBlockHeight {
			delete(av.Validators, k)
		}
	}
}

func PunishByzantineValidator(pubKey types.PubKey) (err error) {
	return punish(pubKey, "Byzantine validator")
}

func PunishAbsentValidator(pubKey types.PubKey, absence *Absence) (err error) {
	maxSlashingBlocks := utils.GetParams().MaxSlashingBlocks
	if absence.GetCount() <= maxSlashingBlocks {
		err = punish(pubKey, "Absent")
	}

	if absence.GetCount() == maxSlashingBlocks {
		err = RemoveAbsentValidator(pubKey)
	}
	return
}

func punish(pubKey types.PubKey, reason string) (err error) {
	totalSlashed := sdk.NewInt(0)
	v := GetCandidateByPubKey(types.PubKeyString(pubKey))
	if v == nil {
		return ErrNoCandidateForAddress()
	}

	if v.ParseShares().Cmp(big.NewInt(0)) <= 0 {
		return nil
	}

	// Get all of the delegators(includes the validator itself)
	delegations := GetDelegationsByPubKey(v.PubKey)
	slashingRatio := utils.GetParams().SlashingRatio
	for _, d := range delegations {
		slash := d.Shares().MulRat(slashingRatio)
		punishDelegator(d, common.HexToAddress(v.OwnerAddress), slash)
		totalSlashed = totalSlashed.Add(slash)
	}

	// Save punishment history
	punishHistory := &PunishHistory{PubKey: pubKey, SlashingRatio: slashingRatio, SlashAmount: totalSlashed, Reason: reason, CreatedAt: utils.GetNow()}
	savePunishHistory(punishHistory)

	return
}

func punishDelegator(d *Delegation, validatorAddress common.Address, amount sdk.Int) {
	//fmt.Printf("punish delegator, address: %s, amount: %d\n", d.DelegatorAddress.String(), amount)
	now := utils.GetNow()
	d.AddSlashAmount(amount)
	d.UpdatedAt = now
	UpdateDelegation(d)

	// accumulate shares of the validator
	val := GetCandidateByAddress(validatorAddress)
	val.AddShares(amount.Neg())
	val.UpdatedAt = now
	updateCandidate(val)
}

func RemoveAbsentValidator(pubKey types.PubKey) (err error) {
	v := GetCandidateByPubKey(types.PubKeyString(pubKey))
	if v == nil {
		return ErrNoCandidateForAddress()
	}

	v.Active = "N"
	v.UpdatedAt = utils.GetNow()
	updateCandidate(v)

	// Save punishment history
	punishHistory := &PunishHistory{PubKey: pubKey, SlashingRatio: sdk.ZeroRat, SlashAmount: sdk.ZeroInt, Reason: "Absent for up to 12 consecutive blocks", CreatedAt: utils.GetNow()}
	savePunishHistory(punishHistory)
	return
}
