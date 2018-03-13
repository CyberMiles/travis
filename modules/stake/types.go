package stake

import (
	"bytes"
	"sort"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
)

// Params defines the high level settings for staking
type Params struct {
	HoldAccount sdk.Actor `json:"hold_account"` // PubKey where all bonded coins are held

	MaxVals          uint16 `json:"max_vals"`           // maximum number of validators
	AllowedBondDenom string `json:"allowed_bond_denom"` // bondable coin denomination

	// gas costs for txs
	GasDeclareCandidacy int64 `json:"gas_declare_candidacy"`
	GasEditCandidacy    int64 `json:"gas_edit_candidacy"`
	GasDelegate         int64 `json:"gas_delegate"`
	GasUnbond           int64 `json:"gas_unbond"`
	GasProposeSlot		int64 `json:"gas_propose_slot"`
	Validators			string `json:"validators"`
}

var DefaultHoldAccount = sdk.NewActor(stakingModuleName, []byte("00000000000000000000000000000000"))

func defaultParams() Params {
	return Params{
		HoldAccount:         DefaultHoldAccount,
		MaxVals:             100,
		AllowedBondDenom:    "cmt",
		GasDeclareCandidacy: 0,
		GasEditCandidacy:    0,
		GasDelegate:         0,
		GasUnbond:           0,
		GasProposeSlot:		 0,
	}
}

//_________________________________________________________________________

// Candidate defines the total amount of bond shares and their exchange rate to
// coins. Accumulation of interest is modelled as an in increase in the
// exchange rate, and slashing as a decrease.  When coins are delegated to this
// candidate, the candidate is credited with a DelegatorBond whose number of
// bond shares is based on the amount of coins delegated divided by the current
// exchange rate. Voting power can be calculated as total bonds multiplied by
// exchange rate.
// NOTE if the Owner.Empty() == true then this is a candidate who has revoked candidacy
type Candidate struct {
	PubKey      crypto.PubKey `json:"pub_key"`      // Pubkey of candidate
	Owner       sdk.Actor     `json:"owner"`        // Sender of BondTx - UnbondTx returns here
	Shares      uint64        `json:"shares"`       // Total number of delegated shares to this candidate, equivalent to coins held in bond account
	VotingPower uint64        `json:"voting_power"` // Voting power if pubKey is a considered a validator
	Description Description   `json:"description"`  // Description terms for the candidate
}

// Description - description fields for a candidate
type Description struct {
	Moniker  string `json:"moniker"`
	Identity string `json:"identity"`
	Website  string `json:"website"`
	Details  string `json:"details"`
}

// NewCandidate - initialize a new candidate
func NewCandidate(pubKey crypto.PubKey, owner sdk.Actor) *Candidate {
	return &Candidate{
		PubKey:      pubKey,
		Owner:       owner,
		Shares:      0,
		VotingPower: 0,
	}
}

// Validator returns a copy of the Candidate as a Validator.
// Should only be called when the Candidate qualifies as a validator.
func (c *Candidate) validator() Validator {
	return Validator(*c)
}

// Validator is one of the top Candidates
type Validator Candidate

// ABCIValidator - Get the validator from a bond value
func (v Validator) ABCIValidator() *abci.Validator {
	return &abci.Validator{
		PubKey: wire.BinaryBytes(v.PubKey),
		Power:  int64(v.VotingPower),
	}
}

//_________________________________________________________________________

// TODO replace with sorted multistore functionality

// Candidates - list of Candidates
type Candidates []*Candidate

var _ sort.Interface = Candidates{} //enforce the sort interface at compile time

// nolint - sort interface functions
func (cs Candidates) Len() int      { return len(cs) }
func (cs Candidates) Swap(i, j int) { cs[i], cs[j] = cs[j], cs[i] }
func (cs Candidates) Less(i, j int) bool {
	vp1, vp2 := cs[i].VotingPower, cs[j].VotingPower
	pk1, pk2 := cs[i].PubKey.Bytes(), cs[j].PubKey.Bytes()

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

//func updateVotingPower(store state.SimpleDB) {
//candidates := loadCandidates(store)
//candidates.updateVotingPower(store)
//}

// update the voting power and save
func (cs Candidates) updateVotingPower(store state.SimpleDB) Candidates {

	// update voting power
	for _, c := range cs {
		if c.VotingPower != c.Shares {
			c.VotingPower = c.Shares
		}
	}
	cs.Sort()
	for i, c := range cs {
		// truncate the power
		if i >= int(loadParams(store).MaxVals) {
			c.VotingPower = 0
		}
		saveCandidate(store, c)
	}
	return cs
}

// Validators - get the most recent updated validator set from the
// Candidates. These bonds are already sorted by VotingPower from
// the UpdateVotingPower function which is the only function which
// is to modify the VotingPower
func (cs Candidates) Validators() Validators {

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
		validators[i] = c.validator()
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
	pk1, pk2 := vs[i].PubKey.Bytes(), vs[j].PubKey.Bytes()
	return bytes.Compare(pk1, pk2) == -1
}

// Sort - Sort validators by pubkey
func (vs Validators) Sort() {
	sort.Sort(vs)
}

// determine all changed validators between two validator sets
func (vs Validators) validatorsChanged(vs2 Validators) (changed []*abci.Validator) {

	//first sort the validator sets
	vs.Sort()
	vs2.Sort()

	max := len(vs) + len(vs2)
	changed = make([]*abci.Validator, max)
	i, j, n := 0, 0, 0 //counters for vs loop, vs2 loop, changed element

	for i < len(vs) && j < len(vs2) {

		if !vs[i].PubKey.Equals(vs2[j].PubKey) {
			// pk1 > pk2, a new validator was introduced between these pubkeys
			if bytes.Compare(vs[i].PubKey.Bytes(), vs2[j].PubKey.Bytes()) == 1 {
				changed[n] = vs2[j].ABCIValidator()
				n++
				j++
				continue
			} // else, the old validator has been removed
			changed[n] = &abci.Validator{vs[i].PubKey.Bytes(), 0}
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
		changed[n] = &abci.Validator{vs[i].PubKey.Bytes(), 0}
	}

	return changed[:n]
}

// UpdateValidatorSet - Updates the voting power for the candidate set and
// returns the subset of validators which have changed for Tendermint
func UpdateValidatorSet(store state.SimpleDB) (change []*abci.Validator, err error) {

	// get the validators before update
	candidates := loadCandidates(store)

	v1 := candidates.Validators()
	v2 := candidates.updateVotingPower(store).Validators()

	change = v1.validatorsChanged(v2)
	return
}

//_________________________________________________________________________

// DelegatorBond represents the bond with tokens held by an account.  It is
// owned by one delegator, and is associated with the voting power of one
// pubKey.
type DelegatorBond struct {
	PubKey crypto.PubKey
	Shares uint64
}

type Slot struct {
	Id string
	ValidatorPubKey crypto.PubKey
	TotalAmount uint64
	AvailableAmount uint64
	ProposedRoi uint64
}

func NewSlot(id string, validatorPubKey crypto.PubKey, totalAmount uint64, availableAmount uint64, proposedRoi uint64) *Slot {
	return &Slot{
		Id: id,
		ValidatorPubKey: validatorPubKey,
		TotalAmount: totalAmount,
		AvailableAmount: availableAmount,
		ProposedRoi: proposedRoi,
	}
}

