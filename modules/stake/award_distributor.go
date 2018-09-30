package stake

import (
	"fmt"
	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/sdk/state"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/libs/log"
	"math"
)

type simpleValidator struct {
	s            int64
	ownerAddress common.Address
	id           int64
	delegators   []*simpleDelegator
	vp           int64
	state        string
}

func (v *simpleValidator) distributeAward(totalAward sdk.Int, totalVotingPower int64) (res sdk.Int) {
	t := sdk.ZeroRat
	res = sdk.ZeroInt

	// distribute to the delegators
	for _, d := range v.delegators {
		a := sdk.OneRat.Sub(d.c)
		b := sdk.NewRat(d.vp*a.Num().Int64(), a.Denom().Int64())
		c := totalAward.MulRat(b.Quo(sdk.NewRat(totalVotingPower, 1)))
		d.distributeAward(v, c)
		res = res.Add(c)
		t = t.Add(sdk.NewRat(d.vp*d.c.Num().Int64(), d.c.Denom().Int64()))
		//fmt.Println(d)
		//fmt.Printf("a: %v, b: %v, c: %v, t: %v, res: %v\n", a, b, c, t, res)
	}

	// distribute to the validator self
	c := totalAward.MulRat(t.Quo(sdk.NewRat(totalVotingPower, 1)))
	v.distributeAwardToSelf(c)
	res = res.Add(c)
	//fmt.Printf("c: %v, res: %v\n", c, res)
	return
}

func (v *simpleValidator) distributeAwardToSelf(award sdk.Int) {
	// A validator is a delegator as well
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
	delegation := GetDelegation(d.address, v.id)
	if delegation == nil {
		return
	}

	delegation.AddAwardAmount(award)
	UpdateDelegation(delegation)

	// accumulate shares of the validator
	val := GetCandidateByAddress(v.ownerAddress)
	val.AddShares(award)
	updateCandidate(val)
}

func (d simpleDelegator) String() string {
	return fmt.Sprintf("[simpleDeligator] address: %s, s: %d, c: %v, vp: %d", d.address.String(), d.s, d.c, d.vp)
}

type AwardInfo struct {
	Address common.Address `json:"address"`
	State   string         `json:"state"`
	Amount  string         `json:"amount"`
}

type AwardInfos []AwardInfo

//_______________________________________________________________________

type awardDistributor struct {
	store            state.SimpleDB
	height           int64
	validators       Validators
	backupValidators Validators
	transactionFees  sdk.Int
	logger           log.Logger
}

func NewAwardDistributor(store state.SimpleDB, height int64, validators, backupValidators Validators, logger log.Logger) *awardDistributor {
	return &awardDistributor{store, height, validators, backupValidators, sdk.NewIntFromBigInt(utils.BlockGasFee), logger}
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
	var awardInfos AwardInfos
	vals, _, totalValVotingPower := ad.buildValidators(ad.validators)
	backups, totalBackupShares, totalBackupVotingPower := ad.buildValidators(ad.backupValidators)
	var rr sdk.Rat
	if len(backups) > 0 && totalBackupShares > 0 {
		rr = utils.GetParams().ValidatorsBlockAwardRatio
	} else {
		rr = sdk.OneRat
	}

	totalAward := ad.getBlockAwardAndTxFees().MulRat(rr)
	awardInfos = ad.distribute(vals, totalAward, totalValVotingPower, awardInfos)

	// distribute to the backup validators
	if len(backups) > 0 && totalBackupShares > 0 {
		rr = sdk.OneRat.Sub(rr)
		totalAward := ad.getBlockAwardAndTxFees().MulRat(rr)
		awardInfos = ad.distribute(backups, totalAward, totalBackupVotingPower, awardInfos)
	}

	commons.Transfer(utils.MintAccount, utils.HoldAccount, ad.getBlockAward())

	// reset block gas fee
	utils.BlockGasFee.SetInt64(0)
	saveAwardInfo(ad.store, awardInfos)
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
		validator.id = candidate.Id
		validator.state = candidate.State

		// Get all delegators
		delegations := GetDelegationsByCandidate(candidate.Id, "Y")
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

func (ad *awardDistributor) distribute(vals []*simpleValidator, totalAward sdk.Int, totalVotingPower int64, awardInfos AwardInfos) AwardInfos {
	for _, val := range vals {
		ad.logger.Debug("Prepare to distribute.", "address", val.ownerAddress, "totalAward", totalAward)
		//fmt.Printf("Prepare to distribute, address: %v, totalAward: %v\n", val.ownerAddress.String(), totalAward)
		award := val.distributeAward(totalAward, totalVotingPower)
		ai := AwardInfo{Address: val.ownerAddress, State: val.state, Amount: award.String()}
		awardInfos = append(awardInfos, ai)
	}
	return awardInfos
}

func (ad awardDistributor) getBlockAwardAndTxFees() sdk.Int {
	return ad.getBlockAward().Add(ad.transactionFees)
}

func saveAwardInfo(store state.SimpleDB, awardInfos AwardInfos) {
	b, err := cdc.MarshalBinary(&awardInfos)
	if err != nil {
		panic(err)
	}

	store.Set(utils.AwardInfosKey, b)
}
