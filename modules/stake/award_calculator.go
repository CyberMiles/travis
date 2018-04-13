package stake

import (
	"github.com/tendermint/go-crypto"
	"math"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/CyberMiles/travis/commons"
)

type awardCalculator struct {
	height          int64
	validators      []crypto.PubKey
	transactionFees *big.Int
}

type validator struct {
	stake uint64
	ownerAddress common.Address
	stakePercentage float64
	delegators []delegator
}

type delegator struct {
	address common.Address
	slotId string
	amount *big.Int
	proposedRoi int64
}

const (
	inflationRate     = 8
	yearlyBlockNumber = 365 * 24 * 3600 / 10
	basicMintAmount   = "1000000000000000000000000000"
)

func NewAwardCalculator(height int64, validators []crypto.PubKey, transacationFees *big.Int) *awardCalculator {
	return &awardCalculator{height, validators, transacationFees}
}

func (ac awardCalculator) getMintableAmount() *big.Int {
	val := new(big.Float)
	bma, _ := val.SetString(basicMintAmount)
	year := ac.height/yearlyBlockNumber
	pow := big.NewFloat(math.Pow(float64(1+inflationRate/100), float64(year)))
	result := new(big.Int)
	z := new(big.Float)
	z.Mul(bma, pow).Int(result)
	return result
}

func (ac awardCalculator) getTotalBlockAward() *big.Int {
	val := big.NewInt(yearlyBlockNumber)
	tmp := new(big.Int)
	tmp.Mul(ac.getMintableAmount(), big.NewInt(inflationRate))
	tmp.Div(tmp, big.NewInt(100))
	tmp.Div(tmp, val)
	return tmp
}

func (ac awardCalculator) AwardAll() {
	var validators []validator
	var delegators []delegator
	var totalStakes uint64

	for _, pk := range ac.validators {
		var validator validator
		candidate := GetCandidateByPubKey(pk.KeyString())
		if candidate.Shares == 0 {
			continue
		}

		validator.stake = candidate.Shares
		validator.ownerAddress = candidate.OwnerAddress
		totalStakes += candidate.Shares

		slots := GetSlotsByValidator(candidate.OwnerAddress)
		for _, slot := range slots {
			delegates := GetSlotDelegatesBySlot(slot.Id)
			for _, delegate := range delegates {
				delegator := delegator{}
				delegator.address = delegate.DelegatorAddress
				delegator.slotId = delegate.SlotId
				delegator.amount = big.NewInt(delegate.Amount)
				delegator.proposedRoi = slot.ProposedRoi
				delegators = append(delegators, delegator)
			}
		}
		validator.delegators = delegators
		validators = append(validators, validator)
	}

	for _, val := range validators {
		val.stakePercentage = float64(val.stake) / float64(totalStakes)
		remainedAward := ac.getValidatorBlockAward(val)

		// award to delegators
		for _, delegator := range val.delegators {
			delegatorAward := ac.getDelegatorAward(delegator)
			remainedAward.Sub(remainedAward, delegatorAward)
			ac.awardToDelegator(delegator, delegatorAward)
		}
		ac.awardToValidator(val, remainedAward)
	}
}

func (ac awardCalculator) getValidatorBlockAward(val validator) *big.Int {
	var l, r *big.Float
	l.SetString(ac.getTotalBlockAward().String())
	r.SetString(ac.transactionFees.String())
	percentage := big.NewFloat(val.stakePercentage)

	tmp := new(big.Float)
	tmp.Add(l, r)
	tmp.Mul(tmp, percentage)

	result := new(big.Int)
	tmp.Int(result)
	return result
}

func (ac awardCalculator) getDelegatorAward(del delegator) *big.Int {
	result := new(big.Int)
	result.Mul(del.amount, big.NewInt(del.proposedRoi))
	result.Div(result, big.NewInt(100))
	result.Div(result, big.NewInt(yearlyBlockNumber))
	return result
}

func (ac awardCalculator) awardToValidator(val validator, amount *big.Int) {
	commons.Transfer(DefaultHoldAccount, val.ownerAddress, amount)
}

func (ac awardCalculator) awardToDelegator(del delegator, amount *big.Int) {
	commons.Transfer(DefaultHoldAccount, del.address, amount)

	// todo add award to stake of the delegator

}
