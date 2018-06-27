package utils

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/CyberMiles/travis/sdk/state"
	"github.com/CyberMiles/travis/types"
)

type Params struct {
	HoldAccount      common.Address `json:"hold_account"` // PubKey where all bonded coins are held
	MaxVals          uint16         `json:"max_vals"`     // maximum number of validators
	Validators       string         `json:"validators"`   // initial validators definition
	SelfStakingRatio string         `json:"self_staking_ratio"`
}

func defaultParams() *Params {
	return &Params{
		HoldAccount:      HoldAccount,
		MaxVals:          100,
		Validators:       "",
		SelfStakingRatio: "0.1",
	}
}

var (
	// Keys for store prefixes
	ParamKey = []byte{0x01} // key for global parameters
)

// load/save the global params
func LoadParams(store state.SimpleDB) (params *Params) {
	b := store.Get(ParamKey)
	if b == nil {
		return defaultParams()
	}

	params = new(Params)

	err := types.Cdc.UnmarshalBinary(b, params)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}

	return
}

func SaveParams(store state.SimpleDB, params *Params) {
	b, _ := types.Cdc.MarshalBinary(*params)
	store.Set(ParamKey, b)
}

