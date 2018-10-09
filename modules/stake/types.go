package stake

import (
	"bytes"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"github.com/tendermint/tendermint/crypto"
	"golang.org/x/crypto/ripemd160"
	"math"
)

//_________________________________________________________________________

// Candidate defines the total Amount of bond shares and their exchange rate to
// coins. Accumulation of interest is modelled as an in increase in the
// exchange rate, and slashing as a decrease.  When coins are delegated to this
// candidate, the candidate is credited with a DelegatorBond whose number of
// bond shares is based on the Amount of coins delegated divided by the current
// exchange rate. Voting power can be calculated as total bonds multiplied by
// exchange rate.
// NOTE if the Owner.Empty() == true then this is a candidate who has revoked candidacy
type Candidate struct {
	Id                 int64        `json:"id"`
	PubKey             types.PubKey `json:"pub_key"`       // Pubkey of candidate
	OwnerAddress       string       `json:"owner_address"` // Sender of BondTx - UnbondTx returns here
	Shares             string       `json:"shares"`        // Total number of delegated shares to this candidate, equivalent to coins held in bond account
	VotingPower        int64        `json:"voting_power"`  // Voting power if pubKey is a considered a validator
	PendingVotingPower int64        `json:"pending_voting_power"`
	MaxShares          string       `json:"max_shares"`
	CompRate           sdk.Rat      `json:"comp_rate"`
	CreatedAt          string       `json:"created_at"`
	Description        Description  `json:"description"`
	Verified           string       `json:"verified"`
	Active             string       `json:"active"`
	BlockHeight        int64        `json:"block_height"`
	Rank               int64        `json:"rank"`
	State              string       `json:"state"`
	NumOfDelegators    int64        `json:"num_of_delegators"`
}

type Description struct {
	Name     string `json:"name"`
	Website  string `json:"website"`
	Location string `json:"location"`
	Email    string `json:"email"`
	Profile  string `json:"profile"`
}

// Validator returns a copy of the Candidate as a Validator.
// Should only be called when the Candidate qualifies as a validator.
func (c *Candidate) Validator() Validator {
	return Validator(*c)
}

func (c *Candidate) ParseShares() sdk.Int {
	return utils.ParseInt(c.Shares)
}

func (c *Candidate) ParseMaxShares() sdk.Int {
	return utils.ParseInt(c.MaxShares)
}

func (c *Candidate) AddShares(value sdk.Int) (res sdk.Int) {
	res = c.ParseShares().Add(value)
	c.Shares = res.String()
	return
}

func (c *Candidate) SelfStakingAmount(ssr sdk.Rat) (res sdk.Int) {
	res = c.ParseMaxShares().MulRat(ssr)
	return
}

func (c *Candidate) Hash() []byte {
	var excludedFields []string
	bs := types.Hash(c, excludedFields)
	hasher := ripemd160.New()
	hasher.Write(bs)
	return hasher.Sum(nil)
}

func (c *Candidate) CalcVotingPower(blockHeight int64) (res int64) {
	res = 0
	minStakingAmount := sdk.NewInt(utils.GetParams().MinStakingAmount).Mul(sdk.E18Int)
	delegations := GetDelegationsByCandidate(c.Id, "Y")
	sharesPercentage := c.computeTotalSharesPercentage()

	for _, d := range delegations {
		var vp int64
		// if the amount of staked CMTs is less than 1000, no awards will be distributed.
		if d.Shares().LT(minStakingAmount) {
			vp = 0
			d.ResetVotingPower()
		} else {
			vp = d.CalcVotingPower(sharesPercentage, blockHeight)
		}
		UpdateDelegation(d) // update delegator's voting power
		res += vp
	}
	return
}

func (c *Candidate) computeTotalSharesPercentage() (res sdk.Rat) {
	totalShares := GetCandidatesTotalShares()
	shares, _ := sdk.NewIntFromString(c.Shares)
	p := sdk.NewRat(shares.Div(sdk.E18Int).Int64(), totalShares.Div(sdk.E18Int).Int64())
	threshold := utils.GetParams().ValidatorSizeThreshold
	if p.GT(threshold) {
		res = threshold.Quo(p)
	} else {
		res = sdk.OneRat
	}

	return
}

func (c Candidate) IsActive() bool {
	return c.Active == "Y"
}

// Validator is one of the top Candidates
type Validator Candidate

// ABCIValidator - Get the validator from a bond value
func (v Validator) ABCIValidator() abci.Validator {
	pk := v.PubKey.PubKey.(crypto.PubKeyEd25519)
	return abci.Validator{
		PubKey: abci.PubKey{
			Type: abci.PubKeyEd25519,
			Data: pk[:],
		},
		Power: v.VotingPower,
	}
}

//_________________________________________________________________________

type Candidates []*Candidate

var _ sort.Interface = Candidates{} //enforce the sort interface at compile time

// nolint - sort interface functions
func (cs Candidates) Len() int      { return len(cs) }
func (cs Candidates) Swap(i, j int) { cs[i], cs[j] = cs[j], cs[i] }
func (cs Candidates) Less(i, j int) bool {
	vp1, vp2 := cs[i].VotingPower, cs[j].VotingPower
	pk1, pk2 := cs[i].PubKey.Address(), cs[j].PubKey.Address()

	//note that all ChainId and App must be the same for a group of candidates
	if vp1 != vp2 {
		return vp1 > vp2
	}
	return bytes.Compare(pk1, pk2) == -1
}

// Sort - Sort the array of bonded values
func (cs Candidates) Sort() {
	sort.Sort(cs)
}

// update the voting power and save
func (cs Candidates) updateVotingPower(blockHeight int64) Candidates {
	// update voting power
	for _, c := range cs {
		c.PendingVotingPower = c.CalcVotingPower(blockHeight)

		if c.Active == "N" {
			c.VotingPower = 0
		} else if c.VotingPower != c.PendingVotingPower {
			c.VotingPower = c.PendingVotingPower
		}
	}

	cs.Sort()
	for i, c := range cs {
		// truncate the power
		if i >= int(utils.GetParams().MaxVals) {
			if i >= (int(utils.GetParams().MaxVals + utils.GetParams().BackupVals)) {
				c.State = "Candidate"
				c.VotingPower = 0
			} else {
				c.State = "Backup Validator"
			}
		} else {
			c.State = "Validator"
		}

		c.Rank = int64(i)
		updateCandidate(c)
	}
	return cs
}

// Validators - get the most recent updated validator set from the
// Candidates. These bonds are already sorted by VotingPower from
// the UpdateVotingPower function which is the only function which
// is to modify the VotingPower
func (cs Candidates) Validators() Validators {
	cs.Sort()

	//test if empty
	if len(cs) == 1 {
		if cs[0].VotingPower == 0 {
			return nil
		}
	}

	validators := make(Validators, len(cs))
	for i, c := range cs {
		if c.VotingPower == 0 { //exit as soon as the first Voting power set to zero is found
			return validators[:i]
		}
		if i >= int(utils.GetParams().MaxVals) {
			return validators[:i]
		}
		validators[i] = c.Validator()
	}

	return validators
}

//_________________________________________________________________________

// Validators - list of Validators
type Validators []Validator

var _ sort.Interface = Validators{} //enforce the sort interface at compile time

// nolint - sort interface functions
func (vs Validators) Len() int      { return len(vs) }
func (vs Validators) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }
func (vs Validators) Less(i, j int) bool {
	//pk1, pk2 := vs[i].PubKey.Bytes(), vs[j].PubKey.Bytes()
	//return bytes.Compare(pk1, pk2) == -1

	pk1, pk2 := vs[i].PubKey, vs[j].PubKey
	return bytes.Compare(pk1.Address(), pk2.Address()) == -1
}

// Sort - Sort validators by pubkey
func (vs Validators) Sort() {
	sort.Sort(vs)
}

// determine all changed validators between two validator sets
func (vs Validators) validatorsChanged(vs2 Validators) (changed []abci.Validator) {

	//first sort the validator sets
	vs.Sort()
	vs2.Sort()

	max := len(vs) + len(vs2)
	changed = make([]abci.Validator, max)
	i, j, n := 0, 0, 0 //counters for vs loop, vs2 loop, changed element

	for i < len(vs) && j < len(vs2) {

		if !vs[i].PubKey.Equals(vs2[j].PubKey) {
			// pk1 > pk2, a new validator was introduced between these pubkeys
			//if bytes.Compare(vs[i].PubKey.Bytes(), vs2[j].PubKey.Bytes()) == 1 {
			if bytes.Compare(vs[i].PubKey.Address(), vs2[j].PubKey.Address()) == 1 {
				changed[n] = vs2[j].ABCIValidator()
				n++
				j++
				continue
			} // else, the old validator has been removed
			pk := vs[i].PubKey.PubKey.(crypto.PubKeyEd25519)
			changed[n] = abci.Ed25519Validator(pk[:], 0)
			n++
			i++
			continue
		}

		if vs[i].VotingPower != vs2[j].VotingPower {
			changed[n] = vs2[j].ABCIValidator()
			n++
		}
		j++
		i++
	}

	// add any excess validators in set 2
	for ; j < len(vs2); j, n = j+1, n+1 {
		changed[n] = vs2[j].ABCIValidator()
	}

	// remove any excess validators left in set 1
	for ; i < len(vs); i, n = i+1, n+1 {
		pk := vs[i].PubKey.PubKey.(crypto.PubKeyEd25519)
		changed[n] = abci.Ed25519Validator(pk[:], 0)
	}

	return changed[:n]
}

func (vs Validators) Remove(i int) Validators {
	copy(vs[i:], vs[i+1:])
	return vs[:len(vs)-1]
}

// UpdateValidatorSet - Updates the voting power for the candidate set and
// returns the subset of validators which have changed for Tendermint
func UpdateValidatorSet(blockHeight int64) (change []abci.Validator, err error) {
	// get the validators before update
	candidates := GetCandidates()
	v1 := candidates.Validators()
	v2 := candidates.updateVotingPower(blockHeight).Validators()
	change = v1.validatorsChanged(v2)

	// clean all of the candidates had been withdrawed
	cleanCandidates()
	return
}

// Deactivate the validators
func (vs Validators) Deactivate() {
	// update voting power
	for _, v := range vs {
		v.Active = "N"
		v.VotingPower = 0
		c := Candidate(v)
		updateCandidate(&c)
	}
}

//_________________________________________________________________________

type Delegation struct {
	Id                    int64          `json:"id"`
	DelegatorAddress      common.Address `json:"delegator_address"`
	PubKey                types.PubKey   `json:"pub_key"`
	ValidatorAddress      string         `json:"validator_address"`
	DelegateAmount        string         `json:"delegate_amount"`
	AwardAmount           string         `json:"award_amount"`
	WithdrawAmount        string         `json:"withdraw_amount"`
	PendingWithdrawAmount string         `json:"pending_withdraw_amount"`
	SlashAmount           string         `json:"slash_amount"`
	CompRate              sdk.Rat        `json:"comp_rate"`
	VotingPower           int64          `json:"voting_power"`
	CreatedAt             string         `json:"created_at"`
	State                 string         `json:"state"`
	BlockHeight           int64          `json:"block_height"`
	AverageStakingDate    int64          `json:"average_staking_date"`
	CandidateId           int64          `json:"candidate_id"`
}

func (d *Delegation) Shares() (res sdk.Int) {
	res = d.ParseDelegateAmount().Add(d.ParseAwardAmount()).Sub(d.ParseWithdrawAmount()).Sub(d.ParseSlashAmount()).Sub(d.ParsePendingWithdrawAmount())
	return
}

func (d *Delegation) ParseDelegateAmount() sdk.Int {
	return utils.ParseInt(d.DelegateAmount)
}

func (d *Delegation) ParseAwardAmount() sdk.Int {
	return utils.ParseInt(d.AwardAmount)
}

func (d *Delegation) ParseWithdrawAmount() sdk.Int {
	return utils.ParseInt(d.WithdrawAmount)
}

func (d *Delegation) ParsePendingWithdrawAmount() sdk.Int {
	return utils.ParseInt(d.PendingWithdrawAmount)
}

func (d *Delegation) ParseSlashAmount() sdk.Int {
	return utils.ParseInt(d.SlashAmount)
}

func (d *Delegation) AddDelegateAmount(value sdk.Int) (res sdk.Int) {
	res = d.ParseDelegateAmount().Add(value)
	d.DelegateAmount = res.String()
	return
}

func (d *Delegation) AddAwardAmount(value sdk.Int) (res sdk.Int) {
	res = d.ParseAwardAmount().Add(value)
	d.AwardAmount = res.String()
	return
}

func (d *Delegation) AddWithdrawAmount(value sdk.Int) (res sdk.Int) {
	res = d.ParseWithdrawAmount().Add(value)
	d.WithdrawAmount = res.String()
	return
}

func (d *Delegation) AddPendingWithdrawAmount(value sdk.Int) (res sdk.Int) {
	res = d.ParsePendingWithdrawAmount().Add(value)
	d.PendingWithdrawAmount = res.String()
	return
}

func (d *Delegation) AddSlashAmount(value sdk.Int) (res sdk.Int) {
	res = d.ParseSlashAmount().Add(value)
	d.SlashAmount = res.String()
	return
}

func (d *Delegation) ResetVotingPower() {
	d.VotingPower = 0
}

func (d *Delegation) CalcVotingPower(sharesPercentage sdk.Rat, blockHeight int64) int64 {
	candidate := GetCandidateById(d.CandidateId)
	tenDaysAgoHeight := blockHeight - utils.ConvertDaysToHeight(10)
	ninetyDaysAgoHeight := blockHeight - utils.ConvertDaysToHeight(90)
	s1 := GetCandidateDailyStakeMaxValue(candidate.Id, tenDaysAgoHeight)
	s2 := GetCandidateDailyStakeMaxValue(candidate.Id, ninetyDaysAgoHeight)
	snum := s1.Div(sdk.E18Int).Int64()
	sdenom := s2.Div(sdk.E18Int).Int64()
	if sdenom == 0 {
		sdenom = 1
	}
	s := d.Shares().Div(sdk.E18Int).MulRat(sharesPercentage).Int64()

	t := d.AverageStakingDate
	if t == 0 {
		t = 1
	} else if t > utils.HalfYear {
		t = utils.HalfYear
	}

	one := sdk.OneRat
	r1 := sdk.NewRat(snum, sdenom)
	r2 := sdk.NewRat(t, 180)
	r3 := sdk.NewRat(candidate.NumOfDelegators*4, 1)
	r4 := sdk.NewRat(s, 1)

	r1 = r1.Mul(r1)
	r2 = r2.Add(one)
	r3 = one.Sub(one.Quo(r3.Add(one)))
	r3 = r3.Mul(r3)
	x, _ := r1.Mul(r3).Mul(r4).Float64()
	f2, _ := r2.Float64()
	f2 = utils.RoundFloat(f2, 2)
	l := math.Log2(f2)
	vp := int64(math.Ceil(x * l))
	d.VotingPower = vp
	return vp
}

func (d *Delegation) Hash() []byte {
	var excludedFields []string
	bs := types.Hash(d, excludedFields)
	hasher := ripemd160.New()
	hasher.Write(bs)
	return hasher.Sum(nil)
}

func (d *Delegation) AccumulateAverageStakingDate() {
	minStakingAmount := sdk.NewInt(utils.GetParams().MinStakingAmount).Mul(sdk.E18Int)
	if d.Shares().GTE(minStakingAmount) {
		d.AverageStakingDate += 1
	}
}

func (d *Delegation) ReduceAverageStakingDate(withdrawAmount sdk.Int) {
	num := withdrawAmount.Div(sdk.E18Int).Int64()
	denom := d.Shares().Div(sdk.E18Int).Int64()
	if denom == 0 {
		d.AverageStakingDate = 0
	} else {
		p := sdk.NewRat(num, denom)
		d.AverageStakingDate = sdk.NewInt(d.AverageStakingDate).MulRat(sdk.OneRat.Sub(p)).Int64()
	}
}

type DelegateHistory struct {
	Id               int64          `json:"id"`
	DelegatorAddress common.Address `json:"delegator_address"`
	Amount           sdk.Int        `json:"amount"`
	OpCode           string         `json:"op_code"`
	BlockHeight      int64          `json:"block_height"`
	CandidateId      int64          `json:"candidate_id"`
}

func (d *DelegateHistory) Hash() []byte {
	var excludedFields []string
	bs := types.Hash(d, excludedFields)
	hasher := ripemd160.New()
	hasher.Write(bs)
	return hasher.Sum(nil)
}

type Slash struct {
	Id          int64   `json:"id"`
	SlashRatio  sdk.Rat `json:"slash_ratio"`
	SlashAmount sdk.Int `json:"slash_amount"`
	Reason      string  `json:"reason"`
	CreatedAt   string  `json:"created_at"`
	BlockHeight int64   `json:"block_height"`
	CandidateId int64   `json:"candidate_id"`
}

func (s *Slash) Hash() []byte {
	var excludedFields []string
	bs := types.Hash(s, excludedFields)
	hasher := ripemd160.New()
	hasher.Write(bs)
	return hasher.Sum(nil)
}

type UnstakeRequest struct {
	Id                   int64          `json:"id"`
	DelegatorAddress     common.Address `json:"delegator_address"`
	InitiatedBlockHeight int64          `json:"initiated_block_height"`
	PerformedBlockHeight int64          `json:"performed_block_height"`
	Amount               string         `json:"amount"`
	State                string         `json:"state"`
	CreatedAt            string         `json:"created_at"`
	CandidateId          int64          `json:"candidate_id"`
}

func (r *UnstakeRequest) Hash() []byte {
	var excludedFields []string
	bs := types.Hash(r, excludedFields)
	hasher := ripemd160.New()
	hasher.Write(bs)
	return hasher.Sum(nil)
}

type CandidateDailyStake struct {
	Id          int64  `json:"id"`
	Amount      string `json:"amount"`
	BlockHeight int64  `json:"block_height"`
	CandidateId int64  `json:"candidate_id"`
}

func (c *CandidateDailyStake) Hash() []byte {
	var excludedFields []string
	bs := types.Hash(c, excludedFields)
	hasher := ripemd160.New()
	hasher.Write(bs)
	return hasher.Sum(nil)
}

type CubePubKey struct {
	CubeBatch string `json:"cube_batch"`
	PubKey    string `json:"pub_key"`
}

type CandidateAccountUpdateRequest struct {
	Id                  int64          `json:"id"`
	CandidateId         int64          `json:"candidate_id"`
	FromAddress         common.Address `json:"from_address"`
	ToAddress           common.Address `json:"to_address"`
	CreatedBlockHeight  int64          `json:"created_block_height"`
	AcceptedBlockHeight int64          `json:"accepted_block_height"`
	State               string         `json:"state"`
}

func (c *CandidateAccountUpdateRequest) Hash() []byte {
	var excludedFields []string
	bs := types.Hash(c, excludedFields)
	hasher := ripemd160.New()
	hasher.Write(bs)
	return hasher.Sum(nil)
}
