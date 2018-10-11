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
	ByteTxRetireProgramPropose     = 0xA4
	ByteTxUpgradeProgramPropose    = 0xA5
	ByteTxVote                     = 0xA6
	TypeTxTransferFundPropose      = governanceModuleName + "/propose/transfer_fund"
	TypeTxChangeParamPropose       = governanceModuleName + "/propose/change_param"
	TypeTxDeployLibEniPropose      = governanceModuleName + "/propose/deploy_libeni"
	TypeTxRetireProgramPropose     = governanceModuleName + "/propose/retire_program"
	TypeTxUpgradeProgramPropose    = governanceModuleName + "/propose/upgrade_program"
	TypeTxVote                     = governanceModuleName + "/vote"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxTransferFundPropose{}, TypeTxTransferFundPropose, ByteTxTransferFundPropose)
	sdk.TxMapper.RegisterImplementation(TxChangeParamPropose{}, TypeTxChangeParamPropose, ByteTxChangeParamPropose)
	sdk.TxMapper.RegisterImplementation(TxDeployLibEniPropose{}, TypeTxDeployLibEniPropose, ByteTxDeployLibEniPropose)
	sdk.TxMapper.RegisterImplementation(TxRetireProgramPropose{}, TypeTxRetireProgramPropose, ByteTxRetireProgramPropose)
	sdk.TxMapper.RegisterImplementation(TxUpgradeProgramPropose{}, TypeTxUpgradeProgramPropose, ByteTxUpgradeProgramPropose)
	sdk.TxMapper.RegisterImplementation(TxVote{}, TypeTxVote, ByteTxVote)
}

//Verify interface at compile time
var _, _, _, _, _ sdk.TxInner = &TxTransferFundPropose{}, &TxChangeParamPropose{}, &TxDeployLibEniPropose{}, &TxRetireProgramPropose{}, &TxUpgradeProgramPropose{}
var _ sdk.TxInner = &TxVote{}

type TxTransferFundPropose struct {
	From               *common.Address   `json:"transfer_from"`
	To                 *common.Address   `json:"transfer_to"`
	Amount             string            `json:"amount"`
	Reason             string            `json:"reason"`
	ExpireTimestamp    *int64            `json:"expire_timestamp"`
	ExpireBlockHeight  *int64            `json:"expire_block_height"`
}

func (tx TxTransferFundPropose) ValidateBasic() error {
	return nil
}

func NewTxTransferFundPropose(fromAddr *common.Address, toAddr *common.Address, amount string, reason string, expireTimestamp, expireBlockHeight *int64) sdk.Tx {
	return TxTransferFundPropose{
		fromAddr,
		toAddr,
		amount,
		reason,
		expireTimestamp,
		expireBlockHeight,
	}.Wrap()
}

func (tx TxTransferFundPropose) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxChangeParamPropose struct {
	Name                  string   `json:"name"`
	Value                 string   `json:"value"`
	Reason                string   `json:"reason"`
	ExpireTimestamp       *int64   `json:"expire_timestamp"`
	ExpireBlockHeight     *int64   `json:"expire_block_height"`
}

func (tx TxChangeParamPropose) ValidateBasic() error {
	return nil
}

func NewTxChangeParamPropose(name string, value string, reason string, expireTimestamp, expireBlockHeight *int64) sdk.Tx {
	return TxChangeParamPropose{
		name,
		value,
		reason,
		expireTimestamp,
		expireBlockHeight,
	}.Wrap()
}

func (tx TxChangeParamPropose) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxDeployLibEniPropose struct {
	Name                  string   `json:"name"`
	Version               string   `json:"version"`
	FileUrl               string   `json:"file_url"`
	Md5                   string   `json:"md5"`
	Reason                string   `json:"reason"`
	ExpireTimestamp       *int64   `json:"deploy_timestamp"`
	ExpireBlockHeight     *int64   `json:"deploy_block_height"`
}

func (tx TxDeployLibEniPropose) ValidateBasic() error {
	return nil
}

func NewTxDeployLibEniPropose(name, version, fileurl, md5, reason string, expireTimestamp, expireBlockHeight *int64) sdk.Tx {
	return TxDeployLibEniPropose {
		name,
		version,
		fileurl,
		md5,
		reason,
		expireTimestamp,
		expireBlockHeight,
	}.Wrap()
}

func (tx TxDeployLibEniPropose) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxRetireProgramPropose struct {
	PreservedValidators string          `json:"preserved_validators"`
	Reason              string          `json:"reason"`
	ExpireBlockHeight   *int64          `json:"retired_block_height"`
}

func (tx TxRetireProgramPropose) ValidateBasic() error {
	return nil
}

func NewTxRetireProgramPropose(preservedValidators, reason string, expireBlockHeight *int64) sdk.Tx {
	return TxRetireProgramPropose {
		preservedValidators,
		reason,
		expireBlockHeight,
	}.Wrap()
}

func (tx TxRetireProgramPropose) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxUpgradeProgramPropose struct {
	Name               string          `json:"name"`
	Version            string          `json:"version"`
	FileUrl            string          `json:"file_url"`
	Md5                string          `json:"md5"`
	Reason             string          `json:"reason"`
	ExpireBlockHeight  *int64          `json:"upgrade_block_height"`
}

func (tx TxUpgradeProgramPropose) ValidateBasic() error {
	return nil
}

func NewTxUpgradeProgramPropose(name, version, fileurl, md5, reason string, expireBlockHeight *int64) sdk.Tx {
	return TxUpgradeProgramPropose {
		name,
		version,
		fileurl,
		md5,
		reason,
		expireBlockHeight,
	}.Wrap()
}

func (tx TxUpgradeProgramPropose) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxVote struct {
	ProposalId       string            `json:"proposal_id"`
	Answer           string            `json:"answer"`
}

func (tx TxVote) ValidateBasic() error {
	return nil
}

func NewTxVote(pid string, answer string) sdk.Tx {
	return TxVote{
		pid,
		answer,
	}.Wrap()
}

func (tx TxVote) Wrap() sdk.Tx { return sdk.Tx{tx} }
