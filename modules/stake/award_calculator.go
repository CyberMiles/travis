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
	sharesPercentage float64
	delegators       []delegator
	cut              float64
}

type delegator struct {
	address common.Address
	slotId  string
	shares  *big.Int
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
		x := new(big.Float)
		y := new(big.Float)
		x.SetString(val.shares.String())
		y.SetString(totalShares.String())
		z, _ := new(big.Float).Quo(x, y).Float64()
		val.sharesPercentage = z
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

func (ac awardCalculator) getValidatorBlockAward(val validator) *big.Int {
	x := new(big.Float)
	y := new(big.Float)
	x.SetString(ac.getTotalBlockAward().String())
	y.SetString(ac.transactionFees.String())
	percentage := big.NewFloat(val.sharesPercentage)

	tmp := new(big.Float)
	tmp.Add(x, y)
	tmp.Mul(tmp, percentage)

	result := new(big.Int)
	tmp.Int(result)
	return result
}

func (ac awardCalculator) getDelegatorAward(del delegator, val validator, blockAward *big.Int) *big.Int {
	tmp := new(big.Float)
	x, _ := new(big.Float).SetString(del.shares.String())
	y, _ := new(big.Float).SetString(val.shares.String())
	tmp.Quo(x, y)
	award, _ := new(big.Float).SetString(blockAward.String())
	tmp.Mul(tmp, award)
	tmp.Mul(tmp, big.NewFloat(val.cut))

	result := new(big.Int)
	tmp.Int(result)
	return result
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
