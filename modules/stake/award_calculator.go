package stake

import (
	"fmt"
	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/go-crypto"
	"math"
	"math/big"
)

type awardCalculator struct {
	height          int64
	validators      Validators
	transactionFees *big.Int
}

type validator struct {
	shares             *big.Int
	ownerAddress       common.Address
	pubKey             crypto.PubKey
	delegators         []delegator
	compRate           float64
	sharesPercentage   *big.Float
	validatorDelegator delegator
	exceedLimit        bool
}

const (
	inflationRate       = 8
	yearlyBlockNumber   = 365 * 24 * 3600 / 10
	basicMintableAmount = "1000000000000000000000000000"
	stakeLimit          = 0.12 // fixme the percentage should be configurable
)

func (v validator) getAwardsForValidatorSelf(totalShares *big.Int, totalAwards *big.Int, ac *awardCalculator) (awards *big.Int) {
	x := new(big.Int)
	z := new(big.Float).SetInt(totalAwards)
	p := new(big.Float)
	if v.exceedLimit {
		p = v.sharesPercentage
	} else {
		p = v.computeSelfSharesPercentage(totalShares)
	}
	z.Mul(z, p)
	z.Int(x)

	t := v.getAwardsForValidator(totalAwards, ac)
	t.Sub(t, x)
	r := new(big.Float).SetFloat64(v.compRate)
	tmp := new(big.Float).SetInt(t)
	tmp.Mul(tmp, r)
	y := new(big.Int)
	tmp.Int(y)
	awards.Add(x, y)

	fmt.Printf("shares percentage: %v, awards for validator self: %v\n", v.sharesPercentage, awards)
	return
}

func (v validator) computeSelfSharesPercentage(totalShares *big.Int) *big.Float {
	x := new(big.Float).SetInt(v.validatorDelegator.shares)
	y := new(big.Float).SetInt(totalShares)
	result := new(big.Float).Quo(x, y)
	return result
}

func (v validator) getAwardsForValidator(totalAwards *big.Int, ac *awardCalculator) (awards *big.Int) {
	awards = new(big.Int)
	z := new(big.Float).SetInt(totalAwards)
	z.Mul(z, v.sharesPercentage)
	z.Int(awards)
	fmt.Printf("shares percentage: %v, awards for whole validator: %v\n", v.sharesPercentage, awards)
	return
}

func (v validator) computeTotalSharesPercentage(totalShares *big.Int, redistribute bool) {
	x := new(big.Float).SetInt(v.shares)
	y := new(big.Float).SetInt(totalShares)
	v.sharesPercentage = new(big.Float).Quo(x, y)
	v.exceedLimit = false

	if !redistribute && v.sharesPercentage.Cmp(big.NewFloat(stakeLimit)) > 0 {
		v.sharesPercentage = big.NewFloat(stakeLimit)
		v.exceedLimit = true
	}
}

type delegator struct {
	address          common.Address
	shares           *big.Int
	sharesPercentage *big.Float
}

func (d delegator) computeSharesPercentage(val validator) {
	d.sharesPercentage = new(big.Float)
	x := new(big.Float).SetInt(d.shares) // shares of the delegator
	tmp := new(big.Int)
	tmp.Sub(val.shares, val.validatorDelegator.shares)
	y := new(big.Float).SetInt(tmp) // total shares of the validator
	d.sharesPercentage.Quo(x, y)
	fmt.Printf("delegator shares: %f, validator shares: %f, percentage: %f\n", x, y, d.sharesPercentage)
}

func (d delegator) getAwardsForDelegator(totalShares *big.Int, totalAwards *big.Int, ac *awardCalculator, val validator) (awards *big.Int) {
	awards = new(big.Int)
	tmp := new(big.Float)
	d.computeSharesPercentage(val)
	ta := new(big.Float).SetInt(totalAwards)
	tmp.Mul(ta, d.sharesPercentage)
	r := big.NewFloat(1 - val.compRate)
	tmp.Mul(tmp, r)
	tmp.Int(awards)
	fmt.Printf("delegator awards: %d\n", awards)
	return
}

func NewAwardCalculator(height int64, validators Validators, transactionFees *big.Int) *awardCalculator {
	fmt.Printf("new award calculator, height: %d, transaction fees: %d\n", height, transactionFees)
	return &awardCalculator{height, validators, transactionFees}
}

func (ac awardCalculator) getMintableAmount() (result *big.Int) {
	result = new(big.Int)
	base, ok := new(big.Float).SetString(basicMintableAmount)
	if !ok {
		return
	}

	year := ac.height / yearlyBlockNumber
	pow := big.NewFloat(math.Pow(float64(1+inflationRate/100), float64(year)))
	new(big.Float).Mul(base, pow).Int(result)
	fmt.Printf("year: %d, mintable amount: %v\n", year, result)
	return
}

func (ac awardCalculator) getBlockAward() (result *big.Int) {
	blocks := big.NewInt(yearlyBlockNumber)
	result = new(big.Int)
	result.Mul(ac.getMintableAmount(), big.NewInt(inflationRate))
	result.Div(result, big.NewInt(100))
	result.Div(result, blocks)
	fmt.Printf("yearly block number: %d, total block award: %v\n", blocks, result)
	return
}

func (ac awardCalculator) DistributeAll() {
	var validators []validator
	totalShares := new(big.Int)

	for _, val := range ac.validators {
		var validator validator
		var delegators []delegator
		candidate := GetCandidateByAddress(val.OwnerAddress)
		if candidate.Shares == "0" {
			continue
		}

		shares := candidate.ParseShares()
		validator.shares = shares
		validator.ownerAddress = candidate.OwnerAddress
		validator.pubKey = candidate.PubKey
		validator.compRate = candidate.ParseCompRate()
		totalShares.Add(totalShares, shares)

		// Get all delegators
		delegations := GetDelegationsByPubKey(candidate.PubKey)
		for _, delegation := range delegations {
			delegator := delegator{}
			delegator.address = delegation.DelegatorAddress
			delegator.shares = delegation.Shares()

			if delegator.address == validator.ownerAddress {
				validator.validatorDelegator = delegator
			} else {
				delegators = append(delegators, delegator)
			}
		}
		validator.delegators = delegators
		validators = append(validators, validator)
	}

	totalAward := ac.getBlockAward()
	actualTotalAward := big.NewInt(0)
	for _, val := range validators {
		actualAward := distribute(val, totalShares, ac, big.NewInt(0))
		actualTotalAward.Add(actualTotalAward, actualAward)
	}

	// If there is remaining distribute, distribute a second round based on stake amount.
	remaining := new(big.Int).Sub(totalAward, actualTotalAward)
	if remaining.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("there is remaining award, distribute a second round based on stake amount. remaining: %d\v", remaining)
		for _, val := range validators {
			distribute(val, totalShares, ac, remaining)
		}
	}
}

func distribute(val validator, totalShares *big.Int, ac awardCalculator, remaining *big.Int) (actualAward *big.Int) {
	again := !(remaining.Cmp(big.NewInt(0)) == 0)
	x := new(big.Float).SetInt(val.shares)
	y := new(big.Float).SetInt(totalShares)
	val.sharesPercentage = new(big.Float).Quo(x, y)
	val.exceedLimit = false

	if !again && val.sharesPercentage.Cmp(big.NewFloat(0.12)) > 0 {
		val.sharesPercentage = big.NewFloat(0.12)
		val.exceedLimit = true
	}

	fmt.Printf("val.shares: %f, totalShares: %f, percentage: %f\n", x, y, val.sharesPercentage)

	if again {
		actualAward = ac.getRemainingAwardForValidator(val, remaining)
	} else {
		actualAward = ac.getBlockAwardForWholeValidator(val)
	}

	// distribute to validator

	remainingAward := actualAward

	// distribute to delegators
	for _, delegator := range val.delegators {
		delegatorAward := ac.getDelegatorAward(delegator, val, actualAward)
		remainingAward.Sub(remainingAward, delegatorAward)
		ac.awardToDelegator(delegator, val, delegatorAward)
	}
	ac.awardToValidator(val, remainingAward)

	return
}

func (ac awardCalculator) getBlockAwardForValidatorSelf(val validator) (result *big.Int) {
	blockAward := ac.getBlockAwardAndTxFees()
	return ac.getAwardForValidator(val, blockAward)
}

func (ac awardCalculator) getAwardForValidatorSelf(val validator, award *big.Int) (result *big.Int) {
	result = new(big.Int)
	z := new(big.Float).SetInt(award)
	z.Mul(z, val.sharesPercentage)
	z.Int(result)
	fmt.Printf("shares percentage: %v, award for validator: %v\n", val.sharesPercentage, result)
	return
}

func (ac awardCalculator) getBlockAwardAndTxFees() *big.Int {
	blockAward := new(big.Int)
	blockAward.Add(ac.getBlockAward(), ac.transactionFees)
	return blockAward
}

func (ac awardCalculator) getBlockAwardForWholeValidator(val validator) (result *big.Int) {
	blockAward := new(big.Int)
	blockAward.Add(ac.getBlockAward(), ac.transactionFees)
	return ac.getAwardForValidator(val, blockAward)
}

func (ac awardCalculator) getRemainingAwardForValidator(val validator, remaining *big.Int) (result *big.Int) {
	return ac.getAwardForValidator(val, remaining)
}

func (ac awardCalculator) getAwardForValidator(val validator, award *big.Int) (result *big.Int) {
	result = new(big.Int)
	z := new(big.Float).SetInt(award)
	z.Mul(z, val.sharesPercentage)
	z.Int(result)
	fmt.Printf("shares percentage: %v, award for validator: %v\n", val.sharesPercentage, result)
	return
}

func (ac awardCalculator) getDelegatorAward(del delegator, val validator, blockAward *big.Int) (result *big.Int) {
	result = new(big.Int)
	z := new(big.Float)
	x := new(big.Float).SetInt(del.shares) // shares of the delegator
	y := new(big.Float).SetInt(val.shares) // total shares of the validator
	z.Quo(x, y)
	fmt.Printf("delegator shares: %f, validator shares: %f, percentage: %f\n", x, y, z)
	award := new(big.Float).SetInt(blockAward)
	z.Mul(z, award)
	cut := big.NewFloat(val.compRate)
	z.Mul(z, cut)
	z.Int(result)
	fmt.Printf("delegator award: %d\n", result)
	return
}

func (ac awardCalculator) awardToValidator(v validator, award *big.Int) {
	fmt.Printf("award to validator, owner_address: %s, award: %d\n", v.ownerAddress.String(), award)

	// validator is also a delegator
	d := delegator{address: v.ownerAddress}
	ac.awardToDelegator(d, v, award)
}

func (ac awardCalculator) awardToDelegator(d delegator, v validator, award *big.Int) {
	fmt.Printf("award to delegator, address: %s, amount: %d\n", d.address.String(), award)
	commons.Transfer(utils.MintAccount, utils.HoldAccount, award)
	now := utils.GetNow()

	// add distribute to stake of the delegator
	delegation := GetDelegation(d.address, v.pubKey)
	if delegation == nil {
		return
	}

	delegation.AddAwardAmount(award)
	delegation.UpdatedAt = now
	UpdateDelegation(delegation)

	// accumulate shares of the validator
	val := GetCandidateByAddress(v.ownerAddress)
	val.AddShares(award)
	val.UpdatedAt = now
	updateCandidate(val)
}
