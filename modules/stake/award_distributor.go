package stake

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tmlibs/log"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
)

const (
	inflationRate       = 8
	yearlyBlockNumber   = 365 * 24 * 3600 / 10
	basicMintableAmount = "1000000000000000000000000000"
	stakeLimit          = 0.12 // fixme the percentage should be configurable
)

type validator struct {
	shares           *big.Int
	ownerAddress     common.Address
	pubKey           types.PubKey
	delegators       []delegator
	compRate         float64
	sharesPercentage *big.Float
	selfDelegator    delegator
	exceedLimit      bool
	totalShares      *big.Int
}

func (v validator) getAwardForValidatorSelf(totalAward *big.Int, ac *awardDistributor) (award *big.Int) {
	award = new(big.Int)
	x := new(big.Int)
	z := new(big.Float).SetInt(totalAward)
	p := v.computeSelfSharesPercentage()
	z.Mul(z, p)
	z.Int(x)

	d := new(big.Int)
	d.Sub(totalAward, x)
	r := new(big.Float).SetFloat64(v.compRate)
	tmp := new(big.Float).SetInt(d)
	tmp.Mul(tmp, r)
	y := new(big.Int)
	tmp.Int(y)

	award.Add(x, y)
	return
}

func (v validator) getTotalAwardForValidator(totalAward *big.Int, ac *awardDistributor) (award *big.Int) {
	award = new(big.Int)
	z := new(big.Float).SetInt(totalAward)
	z.Mul(z, v.sharesPercentage)
	z.Int(award)
	return
}

func (v validator) computeSelfSharesPercentage() *big.Float {
	x := new(big.Float).SetInt(v.selfDelegator.shares)
	y := new(big.Float).SetInt(v.totalShares)
	result := new(big.Float).Quo(x, y)
	return result
}

func (v *validator) computeTotalSharesPercentage(redistribute bool) {
	x := new(big.Float).SetInt(v.shares)
	y := new(big.Float).SetInt(v.totalShares)
	v.sharesPercentage = new(big.Float).Quo(x, y)
	v.exceedLimit = false

	if !redistribute && v.sharesPercentage.Cmp(big.NewFloat(stakeLimit)) > 0 {
		v.sharesPercentage = big.NewFloat(stakeLimit)
		v.exceedLimit = true
	}
}

//_______________________________________________________________________

type delegator struct {
	address          common.Address
	shares           *big.Int
	sharesPercentage *big.Float
}

func (d *delegator) computeSharesPercentage(val *validator) {
	d.sharesPercentage = new(big.Float)
	x := new(big.Float).SetInt(d.shares) // shares of the delegator
	tmp := new(big.Int)
	tmp.Sub(val.shares, val.selfDelegator.shares)
	y := new(big.Float).SetInt(tmp) // total shares of the validator
	d.sharesPercentage.Quo(x, y)
}

func (d delegator) getAwardForDelegator(totalShares, totalAward *big.Int, ac *awardDistributor, val *validator) (award *big.Int) {
	award = new(big.Int)
	tmp := new(big.Float)
	ta := new(big.Float).SetInt(totalAward)
	tmp.Mul(ta, d.sharesPercentage)
	tmp.Int(award)
	return
}

//_______________________________________________________________________

type awardDistributor struct {
	height          int64
	validators      Validators
	transactionFees *big.Int
	logger          log.Logger
	absentValidators *AbsentValidators
}

func NewAwardDistributor(height int64, validators Validators, transactionFees *big.Int, absentValidators *AbsentValidators, logger log.Logger) *awardDistributor {
	return &awardDistributor{height, validators, transactionFees, logger, absentValidators}
}

func (ad awardDistributor) getMintableAmount() (amount *big.Int) {
	amount = new(big.Int)
	base, ok := new(big.Float).SetString(basicMintableAmount)
	if !ok {
		return
	}

	year := ad.height / yearlyBlockNumber
	pow := big.NewFloat(math.Pow(float64(1+inflationRate/100), float64(year)))
	new(big.Float).Mul(base, pow).Int(amount)
	ad.logger.Debug("getMintableAmount", "height", ad.height, "year", year, "amount", amount)
	return
}

func (ad awardDistributor) getBlockAward() (blockAward *big.Int) {
	ybn := big.NewInt(yearlyBlockNumber)
	blockAward = new(big.Int)
	blockAward.Mul(ad.getMintableAmount(), big.NewInt(inflationRate))
	blockAward.Div(blockAward, big.NewInt(100))
	blockAward.Div(blockAward, ybn)
	ad.logger.Debug("getBlockAward", "yearly_block_number", ybn, "total_block_award", blockAward)
	return
}

func (ad awardDistributor) DistributeAll() {
	var validators []*validator
	totalShares := new(big.Int)

	for _, val := range ad.validators {
		if ad.isAbsent(val) {
			ad.logger.Debug("The validator is absent, no award", "validator", val.OwnerAddress)
			continue
		}

		var validator validator
		var delegators []delegator
		candidate := GetCandidateByAddress(common.HexToAddress(val.OwnerAddress))
		if candidate.Shares == "0" {
			continue
		}

		shares := candidate.ParseShares()
		validator.shares = shares
		validator.ownerAddress = common.HexToAddress(candidate.OwnerAddress)
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
				validator.selfDelegator = delegator
			} else {
				delegators = append(delegators, delegator)
			}
		}
		validator.delegators = delegators
		validators = append(validators, &validator)
	}

	totalAward := ad.getBlockAwardAndTxFees()
	actualDistributed := big.NewInt(0)
	for _, val := range validators {
		val.totalShares = totalShares
		val.computeTotalSharesPercentage(false)
		actualAward := ad.distribute(val, totalAward)
		actualDistributed.Add(actualDistributed, actualAward)
	}

	// If there is remaining distribute, distribute a second round based on stake amount.
	remaining := new(big.Int).Sub(totalAward, actualDistributed)
	if remaining.Cmp(big.NewInt(0)) > 0 {
		ad.logger.Debug("there is remaining award, distribute a second round based on stake amount.", "remaining", remaining)
		for _, val := range validators {
			val.computeTotalSharesPercentage(true)
			ad.distribute(val, remaining)
		}
	}
}

func (ad *awardDistributor) distribute(val *validator, totalAward *big.Int) (actualTotalAward *big.Int) {
	ad.logger.Debug("########## distribute begin ########")
	actualTotalAward = val.getTotalAwardForValidator(totalAward, ad)

	// distribute to the validator
	valAward := val.getAwardForValidatorSelf(actualTotalAward, ad)
	ad.awardToValidator(val, valAward)

	remainingAward := new(big.Int)
	remainingAward.Sub(actualTotalAward, valAward)

	// distribute to the delegators
	for _, delegator := range val.delegators {
		delegator.computeSharesPercentage(val)
		delegatorAward := delegator.getAwardForDelegator(val.totalShares, remainingAward, ad, val)
		ad.awardToDelegator(delegator, val, delegatorAward)
	}

	ad.logger.Debug("########## distribute end ########")

	return
}

func (ad awardDistributor) getBlockAwardAndTxFees() *big.Int {
	blockAward := new(big.Int)
	blockAward.Add(ad.getBlockAward(), ad.transactionFees)
	return blockAward
}

func (ad awardDistributor) awardToValidator(v *validator, award *big.Int) {
	ad.logger.Debug("awardToValidator", "validator_address", v.ownerAddress.String(), "award", award)

	// validator is also a delegator
	d := delegator{address: v.ownerAddress}
	ad.awardToDelegator(d, v, award)
}

func (ad awardDistributor) awardToDelegator(d delegator, v *validator, award *big.Int) {
	ad.logger.Debug("awardToDelegator", "delegator_address", d.address.String(), "award", award)
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

func (ad awardDistributor) isAbsent(val Validator) bool {
	for k := range ad.absentValidators.Validators {
		if k == val.PubKey {
			return true
		}
	}

	return false
}