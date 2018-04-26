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
	shares           *big.Int
	ownerAddress     common.Address
	delegators       []delegator
	cut              int64
	sharesPercentage *big.Float
}

type delegator struct {
	address common.Address
	slotId  string
	shares  *big.Int
}

const (
	inflationRate      = 8
	yearlyBlockNumber  = 365 * 24 * 3600 / 10
	basicMinableAmount = "1000000000000000000000000000"
)

func NewAwardCalculator(height int64, validators Validators, transactionFees *big.Int) *awardCalculator {
	return &awardCalculator{height, validators, transactionFees}
}

func (ac awardCalculator) getMinableAmount() (result *big.Int) {
	result = new(big.Int)
	base, ok := new(big.Float).SetString(basicMinableAmount)
	if !ok {
		// should never run into this block
		panic(ErrBadAmount())
	}

	year := ac.height / yearlyBlockNumber
	pow := big.NewFloat(math.Pow(float64(1+inflationRate/100), float64(year)))
	new(big.Float).Mul(base, pow).Int(result)
	return
}

func (ac awardCalculator) getTotalBlockAward() (result *big.Int) {
	blocks := big.NewInt(yearlyBlockNumber)
	result = new(big.Int)
	result.Mul(ac.getMinableAmount(), big.NewInt(inflationRate))
	result.Div(result, big.NewInt(100))
	result.Div(result, blocks)
	return
}

func (ac awardCalculator) AwardAll() {
	var validators []validator
	var delegators []delegator
	totalShares := new(big.Int)

	for _, val := range ac.validators {
		var validator validator
		candidate := GetCandidateByPubKey(val.PubKey.KeyString())
		if candidate.Shares.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		validator.shares = candidate.Shares
		validator.ownerAddress = candidate.OwnerAddress
		validator.cut = candidate.Cut
		totalShares.Add(totalShares, candidate.Shares)

		// Get all of the delegators
		delegations := GetDelegationsByCandidate(candidate.OwnerAddress)
		for _, delegation := range delegations {
			delegator := delegator{}
			delegator.address = delegation.DelegatorAddress
			delegator.shares = delegation.Shares
			delegators = append(delegators, delegator)
		}
		validator.delegators = delegators
		validators = append(validators, validator)
	}

	for _, val := range validators {
		x := new(big.Float).SetInt(val.shares)
		y := new(big.Float).SetInt(totalShares)
		val.sharesPercentage = new(big.Float).Quo(x, y)
		blockAward := ac.getValidatorBlockAward(val)
		remainedAward := blockAward

		// award to delegators
		for _, delegator := range val.delegators {
			delegatorAward := ac.getDelegatorAward(delegator, val, blockAward)
			remainedAward.Sub(remainedAward, delegatorAward)
			ac.awardToDelegator(delegator, val, delegatorAward)
		}
		ac.awardToValidator(val, remainedAward)
	}
}

func (ac awardCalculator) getValidatorBlockAward(val validator) (result *big.Int) {
	result = new(big.Int)
	z := new(big.Float)
	x := new(big.Float).SetInt(ac.getTotalBlockAward())
	y := new(big.Float).SetInt(ac.transactionFees)
	z.Add(x, y)
	z.Mul(z, val.sharesPercentage)
	z.Int(result)
	return
}

func (ac awardCalculator) getDelegatorAward(del delegator, val validator, blockAward *big.Int) (result *big.Int) {
	result = new(big.Int)
	z := new(big.Float)
	x := new(big.Float).SetInt(del.shares) // shares of the delegator
	y := new(big.Float).SetInt(val.shares) // total shares of the validator
	z.Quo(x, y)
	award := new(big.Float).SetInt(blockAward)
	z.Mul(z, award)
	cut := new(big.Float).SetInt64(val.cut)
	z.Mul(z, cut) // format: 123 -> 0.0123 -> 1.23%
	z.Quo(z, new(big.Float).SetInt64(10000))
	z.Int(result)
	return
}

func (ac awardCalculator) awardToValidator(val validator, amount *big.Int) {
	commons.Transfer(DefaultHoldAccount, val.ownerAddress, amount)
}

func (ac awardCalculator) awardToDelegator(del delegator, val validator, award *big.Int) {
	commons.Transfer(DefaultHoldAccount, del.address, award)

	// add award to stake of the delegator
	delegation := GetDelegation(del.address, val.ownerAddress)
	delegation.Shares.Add(delegation.Shares, award)
	delegation.UpdatedAt = utils.GetNow()
	UpdateDelegation(delegation)
}
