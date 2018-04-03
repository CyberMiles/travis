package governance

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/ethereum/go-ethereum/common"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with its validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	ByteTxPropose      = 0xA1
	TypeTxPropose      = governanceModuleName + "/propose"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxPropose{}, TypeTxPropose, ByteTxPropose)
}

//Verify interface at compile time
var _ sdk.TxInner = &TxPropose{}

type TxPropose struct {
	Proposer     common.Address   `json:"proposer"`
	From         common.Address   `json:"from"`
	To           common.Address   `json:"to"`
	Amount       big.Int          `json:"amount"`
	Reason       string           `json:"reason"`
}

func (tx TxPropose) ValidateBasic() error {
	return nil
}

func NewTxPropose(proposer common.Address, fromAddr common.Address, toAddr common.Address, amount *big.Int, reason string) sdk.Tx {
	return TxPropose{
		proposer,
		fromAddr,
		toAddr,
		*amount,
		reason,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxPropose) Wrap() sdk.Tx { return sdk.Tx{tx} }