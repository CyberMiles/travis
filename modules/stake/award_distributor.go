package stake

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/libs/log"
	"math"
	"math/big"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
)

const (
	yearlyBlockNumber   = 365 * 24 * 3600 / 10
	basicMintableAmount = "1000000000000000000000000000"
)

type validator struct {
	shares           int64
	ownerAddress     common.Address
	pubKey           types.PubKey
	delegators       []*delegator
	selfDelegator    *delegator
	totalVotingPower int64
	votingPower      int64
}

func (v *validator) computeTotalSharesPercentage(totalShares int64, redistribute bool) sdk.Rat {
	p := sdk.NewRat(v.shares, totalShares)
	threshold := utils.GetParams().ValidatorSizeThreshold
	if !redistribute && p.Cmp(threshold) > 0 {
		p = threshold
	}

	return p
}

//_______________________________________________________________________

type delegator struct {
	address  common.Address
	shares   int64
	compRate sdk.Rat
	V        int64
	S1       int64
	S2       int64
	T        int64
	N        int64
}

//_______________________________________________________________________

type awardDistributor struct {
	height           int64
	validators       Validators
	backupValidators Validators
	transactionFees  sdk.Int
	logger           log.Logger
}

func NewAwardDistributor(height int64, validators, backupValidators Validators, transactionFees sdk.Int, logger log.Logger) *awardDistributor {
	return &awardDistributor{height, validators, backupValidators, transactionFees, logger}
}

func (ad awardDistributor) getMintableAmount() (amount sdk.Int) {
	base, ok := sdk.NewIntFromString(basicMintableAmount)
	if !ok {
		return
	}

	year := ad.height / yearlyBlockNumber
	b, _ := utils.GetParams().InflationRate.Add(sdk.OneRat()).Float64()
	pow := math.Pow(b, float64(year))
	pow = utils.RoundFloat(pow, 2)
	r := sdk.NewRat(int64(pow*100), 100)
	amount = base.MulRat(r)
	ad.logger.Debug("getMintableAmount", "height", ad.height, "year", year, "amount", amount)
	return
}

func (ad awardDistributor) getBlockAward() (blockAward sdk.Int) {
	ybn := sdk.NewInt(yearlyBlockNumber)
	blockAward = ad.getMintableAmount().MulRat(utils.GetParams().InflationRate).Div(ybn)
	ad.logger.Debug("getBlockAward", "yearly_block_number", ybn, "total_block_award", blockAward)
	return
}

func (ad awardDistributor) Distribute() {
	// distribute to the validators
	normalizedValidators, totalValidatorsShares := ad.buildValidators(ad.validators)
	normalizedBackupValidators, totalBackupsShares := ad.buildValidators(ad.backupValidators)
	var rr, rs sdk.Rat
	if len(normalizedBackupValidators) > 0 && totalBackupsShares > 0 {
		rr = utils.GetParams().ValidatorsBlockAwardRatio
		rs = sdk.NewRat(totalValidatorsShares, totalValidatorsShares+totalBackupsShares)
	} else {
		rr = sdk.OneRat()
		rs = sdk.OneRat()
	}

	ad.distributeToValidators(normalizedValidators, totalValidatorsShares, ad.getBlockAwardAndTxFees(), rr, rs)

	// distribute to the backup validators
	if len(normalizedBackupValidators) > 0 && totalBackupsShares > 0 {
		rr = sdk.OneRat().Sub(utils.GetParams().ValidatorsBlockAwardRatio)
		rs = sdk.NewRat(totalBackupsShares, totalValidatorsShares+totalBackupsShares)
		ad.distributeToValidators(normalizedBackupValidators, totalBackupsShares, ad.getBlockAwardAndTxFees(), rr, rs)
	}

	commons.Transfer(utils.MintAccount, utils.HoldAccount, ad.getBlockAward().Mul(sdk.NewInt(utils.BlocksPerHour)).Int)
}

func (ad *awardDistributor) buildValidators(rawValidators Validators) (normalizedValidators []*validator, totalShares int64) {
	totalShares = 0

	for _, val := range rawValidators {
		var validator validator
		var delegators []*delegator
		candidate := GetCandidateByAddress(common.HexToAddress(val.OwnerAddress))
		if candidate.Shares == "0" {
			continue
		}

		validator.ownerAddress = common.HexToAddress(candidate.OwnerAddress)
		validator.pubKey = candidate.PubKey

		// Get all delegators
		delegations := GetDelegationsByPubKey(candidate.PubKey)
		for _, delegation := range delegations {
			// if the amount of staked CMTs is less than 1000, no awards will be distributed.
			minStakingAmount := new(big.Int).Mul(big.NewInt(utils.GetParams().MinStakingAmount), big.NewInt(1e18))
			if delegation.Shares().Cmp(minStakingAmount) < 0 {
				continue
			}

			delegator := delegator{}
			delegator.address = delegation.DelegatorAddress
			delegator.shares = sdk.NewIntFromBigInt(delegation.Shares()).Div(sdk.NewInt(1e18)).Int64()
			delegators = append(delegators, &delegator)

			tenDaysAgo, _ := utils.GetTimeBefore(10 * 24)
			ninetyDaysAgo, _ := utils.GetTimeBefore(90 * 24)
			m1 := GetCandidateDailyStakeMax(delegation.PubKey, tenDaysAgo)
			m2 := GetCandidateDailyStakeMax(delegation.PubKey, ninetyDaysAgo)
			s1, _ := sdk.NewIntFromString(m1)
			s2, _ := sdk.NewIntFromString(m2)
			delegator.S1 = s1.Int64()
			delegator.S2 = s2.Int64()
			delegator.compRate = sdk.NewRat(int64(delegation.ParseCompRate()*1000), 1000)

			t, _ := utils.Diff(delegation.CreatedAt)
			if t > 180 {
				t = 180
			}
			delegator.T = t
			delegator.N += 1
		}

		// calculator voting power for delegators
		for _, d := range delegators {
			d.V = calcVotingPowerForDelegator(d.S1, d.S2, d.T, d.N, d.shares)
			validator.votingPower += d.V
			validator.shares += d.shares
			totalShares += d.shares
		}

		validator.delegators = delegators
		normalizedValidators = append(normalizedValidators, &validator)
	}

	return
}

func (ad *awardDistributor) distributeToValidators(normalizedValidators []*validator, totalShares int64, totalAward sdk.Int, rr, rs sdk.Rat) {
	actualDistributed := sdk.NewInt(0)
	for _, val := range normalizedValidators {
		p := val.computeTotalSharesPercentage(totalShares, false)
		award := totalAward.MulRat(p)
		ad.doDistribute(val, award, rr, rs)
		actualDistributed.Add(award)
	}

	// If there is remaining doDistribute, doDistribute a second round based on stake amount.
	remaining := totalAward.Sub(actualDistributed)
	if remaining.Cmp(big.NewInt(0)) > 0 {
		ad.logger.Debug("there is remaining award, doDistribute a second round based on stake amount.", "remaining", remaining)
		for _, val := range normalizedValidators {
			val.computeTotalSharesPercentage(totalShares, true)
			ad.doDistribute(val, remaining, rr, rs)
		}
	}
}

func (ad *awardDistributor) doDistribute(val *validator, totalAward sdk.Int, rr, rs sdk.Rat) {
	var t int64
	t = 0

	// doDistribute to the delegators
	for _, d := range val.delegators {
		c := d.compRate
		a := sdk.OneRat().Sub(c)
		b := d.V * a.Num().Int64() / a.Denom().Int64()
		r := totalAward.MulRat(sdk.NewRat(b, val.votingPower).Mul(rr).Quo(rs))
		ad.awardToDelegator(d, val, r)
		t += d.V * c.Num().Int64() / c.Denom().Int64()
	}

	// validator
	r := totalAward.MulRat(sdk.NewRat(t, val.votingPower).Mul(rr).Quo(rs))
	ad.awardToValidator(val, r)
	return
}

func (ad awardDistributor) getBlockAwardAndTxFees() sdk.Int {
	return ad.getBlockAward().Mul(sdk.NewInt(utils.BlocksPerHour)).Add(ad.transactionFees)
}

func (ad awardDistributor) awardToValidator(v *validator, award sdk.Int) {
	// A validator is also a delegator
	d := delegator{address: v.ownerAddress}
	ad.awardToDelegator(&d, v, award)
}

func (ad awardDistributor) awardToDelegator(d *delegator, v *validator, award sdk.Int) {
	ad.logger.Debug("awardToDelegator", "delegator_address", d.address.String(), "award", award)
	now := utils.GetNow()

	// add doDistribute to stake of the delegator
	delegation := GetDelegation(d.address, v.pubKey)
	if delegation == nil {
		return
	}

	delegation.AddAwardAmount(award.Int)
	delegation.UpdatedAt = now
	UpdateDelegation(delegation)

	// accumulate shares of the validator
	val := GetCandidateByAddress(v.ownerAddress)
	val.AddShares(award.Int)
	val.UpdatedAt = now
	updateCandidate(val)
}

func calcVotingPowerForDelegator(s1, s2, t, n, s int64) int64 {
	one := sdk.OneRat()
	r1 := sdk.NewRat(s1, s2)
	r2 := sdk.NewRat(t, 180)
	r3 := sdk.NewRat(n, 10)
	r4 := sdk.NewRat(s, 1)

	r1 = r1.Mul(r1)
	r2 = r2.Add(one)
	r3 = one.Sub(one.Quo(r3.Add(one)))
	r3 = r3.Mul(r3)
	v, _ := r1.Mul(r3).Mul(r4).Float64()
	f2, _ := r2.Float64()
	f2 = utils.RoundFloat(f2, 2)
	return int64(v * math.Log2(f2))
}
