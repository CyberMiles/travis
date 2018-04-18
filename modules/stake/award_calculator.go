package stake

import (
	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"math"
	"math/big"
)

type awardCalculator struct {
	height          int64
	validators      Validators
	transactionFees *big.Int
}

type validator struct {
	stake           *big.Int
	ownerAddress    common.Address
	stakePercentage float64
	delegators      []delegator
}

type delegator struct {
	address     common.Address
	slotId      string
	amount      *big.Int
	proposedRoi int64
}

const (
	inflationRate       = 8
	yearlyBlockNumber   = 365 * 24 * 3600 / 10
	basicMintableAmount = "1000000000000000000000000000"
)

func NewAwardCalculator(height int64, validators Validators, transactionFees *big.Int) *awardCalculator {
	return &awardCalculator{height, validators, transactionFees}
}

func (ac awardCalculator) getMintableAmount() *big.Int {
	val := new(big.Float)
	bma, _ := val.SetString(basicMintableAmount)
	year := ac.height / yearlyBlockNumber
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
	var totalStakes *big.Int

	for _, val := range ac.validators {
		var validator validator
		candidate := GetCandidateByPubKey(val.PubKey.KeyString())
		if candidate.Shares.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		validator.stake = candidate.Shares
		validator.ownerAddress = candidate.OwnerAddress
		totalStakes.Add(totalStakes, candidate.Shares)

		slots := GetSlotsByValidator(candidate.OwnerAddress)
		for _, slot := range slots {
			delegates := GetSlotDelegatesBySlot(slot.Id)
			for _, delegate := range delegates {
				delegator := delegator{}
				delegator.address = delegate.DelegatorAddress
				delegator.slotId = delegate.SlotId
				delegator.amount = delegate.Amount
				delegator.proposedRoi = slot.ProposedRoi
				delegators = append(delegators, delegator)
			}
		}
		validator.delegators = delegators
		validators = append(validators, validator)
	}

	for _, val := range validators {
		var x, y *big.Float
		x.SetString(val.stake.String())
		y.SetString(totalStakes.String())
		z, _ := new(big.Float).Quo(x, y).Float64()
		val.stakePercentage = z
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
	var x, y *big.Float
	x.SetString(ac.getTotalBlockAward().String())
	y.SetString(ac.transactionFees.String())
	percentage := big.NewFloat(val.stakePercentage)

	tmp := new(big.Float)
	tmp.Add(x, y)
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

	// add award to stake of the delegator
	slotDelegate := GetSlotDelegate(del.address, del.slotId)
	slotDelegate.Amount = new(big.Int).Add(slotDelegate.Amount, amount)
	slotDelegate.UpdatedAt = utils.GetNow()
	updateSlotDelegate(slotDelegate)
}
