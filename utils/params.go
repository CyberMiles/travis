package utils

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/CyberMiles/travis/types"
)

type Params struct {
	HoldAccount      common.Address `json:"hold_account"` // PubKey where all bonded coins are held
	MaxVals              uint16     `json:"max_vals"`     // maximum number of validators
	Validators           string     `json:"validators"`   // initial validators definition
	SelfStakingRatio     string     `json:"self_staking_ratio"`
	InflationRate        int64      `json:"inflation_rate"`
	StakeLimit           string     `json:"stake_limit"`
	UnstakeWaitPeriod    uint64      `json:"unstake_wait_period"`
	ProposalExpirePeriod uint64     `json:"proposal_expire_period"`

	DeclareCandidacy	uint64		`json:"declare_candidacy"`
	UpdateCandidacy		uint64		`json:"update_candidacy"`
	GovernancePropose	uint64		`json:"governance_proposal"`
	GasPrice			uint64		`json:"gas_price"`
}

func defaultParams() *Params {
	return &Params{
		HoldAccount:          HoldAccount,
		MaxVals:              100,
		Validators:           "",
		SelfStakingRatio:     "0.1",
		InflationRate:        8,
		StakeLimit:           "0.12",
		UnstakeWaitPeriod:    7 * 24 * 3600 / 10,
		ProposalExpirePeriod: 7 * 24 * 3600 / 10,
		DeclareCandidacy: 		1e6,
		UpdateCandidacy: 		1e6,
		GovernancePropose: 		2e6,
		GasPrice: 				2e9,
	}
}

var (
	// Keys for store prefixes
	ParamKey = []byte{0x01} // key for global parameters
	params = defaultParams()
)

// load/save the global params
func LoadParams(b []byte) {
	types.Cdc.UnmarshalBinary(b, params)
}

func UnloadParams() (b []byte) {
	b, _ = types.Cdc.MarshalBinary(*params)
	return
}

func GetParams() *Params {
	return params
}
