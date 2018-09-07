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

type simpleValidator struct {
	s            int64
	ownerAddress common.Address
	pk           types.PubKey
	delegators   []*simpleDelegator
	vp           int64
}

func (v *simpleValidator) computeTotalSharesPercentage(totalShares int64, redistribute bool) sdk.Rat {
	p := sdk.NewRat(v.s, totalShares)
	threshold := utils.GetParams().ValidatorSizeThreshold
	if !redistribute && p.Cmp(threshold) > 0 {
		p = threshold
	}

	return p
}

func (v *simpleValidator) distributeToAll(totalAward sdk.Int, totalVotingPower int64, rr, rs sdk.Rat) {
	//ad.logger.Debug("Distribute", "ownerAddress", val.ownerAddress, "totalAward", totalAward, "totalVotingPower", totalVotingPower, "rr", rr, "rs", rs)
	t := sdk.ZeroRat

	// distribute to the delegators
	for _, d := range v.delegators {
		a := sdk.OneRat.Sub(d.c)
		b := sdk.NewRat(d.vp*a.Num().Int64(), a.Denom().Int64())
		c := totalAward.MulRat(b.Mul(rr).Quo(rs).Quo(sdk.NewRat(totalVotingPower, 1)))
		d.distributeAward(v, c)
		t = t.Add(sdk.NewRat(d.vp*d.c.Num().Int64(), d.c.Denom().Int64()))
		//ad.logger.Debug("Distribute to simpleDelegator", "address", d.address, "award", c)
	}

	// distribute to the validator self
	c := totalAward.MulRat(t.Mul(rr).Quo(rs).Quo(sdk.NewRat(totalVotingPower, 1)))
	v.distributeAwardToSelf(c)
	//ad.logger.Debug("Distribute to simpleValidator", "address", val.ownerAddress, "award", c)
	return
}

func (v *simpleValidator) distributeAwardToSelf(award sdk.Int) {
	// A simpleValidator is also a simpleDelegator
	d := simpleDelegator{address: v.ownerAddress}
	d.distributeAward(v, award)
}

func (v simpleValidator) String() string {
	return fmt.Sprintf("[simpleValidator] ownerAddress: %s, delegators: %d, vp: %d", v.ownerAddress.String(), len(v.delegators), v.vp)
}

//_______________________________________________________________________

type simpleDelegator struct {
	address common.Address
	s       int64
	c       sdk.Rat
	vp      int64
}

func (d simpleDelegator) distributeAward(v *simpleValidator, award sdk.Int) {
	now := utils.GetNow()

	// add doDistribute to stake of the simpleDelegator
	delegation := GetDelegation(d.address, v.pk)
	if delegation == nil {
		return
	}

	delegation.AddAwardAmount(award)
	delegation.UpdatedAt = now
	UpdateDelegation(delegation)

	// accumulate shares of the simpleValidator
	val := GetCandidateByAddress(v.ownerAddress)
	val.AddShares(award)
	val.UpdatedAt = now
	updateCandidate(val)
}

func (d simpleDelegator) String() string {
	return fmt.Sprintf("[simpleDeligator] address: %s, s: %d, c: %vp, vp: %d", d.address.String(), d.s, d.c, d.vp)
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
	base, ok := sdk.NewIntFromString(utils.BasicMintableAmount)
	if !ok {
		return
	}

	year := ad.height / utils.YearlyBlockNumber
	b, _ := utils.GetParams().InflationRate.Add(sdk.OneRat).Float64()
	pow := math.Pow(b, float64(year))
	pow = utils.RoundFloat(pow, 2)
	r := sdk.NewRat(int64(pow*100), 100)
	amount = base.MulRat(r)
	ad.logger.Debug("getMintableAmount", "height", ad.height, "year", year, "amount", amount)
	return
}

func (ad awardDistributor) getBlockAward() (blockAward sdk.Int) {
	ybn := sdk.NewInt(utils.YearlyBlockNumber)
	blockAward = ad.getMintableAmount().MulRat(utils.GetParams().InflationRate).Div(ybn)
	ad.logger.Debug("getBlockAward", "yearly_block_number", ybn, "total_block_award", blockAward)
	return
}

func (ad awardDistributor) Distribute() {
	vals, totalValShares, totalValVotingPower := ad.buildValidators(ad.validators)
	backups, totalBackupShares, totalBackupVotingPower := ad.buildValidators(ad.backupValidators)
	totalVotingPower := totalValVotingPower + totalBackupVotingPower
	totalShares := totalValShares + totalBackupShares
	var rr, rs sdk.Rat
	if len(backups) > 0 && totalBackupShares > 0 {
		rr = utils.GetParams().ValidatorsBlockAwardRatio
		rs = sdk.NewRat(totalValShares, totalShares)
	} else {
		rr = sdk.OneRat
		rs = sdk.OneRat
	}

	ad.distribute(vals, totalShares, ad.getBlockAwardAndTxFees(), totalVotingPower, rr, rs)

	// distribute to the backup validators
	if len(backups) > 0 && totalBackupShares > 0 {
		rr = sdk.OneRat.Sub(utils.GetParams().ValidatorsBlockAwardRatio)
		rs = sdk.NewRat(totalBackupShares, totalValShares+totalBackupShares)
		ad.distribute(backups, totalShares, ad.getBlockAwardAndTxFees(), totalVotingPower, rr, rs)
	}

	commons.Transfer(utils.MintAccount, utils.HoldAccount, ad.getBlockAward().Mul(sdk.NewInt(utils.GetRewardInterval())))

	// reset block gas fee
	utils.BlockGasFee.SetInt64(0)
}

func (ad *awardDistributor) buildValidators(rawValidators Validators) (normalizedValidators []*simpleValidator, totalShares int64, totalVotingPower int64) {
	totalShares = 0
	totalVotingPower = 0

	for _, val := range rawValidators {
		var validator simpleValidator
		var delegators []*simpleDelegator
		candidate := GetCandidateByAddress(common.HexToAddress(val.OwnerAddress))
		if candidate == nil || candidate.ParseShares() == sdk.ZeroInt {
			continue
		}

		validator.ownerAddress = common.HexToAddress(candidate.OwnerAddress)
		validator.pk = candidate.PubKey

		// Get all delegators
		delegations := GetDelegationsByPubKey(candidate.PubKey)
		for _, delegation := range delegations {
			// if the amount of staked CMTs is less than 1000, no awards will be distributed.
			if delegation.VotingPower == 0 {
				continue
			}

			d := simpleDelegator{}
			d.address = delegation.DelegatorAddress
			d.s = delegation.Shares().Div(sdk.E18Int).Int64()
			d.c = delegation.CompRate
			d.vp = delegation.VotingPower
			delegators = append(delegators, &d)
			validator.s += d.s
			totalShares += d.s
			totalVotingPower += d.vp
		}

		// update pending voting power
		candidate.PendingVotingPower = validator.vp
		updateCandidate(candidate)

		validator.delegators = delegators
		normalizedValidators = append(normalizedValidators, &validator)
	}

	return
}

func (ad *awardDistributor) distribute(vals []*simpleValidator, totalShares int64, totalAward sdk.Int, totalVotingPower int64, rr, rs sdk.Rat) {
	award, remaining := sdk.ZeroInt, totalAward

	for _, val := range vals {
		p := val.computeTotalSharesPercentage(totalShares, false)
		award = totalAward.MulRat(p)
		ad.logger.Debug("Prepare to distribute.", "address", val.ownerAddress, "totalAward", totalAward, "p", p, "award", award)
		val.distributeToAll(award, totalVotingPower, rr, rs)
		remaining = remaining.Sub(award)
	}

	// If there is remaining, distribute a second round.
	if remaining.GT(sdk.ZeroInt) {
		ad.logger.Debug("there is remaining award, distribute a second round.", "remaining", remaining)
		for _, val := range vals {
			val.distributeToAll(remaining, totalVotingPower, rr, rs)
		}
	}
}

func (ad awardDistributor) getBlockAwardAndTxFees() sdk.Int {
	return ad.getBlockAward().Mul(sdk.NewInt(utils.GetRewardInterval())).Add(ad.transactionFees)
}
