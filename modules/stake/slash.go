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

func SlashByzantineValidator(pubKey types.PubKey, blockTime, blockHeight int64) (err error) {
	slashRatio := utils.GetParams().SlashRatio
	return slash(pubKey, "Byzantine validator", slashRatio, blockTime, blockHeight)
}

func SlashAbsentValidator(pubKey types.PubKey, absence *Absence, blockTime, blockHeight int64) (err error) {
	slashRatio := utils.GetParams().SlashRatio
	maxSlashBlocks := utils.GetParams().MaxSlashBlocks
	if absence.GetCount() <= maxSlashBlocks {
		err = slash(pubKey, "Absent validator", slashRatio, blockTime, blockHeight)
	}

	if absence.GetCount() == maxSlashBlocks {
		err = RemoveValidator(pubKey, blockTime, blockHeight)
	}
	return
}

func SlashBadProposer(pubKey types.PubKey, blockTime, blockHeight int64) (err error) {
	maxSlashBlocks := int64(utils.GetParams().MaxSlashBlocks)
	slashRatio := utils.GetParams().SlashRatio
	slashRatio = slashRatio.Mul(sdk.NewRat(maxSlashBlocks, 1))

	err = slash(pubKey, "Bad block proposer", slashRatio, blockTime, blockHeight)
	if err != nil {
		return err
	}

	err = RemoveValidator(pubKey, blockTime, blockHeight)
	return
}

func slash(pubKey types.PubKey, reason string, slashRatio sdk.Rat, blockTime, blockHeight int64) (err error) {
	totalDeduction := sdk.NewInt(0)
	v := GetCandidateByPubKey(pubKey)
	if v == nil {
		return ErrBadValidatorAddr()
	}

	if v.ParseShares().Cmp(big.NewInt(0)) <= 0 {
		return nil
	}

	// Get all of the delegators(includes the validator itself)
	delegations := GetDelegationsByCandidate(v.Id, "Y")
	slashAmount := sdk.ZeroInt
	for _, d := range delegations {
		if utils.GetParams().SlashEnabled {
			slashAmount = d.Shares().MulRat(slashRatio)
		}
		slashDelegator(d, common.HexToAddress(v.OwnerAddress), slashAmount)
		totalDeduction = totalDeduction.Add(slashAmount)
	}

	// Save slash history
	now := utils.FormatUnixTime(blockTime)
	slash := &Slash{CandidateId: v.Id, SlashRatio: slashRatio, SlashAmount: totalDeduction, Reason: reason, CreatedAt: now, BlockHeight: blockHeight}
	saveSlash(slash)

	return
}

func slashDelegator(d *Delegation, validatorAddress common.Address, amount sdk.Int) {
	//fmt.Printf("slash delegator, address: %s, amount: %d\n", d.DelegatorAddress.String(), amount)
	d.AddSlashAmount(amount)
	UpdateDelegation(d)

	// accumulate shares of the validator
	val := GetCandidateByAddress(validatorAddress)
	val.AddShares(amount.Neg())
	updateCandidate(val)
}

func RemoveValidator(pubKey types.PubKey, blockTime, blockHeight int64) (err error) {
	v := GetCandidateByPubKey(pubKey)
	if v == nil {
		return ErrBadValidatorAddr()
	}

	v.Active = "N"
	updateCandidate(v)

	// Save slash history
	now := utils.FormatUnixTime(blockTime)
	slash := &Slash{CandidateId: v.Id, SlashRatio: sdk.ZeroRat, SlashAmount: sdk.ZeroInt, Reason: "Absent for up to 12 consecutive blocks", CreatedAt: now, BlockHeight: blockHeight}
	saveSlash(slash)
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
