package stake

import (
	"github.com/CyberMiles/travis/sdk/state"
	"github.com/CyberMiles/travis/types"
)

// nolint
var (
	// Keys for store prefixes
	ParamKey = []byte{0x01} // key for global parameters relating to staking
)

//---------------------------------------------------------------------

// load/save the global staking params
func loadParams(store state.SimpleDB) (params Params) {
	b := store.Get(ParamKey)
	if b == nil {
		return defaultParams()
	}

	err := types.Cdc.UnmarshalBinary(b, &params)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}

	return
}
func saveParams(store state.SimpleDB, params Params) {
	b, _ := types.Cdc.MarshalBinary(params)
	store.Set(ParamKey, b)
}
