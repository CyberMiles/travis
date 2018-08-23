package stake

import (
	"fmt"
	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/libs/log"
	"math"
)

const (
	yearlyBlockNumber   = 365 * 24 * 3600 / 10
	basicMintableAmount = "1000000000000000000000000000"
	HalfYear            = 180
)

type validator struct {
	s            int64
	ownerAddress common.Address
	pk           types.PubKey
	delegators   []*delegator
	vp           int64
}

func (v *validator) computeTotalSharesPercentage(totalShares int64, redistribute bool) sdk.Rat {
	p := sdk.NewRat(v.s, totalShares)
	threshold := utils.GetParams().ValidatorSizeThreshold
	if !redistribute && p.Cmp(threshold) > 0 {
		p = threshold
	}

	return p
}

func (v validator) String() string {
	return fmt.Sprintf("[validator] ownerAddress: %s, delegators: %d, vp: %d", v.ownerAddress.String(), len(v.delegators), v.vp)
}

//_______________________________________________________________________

type delegator struct {
	address common.Address
	s       int64
	c       sdk.Rat
	vp      int64
	s1      int64
	s2      int64
	t       int64
	n       int64
}

func (d delegator) String() string {
	return fmt.Sprintf("[deligator] address: %s, s: %d, c: %vp, vp: %d, s1: %d, s2: %d, t: %d, n: %d", d.address.String(), d.s, d.c, d.vp, d.s1, d.s2, d.t, d.n)
}

//_______________________________________________________________________

type awardDistributor struct {
	height           int64
	validators       Validators
	backupValidators Validators
	transactionFees  sdk.Int
	logger           log.Logger
}

func NewAwardDistributor(height int64, validators, backupValidators Validators, logger log.Logger) *awardDistributor {
	return &awardDistributor{height, validators, backupValidators, sdk.NewIntFromBigInt(utils.BlockGasFee), logger}
}

func (ad awardDistributor) getMintableAmount() (amount sdk.Int) {
	base, ok := sdk.NewIntFromString(basicMintableAmount)
	if !ok {
		return
	}

	year := ad.height / yearlyBlockNumber
	b, _ := utils.GetParams().InflationRate.Add(sdk.OneRat).Float64()
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
	vals, totalValShares, totalValVotingPower := ad.calcVotingPower(ad.validators)
	backups, totalBackupShares, totalBackupVotingPower := ad.calcVotingPower(ad.backupValidators)
	totalVotingPower := totalValVotingPower + totalBackupVotingPower
	var rr, rs sdk.Rat
	if len(backups) > 0 && totalBackupShares > 0 {
		rr = utils.GetParams().ValidatorsBlockAwardRatio
		rs = sdk.NewRat(totalValShares, totalValShares+totalBackupShares)
	} else {
		rr = sdk.OneRat
		rs = sdk.OneRat
	}

	ad.distributeToValidators(vals, totalValShares, ad.getBlockAwardAndTxFees(), totalVotingPower, rr, rs)

	// distribute to the backup validators
	if len(backups) > 0 && totalBackupShares > 0 {
		rr = sdk.OneRat.Sub(utils.GetParams().ValidatorsBlockAwardRatio)
		rs = sdk.NewRat(totalBackupShares, totalValShares+totalBackupShares)
		ad.distributeToValidators(backups, totalBackupShares, ad.getBlockAwardAndTxFees(), totalVotingPower, rr, rs)
	}

	commons.Transfer(utils.MintAccount, utils.HoldAccount, ad.getBlockAward().Mul(sdk.NewInt(utils.BlocksPerHour)))

	// reset block gas fee
	utils.BlockGasFee.SetInt64(0)
}

func (ad *awardDistributor) calcVotingPower(rawValidators Validators) (normalizedValidators []*validator, totalShares int64, totalVotingPower int64) {
	totalShares = 0
	totalVotingPower = 0

	for _, val := range rawValidators {
		var validator validator
		var delegators []*delegator
		candidate := GetCandidateByAddress(common.HexToAddress(val.OwnerAddress))
		if candidate.ParseShares() == sdk.ZeroInt {
			continue
		}

		validator.ownerAddress = common.HexToAddress(candidate.OwnerAddress)
		validator.pk = candidate.PubKey

		// Get all delegators
		delegations := GetDelegationsByPubKey(candidate.PubKey)
		for _, delegation := range delegations {
			// if the amount of staked CMTs is less than 1000, no awards will be distributed.
			minStakingAmount := sdk.NewInt(utils.GetParams().MinStakingAmount).Mul(sdk.E18Int)
			if delegation.Shares().LT(minStakingAmount) {
				continue
			}

			d := delegator{}
			d.address = delegation.DelegatorAddress
			d.s = delegation.Shares().Div(sdk.E18Int).Int64()
			delegators = append(delegators, &d)

			tenDaysAgo, _ := utils.GetTimeBefore(10 * 24)
			ninetyDaysAgo, _ := utils.GetTimeBefore(90 * 24)
			m1 := GetCandidateDailyStakeMax(delegation.PubKey, tenDaysAgo)
			m2 := GetCandidateDailyStakeMax(delegation.PubKey, ninetyDaysAgo)
			s1, _ := sdk.NewIntFromString(m1)
			s2, _ := sdk.NewIntFromString(m2)
			d.s1 = s1.Div(sdk.E18Int).Int64()
			d.s2 = s2.Div(sdk.E18Int).Int64()
			d.c = sdk.NewRat(int64(delegation.ParseCompRate()*1000), 1000)

			t, _ := utils.Diff(delegation.CreatedAt)
			if t > HalfYear {
				t = HalfYear
			}
			d.t = t
			d.n += 1
		}

		// calculator voting power for delegators
		for _, d := range delegators {
			d.vp = calcVotingPowerForDelegator(d.s1, d.s2, d.t, d.n, d.s)
			ad.logger.Debug("Calculating voting power for delegator", "address", d.address, "s1", d.s1, "s2", d.s2, "t", d.t, "n", d.n, "s", d.s, "vp", d.vp)
			validator.vp += d.vp
			totalVotingPower += d.vp
			validator.s += d.s
			totalShares += d.s
		}

		// update pending voting power
		candidate.PendingVotingPower = validator.vp
		updateCandidate(candidate)

		validator.delegators = delegators
		normalizedValidators = append(normalizedValidators, &validator)
	}

	return
}

func (ad *awardDistributor) distributeToValidators(vals []*validator, totalShares int64, totalAward sdk.Int, totalVotingPower int64, rr, rs sdk.Rat) {
	actualDistributed := sdk.NewInt(0)
	for _, val := range vals {
		p := val.computeTotalSharesPercentage(totalShares, false)
		award := totalAward.MulRat(p)
		ad.logger.Debug("Prepare to distribute.", "address", val.ownerAddress, "totalAward", totalAward, "p", p, "award", award)
		ad.doDistribute(val, award, totalVotingPower, rr, rs)
		actualDistributed = actualDistributed.Add(award)
	}

	// If there is remaining, distribute a second round.
	remaining := totalAward.Sub(actualDistributed)
	if remaining.GT(sdk.ZeroInt) {
		ad.logger.Debug("there is remaining award, distribute a second round.", "remaining", remaining)
		for _, val := range vals {
			ad.doDistribute(val, remaining, totalVotingPower, rr, rs)
		}
	}
}

func (ad *awardDistributor) doDistribute(val *validator, totalAward sdk.Int, totalVotingPower int64, rr, rs sdk.Rat) {
	ad.logger.Debug("Distribute", "ownerAddress", val.ownerAddress, "totalAward", totalAward, "totalVotingPower", totalVotingPower, "rr", rr, "rs", rs)
	t := sdk.ZeroRat

	// distribute to the delegators
	for _, d := range val.delegators {
		a := sdk.OneRat.Sub(d.c)
		b := sdk.NewRat(d.vp*a.Num().Int64(), a.Denom().Int64())
		c := totalAward.MulRat(b.Mul(rr).Quo(rs).Quo(sdk.NewRat(totalVotingPower, 1)))
		ad.awardToDelegator(d, val, c)
		t = t.Add(sdk.NewRat(d.vp*d.c.Num().Int64(), d.c.Denom().Int64()))
		ad.logger.Debug("Distribute to delegator", "address", d.address, "award", c)
	}

	// distribute to the validator
	c := totalAward.MulRat(t.Mul(rr).Quo(rs).Quo(sdk.NewRat(totalVotingPower, 1)))
	ad.awardToValidator(val, c)
	ad.logger.Debug("Distribute to validator", "address", val.ownerAddress, "award", c)
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
	now := utils.GetNow()

	// add doDistribute to stake of the delegator
	delegation := GetDelegation(d.address, v.pk)
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

func calcVotingPowerForDelegator(s1, s2, t, n, s int64) int64 {
	one := sdk.OneRat
	r1 := sdk.NewRat(s1, s2)
	r2 := sdk.NewRat(t, 180)
	r3 := sdk.NewRat(n, 10)
	r4 := sdk.NewRat(s, 1)

	r1 = r1.Mul(r1)
	r2 = r2.Add(one)
	r3 = one.Sub(one.Quo(r3.Add(one)))
	r3 = r3.Mul(r3)
	x, _ := r1.Mul(r3).Mul(r4).Float64()
	f2, _ := r2.Float64()
	f2 = utils.RoundFloat(f2, 2)
	l := math.Log2(f2)
	return int64(x * l)
}
