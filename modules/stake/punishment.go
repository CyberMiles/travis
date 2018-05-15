package stake

import (
	"fmt"
	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/CyberMiles/travis/types"
)

const (
	byzantine_deduction_ratio = 5 // deduction ratio %5 for byzantine validators
	absent_deduction_ratio    = 1 // deduction ratio %1 for those validators absent for up to 3 hours
)

func PunishByzantineValidator(pubKey types.PubKey) (err error) {
	return punish(pubKey, byzantine_deduction_ratio, "Byzantine validator")
}

func PunishAbsentValidator(pubKey types.PubKey) (err error) {
	return punish(pubKey, absent_deduction_ratio, "Absent for up to 3 hours")
}

func punish(pubKey types.PubKey, ratio int64, reason string) (err error) {
	totalDeduction := new(big.Int)
	v := GetCandidateByPubKey(types.PubKeyString(pubKey))
	if v == nil {
		return ErrNoCandidateForAddress()
	}

	v.Active = "N"
	v.UpdatedAt = utils.GetNow()
	updateCandidate(v)

	// Get all of the delegators(includes the validator itself)
	delegations := GetDelegationsByPubKey(v.PubKey)
	for _, delegation := range delegations {
		deduction := new(big.Int)
		deduction.Mul(delegation.DelegateAmount, big.NewInt(ratio))
		deduction.Div(deduction, big.NewInt(100))
		punishDelegator(delegation, common.HexToAddress(v.OwnerAddress), deduction)
		totalDeduction.Add(totalDeduction, deduction)
	}

	// Save punishment history
	punishHistory := &PunishHistory{PubKey: pubKey, DeductionRatio: ratio, Deduction: totalDeduction, Reason: reason, CreatedAt: utils.GetNow()}
	savePunishHistory(punishHistory)

	return
}

func punishDelegator(d *Delegation, validatorAddress common.Address, amount *big.Int) {
	fmt.Printf("punish delegator, address: %s, amount: %d\n", d.DelegatorAddress.String(), amount)

	commons.Transfer(d.DelegatorAddress, utils.MintAccount, amount)
	now := utils.GetNow()

	d.DelegateAmount.Sub(d.DelegateAmount, amount)
	d.UpdatedAt = now
	UpdateDelegation(d)

	// accumulate shares of the validator
	val := GetCandidateByAddress(validatorAddress)
	val.Shares.Sub(val.Shares, amount)
	val.UpdatedAt = now
	updateCandidate(val)
}
