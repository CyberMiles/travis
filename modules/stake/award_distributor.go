package stake

import (
	"math"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
)

const (
	yearlyBlockNumber   = 365 * 24 * 3600 / 10
	basicMintableAmount = "1000000000000000000000000000"
)

type validator struct {
	shares           *big.Int
	ownerAddress     common.Address
	pubKey           types.PubKey
	delegators       []*delegator
	compRate         float64
	sharesPercentage *big.Float
	selfDelegator    *delegator
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
	y := new(big.Float).SetInt(v.shares)
	result := new(big.Float).Quo(x, y)
	return result
}

func (v *validator) computeTotalSharesPercentage(redistribute bool) {
	x := new(big.Float).SetInt(v.shares)
	y := new(big.Float).SetInt(v.totalShares)
	v.sharesPercentage = new(big.Float).Quo(x, y)
	v.exceedLimit = false

	slf, err := strconv.ParseFloat(utils.GetParams().ValidatorSizeThreshold, 64)
	if err != nil {
		panic(err)
	}
	threshold := big.NewFloat(slf)
	if !redistribute && v.sharesPercentage.Cmp(threshold) > 0 {
		v.sharesPercentage = threshold
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
	height           int64
	validators       Validators
	backupValidators Validators
	transactionFees  *big.Int
	logger           log.Logger
}

func NewAwardDistributor(height int64, validators, backupValidators Validators, transactionFees *big.Int, logger log.Logger) *awardDistributor {
	return &awardDistributor{height, validators, backupValidators, transactionFees, logger}
}

func (ad awardDistributor) getMintableAmount() (amount *big.Int) {
	amount = new(big.Int)
	base, ok := new(big.Float).SetString(basicMintableAmount)
	if !ok {
		return
	}

	year := ad.height / yearlyBlockNumber
	pow := big.NewFloat(math.Pow(float64(1+utils.GetParams().InflationRate/100), float64(year)))
	new(big.Float).Mul(base, pow).Int(amount)
	ad.logger.Debug("getMintableAmount", "height", ad.height, "year", year, "amount", amount)
	return
}

func (ad awardDistributor) getBlockAward() (blockAward *big.Int) {
	ybn := big.NewInt(yearlyBlockNumber)
	blockAward = new(big.Int)
	blockAward.Mul(ad.getMintableAmount(), big.NewInt(utils.GetParams().InflationRate))
	blockAward.Div(blockAward, big.NewInt(100))
	blockAward.Div(blockAward, ybn)
	ad.logger.Debug("getBlockAward", "yearly_block_number", ybn, "total_block_award", blockAward)
	return
}

func (ad awardDistributor) Distribute() {
	// distribute to the validators
	normalizedValidators, totalValidatorsShares := ad.buildValidators(ad.validators)
	normalizedBackupValidators, totalBackupsShares := ad.buildValidators(ad.backupValidators)
	totalAward := new(big.Int)
	if len(normalizedBackupValidators) > 0 {
		totalAward.Mul(ad.getBlockAwardAndTxFees(), big.NewInt(utils.GetParams().ValidatorsBlockAwardRatio))
		totalAward.Div(totalAward, big.NewInt(100))
	} else {
		totalAward = ad.getBlockAwardAndTxFees()
	}
	ad.distributeToValidators(normalizedValidators, totalValidatorsShares, totalAward)

	// distribute to the backup validators
	if len(normalizedBackupValidators) > 0 {
		totalAward = new(big.Int)
		totalAward.Mul(ad.getBlockAwardAndTxFees(), big.NewInt(100-utils.GetParams().ValidatorsBlockAwardRatio))
		totalAward.Div(totalAward, big.NewInt(100))
		ad.distributeToValidators(normalizedBackupValidators, totalBackupsShares, totalAward)
	}

	commons.Transfer(utils.MintAccount, utils.HoldAccount, ad.getBlockAward())
}

func (ad *awardDistributor) buildValidators(rawValidators Validators) (normalizedValidators []*validator, totalShares *big.Int) {
	totalShares = new(big.Int)

	for _, val := range rawValidators {
		var validator validator
		var delegators []*delegator
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
		skippedShares := big.NewInt(0)

		// Get all delegators
		delegations := GetDelegationsByPubKey(candidate.PubKey)
		for _, delegation := range delegations {
			// if the amount of staked CMTs is less than 1000, no awards will be distributed.
			//minStakingAmount := new(big.Int).Mul(big.NewInt(utils.GetParams().MinStakingAmount), big.NewInt(1e18))
			//if delegation.Shares().Cmp(minStakingAmount) < 0 {
			//	skippedShares.Add(skippedShares, delegation.Shares())
			//	continue
			//}

			delegator := delegator{}
			delegator.address = delegation.DelegatorAddress
			delegator.shares = delegation.Shares()

			if delegator.address == validator.ownerAddress {
				validator.selfDelegator = &delegator
			} else {
				delegators = append(delegators, &delegator)
			}
		}

		totalShares.Sub(totalShares, skippedShares)
		validator.shares.Sub(validator.shares, skippedShares)

		if validator.selfDelegator != nil {
			validator.delegators = delegators
			normalizedValidators = append(normalizedValidators, &validator)
		}
	}

	return
}

func (ad *awardDistributor) distributeToValidators(normalizedValidators []*validator, totalShares *big.Int, totalAward *big.Int) {
	actualDistributed := big.NewInt(0)
	for _, val := range normalizedValidators {
		val.totalShares = totalShares
		val.computeTotalSharesPercentage(false)
		actualAward := ad.doDistribute(val, totalAward)
		actualDistributed.Add(actualDistributed, actualAward)
	}

	// If there is remaining doDistribute, doDistribute a second round based on stake amount.
	remaining := new(big.Int).Sub(totalAward, actualDistributed)
	if remaining.Cmp(big.NewInt(0)) > 0 {
		ad.logger.Debug("there is remaining award, doDistribute a second round based on stake amount.", "remaining", remaining)
		for _, val := range normalizedValidators {
			val.computeTotalSharesPercentage(true)
			ad.doDistribute(val, remaining)
		}
	}
}

func (ad *awardDistributor) doDistribute(val *validator, totalAward *big.Int) (actualTotalAward *big.Int) {
	ad.logger.Debug("########## doDistribute begin ########")
	actualTotalAward = val.getTotalAwardForValidator(totalAward, ad)

	// doDistribute to the validator
	valAward := val.getAwardForValidatorSelf(actualTotalAward, ad)
	ad.awardToValidator(val, valAward)

	remainingAward := new(big.Int)
	remainingAward.Sub(actualTotalAward, valAward)

	ad.logger.Debug("doDistribute", "totalAward", totalAward, "actualTotalAward", actualTotalAward, "valAward", valAward, "remainingAward", remainingAward, "valSharesPercentage", val.sharesPercentage)

	// doDistribute to the delegators
	for _, delegator := range val.delegators {
		delegator.computeSharesPercentage(val)
		delegatorAward := delegator.getAwardForDelegator(val.totalShares, remainingAward, ad, val)
		ad.awardToDelegator(delegator, val, delegatorAward)
	}

	ad.logger.Debug("########## doDistribute end ########")

	return
}

func (ad awardDistributor) getBlockAwardAndTxFees() *big.Int {
	blockAward := new(big.Int)
	blockAward.Add(ad.getBlockAward(), ad.transactionFees)
	return blockAward
}

func (ad awardDistributor) awardToValidator(v *validator, award *big.Int) {
	// A validator is also a delegator
	d := delegator{address: v.ownerAddress}
	ad.awardToDelegator(&d, v, award)
}

func (ad awardDistributor) awardToDelegator(d *delegator, v *validator, award *big.Int) {
	ad.logger.Debug("awardToDelegator", "delegator_address", d.address.String(), "award", award)
	now := utils.GetNow()

	// add doDistribute to stake of the delegator
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
