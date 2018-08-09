package sdk

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"math/big"
	"github.com/tendermint/tendermint/libs/common"
)

const (
	// ModuleNameBase is the module name for internal functionality
	ModuleNameBase = "base"
	// ChainKey is the option key for setting the chain id
	ChainKey = "chain_id"
)

type Result interface {
	GetData() []byte
}

// CheckResult captures any non-error abci result
// to make sure people use error for error cases
type CheckResult struct {
	Data []byte
	Log  string
	// GasAllocated is the maximum units of work we allow this tx to perform
	GasAllocated int64
	// GasPayment is the total fees for this tx (or other source of payment)
	GasPayment int64
}

// NewCheck sets the gas used and the response data but no more info
// these are the most common info needed to be set by the Handler
func NewCheck(gasAllocated int64, log string) CheckResult {
	return CheckResult{
		GasAllocated: gasAllocated,
		Log:          log,
	}
}

func (c CheckResult) ToABCI() abci.ResponseCheckTx {
	return abci.ResponseCheckTx{
		Data: c.Data,
		Log:  c.Log,
		GasUsed:  c.GasAllocated,
	}
}

func (c CheckResult) GetData() []byte {
	return c.Data
}

// DeliverResult captures any non-error abci result
// to make sure people use error for error cases
type DeliverResult struct {
	Data    []byte
	Log     string
	Diff    []*abci.Validator
	GasUsed int64 // unused
	GasFee  *big.Int
}

func (d DeliverResult) ToABCI() abci.ResponseDeliverTx {
	var tags []common.KVPair
	if d.GasFee.Cmp(big.NewInt(0)) > 0 {
		tags = append(tags, common.KVPair{Key: []byte("GasFee"), Value: d.GasFee.Bytes()})
	}
	return abci.ResponseDeliverTx{
		Data: d.Data,
		Log:  d.Log,
		Tags: tags,
		GasUsed: d.GasUsed,

	}
}

func (d DeliverResult) GetData() []byte {
	return d.Data
}
