package stake

import (
	"github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/state"
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

	err := wire.ReadBinaryBytes(b, &params)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}

	return
}
func saveParams(store state.SimpleDB, params Params) {
	b := wire.BinaryBytes(params)
	store.Set(ParamKey, b)
}
