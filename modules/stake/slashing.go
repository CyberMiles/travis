package stake

import (
	"github.com/CyberMiles/travis/sdk/state"
	"github.com/tendermint/go-amino"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
)

var cdc = amino.NewCodec()

type Absence struct {
	Count           int16
	LastBlockHeight int64
}

func (a *Absence) Accumulate() {
	a.Count++
	a.LastBlockHeight++
}

func (a Absence) GetCount() int16 {
	return a.Count
}

type AbsentValidators struct {
	Validators map[string]*Absence
}

//func NewAbsentValidators() *AbsentValidators {
//	return &AbsentValidators{Validators: make(map[types.PubKey]*Absence)}
//}

func (av AbsentValidators) Add(pk types.PubKey, height int64) {
	pkStr := types.PubKeyString(pk)
	absence := av.Validators[pkStr]
	if absence == nil {
		absence = &Absence{Count: 1, LastBlockHeight: height}
	} else {
		absence.Accumulate()
	}
	av.Validators[pkStr] = absence
}

func (av AbsentValidators) Remove(pk types.PubKey) {
	pkStr := types.PubKeyString(pk)
	delete(av.Validators, pkStr)
}

func (av AbsentValidators) Clear(currentBlockHeight int64) {
	for k, v := range av.Validators {
		if v.LastBlockHeight != currentBlockHeight {
			delete(av.Validators, k)
		}
	}
}

func (av AbsentValidators) Contains(pk types.PubKey) bool {
	pkStr := types.PubKeyString(pk)
	if _, exists := av.Validators[pkStr]; exists {
		return true
	}
	return false
}

func SlashByzantineValidator(pubKey types.PubKey) (err error) {
	slashingRatio := utils.GetParams().SlashingRatio
	return slash(pubKey, "Byzantine validator", slashingRatio)
}

func SlashAbsentValidator(pubKey types.PubKey, absence *Absence) (err error) {
	slashingRatio := utils.GetParams().SlashingRatio
	maxSlashingBlocks := utils.GetParams().MaxSlashingBlocks
	if absence.GetCount() <= maxSlashingBlocks {
		err = slash(pubKey, "Absent validator", slashingRatio)
	}

	if absence.GetCount() == maxSlashingBlocks {
		err = RemoveValidator(pubKey)
	}
	return
}

func SlashBadProposer(pubKey types.PubKey) (err error) {
	maxSlashingBlocks := int64(utils.GetParams().MaxSlashingBlocks)
	slashingRatio := utils.GetParams().SlashingRatio
	slashingRatio = slashingRatio.Mul(sdk.NewRat(maxSlashingBlocks, 1))

	err = slash(pubKey, "Bad block proposer", slashingRatio)
	if err != nil {
		return err
	}

	err = RemoveValidator(pubKey)
	return
}

func slash(pubKey types.PubKey, reason string, slashingRatio sdk.Rat) (err error) {
	totalDeduction := sdk.NewInt(0)
	v := GetCandidateByPubKey(pubKey)
	if v == nil {
		return ErrBadValidatorAddr()
	}

	if v.ParseShares().Cmp(big.NewInt(0)) <= 0 {
		return nil
	}

	// Get all of the delegators(includes the validator itself)
	delegations := GetDelegationsByPubKey(v.PubKey, "Y")
	for _, d := range delegations {
		slash := d.Shares().MulRat(slashingRatio)
		slashDelegator(d, common.HexToAddress(v.OwnerAddress), slash)
		totalDeduction.Add(slash)
	}

	// Save punishment history
	punishHistory := &PunishHistory{PubKey: pubKey, SlashingRatio: slashingRatio, SlashAmount: totalDeduction, Reason: reason, CreatedAt: utils.GetNow()}
	savePunishHistory(punishHistory)

	return
}

func slashDelegator(d *Delegation, validatorAddress common.Address, amount sdk.Int) {
	//fmt.Printf("slash delegator, address: %s, amount: %d\n", d.DelegatorAddress.String(), amount)
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

func RemoveValidator(pubKey types.PubKey) (err error) {
	v := GetCandidateByPubKey(pubKey)
	if v == nil {
		return ErrBadValidatorAddr()
	}

	v.Active = "N"
	v.UpdatedAt = utils.GetNow()
	updateCandidate(v)

	// Save punishment history
	punishHistory := &PunishHistory{PubKey: pubKey, SlashingRatio: sdk.ZeroRat, SlashAmount: sdk.ZeroInt, Reason: "Absent for up to 12 consecutive blocks", CreatedAt: utils.GetNow()}
	savePunishHistory(punishHistory)
	return
}

func LoadAbsentValidators(store state.SimpleDB) *AbsentValidators {
	b := store.Get(utils.AbsentValidatorsKey)
	if b == nil {
		return &AbsentValidators{Validators: make(map[string]*Absence)}
	}

	absentValidators := new(AbsentValidators)
	err := cdc.UnmarshalJSON(b, absentValidators)
	if err != nil {
		panic(err) // This error should never occure big problem if does
	}

	return absentValidators
}

func SaveAbsentValidators(store state.SimpleDB, absentValidators *AbsentValidators) {
	if len(absentValidators.Validators) == 0 {
		return
	}

	b, err := cdc.MarshalJSON(AbsentValidators{Validators: absentValidators.Validators})
	if err != nil {
		panic(err)
	}
	store.Set(utils.AbsentValidatorsKey, b)
}
