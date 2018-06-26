package governance

import (
	"github.com/CyberMiles/travis/sdk"
	"github.com/ethereum/go-ethereum/common"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with its validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	ByteTxTransferFundPropose      = 0xA1
	ByteTxChangeParamPropose       = 0xA2
	ByteTxDeployLibEniPropose      = 0xA3
	ByteTxVote                     = 0xA4
	TypeTxTransferFundPropose      = governanceModuleName + "/propose/transfer_fund"
	TypeTxChangeParamPropose       = governanceModuleName + "/propose/change_param"
	TypeTxVote         = governanceModuleName + "/vote"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxTransferFundPropose{}, TypeTxTransferFundPropose, ByteTxTransferFundPropose)
	sdk.TxMapper.RegisterImplementation(TxChangeParamPropose{}, TypeTxChangeParamPropose, ByteTxChangeParamPropose)
	sdk.TxMapper.RegisterImplementation(TxVote{}, TypeTxVote, ByteTxVote)
}

//Verify interface at compile time
var _, _, _ sdk.TxInner = &TxTransferFundPropose{}, &TxChangeParamPropose{}, &TxVote{}

type TxTransferFundPropose struct {
	Proposer     *common.Address   `json:"proposer"`
	From         *common.Address   `json:"from"`
	To           *common.Address   `json:"to"`
	Amount       string            `json:"amount"`
	Reason       string            `json:"reason"`
	Expire       uint64	           `json:"expire"`
}

func (tx TxTransferFundPropose) ValidateBasic() error {
	return nil
}

func NewTxTransferFundPropose(proposer *common.Address, fromAddr *common.Address, toAddr *common.Address, amount string, reason string, expire uint64) sdk.Tx {
	return TxTransferFundPropose{
		proposer,
		fromAddr,
		toAddr,
		amount,
		reason,
		expire,
	}.Wrap()
}

func (tx TxTransferFundPropose) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxChangeParamPropose struct {
	Proposer     *common.Address   `json:"proposer"`
	Name         string            `json:"name"`
	Value        string            `json:"value"`
	Reason       string            `json:"reason"`
}

func (tx TxChangeParamPropose) ValidateBasic() error {
	return nil
}

func NewTxChangeParamPropose(proposer *common.Address, name string, value string, reason string) sdk.Tx {
	return TxChangeParamPropose{
		proposer,
		name,
		value,
		reason,
	}.Wrap()
}

func (tx TxChangeParamPropose) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxVote struct {
	ProposalId       string            `json:"proposal_id"`
	Voter            common.Address    `json:"voter"`
	Answer           string            `json:"answer"`
}

func (tx TxVote) ValidateBasic() error {
	return nil
}

func NewTxVote(pid string, voter common.Address, answer string) sdk.Tx {
	return TxVote{
		pid,
		voter,
		answer,
	}.Wrap()
}

func (tx TxVote) Wrap() sdk.Tx { return sdk.Tx{tx} }
