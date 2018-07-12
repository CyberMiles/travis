package stake

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/types"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with its validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	ByteTxDeclareCandidacy  = 0x55
	ByteTxUpdateCandidacy   = 0x56
	ByteTxWithdrawCandidacy = 0x57
	ByteTxVerifyCandidacy   = 0x58
	ByteTxActivateCandidacy = 0x59
	ByteTxDelegate          = 0x60
	ByteTxWithdraw          = 0x61
	TypeTxDeclareCandidacy  = stakingModuleName + "/declareCandidacy"
	TypeTxUpdateCandidacy   = stakingModuleName + "/updateCandidacy"
	TypeTxVerifyCandidacy   = stakingModuleName + "/verifyCandidacy"
	TypeTxWithdrawCandidacy = stakingModuleName + "/withdrawCandidacy"
	TypeTxActivateCandidacy = stakingModuleName + "/activateCandidacy"
	TypeTxDelegate          = stakingModuleName + "/delegate"
	TypeTxWithdraw          = stakingModuleName + "/withdraw"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxDeclareCandidacy{}, TypeTxDeclareCandidacy, ByteTxDeclareCandidacy)
	sdk.TxMapper.RegisterImplementation(TxUpdateCandidacy{}, TypeTxUpdateCandidacy, ByteTxUpdateCandidacy)
	sdk.TxMapper.RegisterImplementation(TxWithdrawCandidacy{}, TypeTxWithdrawCandidacy, ByteTxWithdrawCandidacy)
	sdk.TxMapper.RegisterImplementation(TxVerifyCandidacy{}, TypeTxVerifyCandidacy, ByteTxVerifyCandidacy)
	sdk.TxMapper.RegisterImplementation(TxActivateCandidacy{}, TypeTxActivateCandidacy, ByteTxActivateCandidacy)
	sdk.TxMapper.RegisterImplementation(TxDelegate{}, TypeTxDelegate, ByteTxDelegate)
	sdk.TxMapper.RegisterImplementation(TxWithdraw{}, TypeTxWithdraw, ByteTxWithdraw)
}

//Verify interface at compile time
var _, _, _, _, _, _ sdk.TxInner = &TxDeclareCandidacy{}, &TxUpdateCandidacy{}, &TxWithdrawCandidacy{}, TxVerifyCandidacy{}, &TxDelegate{}, &TxWithdraw{}

type TxDeclareCandidacy struct {
	PubKey      string      `json:"pub_key"`
	MaxAmount   string      `json:"max_amount"`
	CompRate    string      `json:"comp_rate"`
	Description Description `json:"description"`
}

func (tx TxDeclareCandidacy) ValidateBasic() error {
	return nil
}

func (tx TxDeclareCandidacy) SelfStakingAmount(ratio string) (amount *big.Int) {
	amount = new(big.Int)
	maxAmount, _ := new(big.Float).SetString(tx.MaxAmount)
	z := new(big.Float)
	r, _ := new(big.Float).SetString(ratio)
	z.Mul(maxAmount, r)
	z.Int(amount)
	return
}

func NewTxDeclareCandidacy(pubKey types.PubKey, maxAmount, compRate string, descrpition Description) sdk.Tx {
	return TxDeclareCandidacy{
		PubKey:      types.PubKeyString(pubKey),
		MaxAmount:   maxAmount,
		CompRate:    compRate,
		Description: descrpition,
	}.Wrap()
}

func (tx TxDeclareCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxUpdateCandidacy struct {
	MaxAmount   string      `json:"max_amount"`
	Description Description `json:"description"`
}

func (tx TxUpdateCandidacy) ValidateBasic() error {
	return nil
}

func NewTxUpdateCandidacy(maxAmount string, description Description) sdk.Tx {
	return TxUpdateCandidacy{
		MaxAmount:   maxAmount,
		Description: description,
	}.Wrap()
}

func (tx TxUpdateCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxVerifyCandidacy struct {
	CandidateAddress common.Address `json:"candidate_address"`
	Verified         bool           `json:"verified"`
}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (tx TxVerifyCandidacy) ValidateBasic() error {
	return nil
}

func NewTxVerifyCandidacy(candidateAddress common.Address, verified bool) sdk.Tx {
	return TxVerifyCandidacy{
		CandidateAddress: candidateAddress,
		Verified:         verified,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxVerifyCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxWithdrawCandidacy struct{}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (tx TxWithdrawCandidacy) ValidateBasic() error {
	return nil
}

func NewTxWithdrawCandidacy() sdk.Tx {
	return TxWithdrawCandidacy{}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxWithdrawCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxActivateCandidacy struct{}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (tx TxActivateCandidacy) ValidateBasic() error {
	return nil
}

func NewTxActivateCandidacy() sdk.Tx {
	return TxActivateCandidacy{}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxActivateCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

// TxDelegate - struct for bonding or unbonding transactions
type TxDelegate struct {
	ValidatorAddress common.Address `json:"validator_address"`
	Amount           string         `json:"amount"`
	CubeBatch        string         `json:"cube_batch"`
	Sig              string         `json:"sig"`
}

func (tx TxDelegate) ValidateBasic() error {
	return nil
}

func NewTxDelegate(validatorAddress common.Address, amount, cubeBatch, sig string) sdk.Tx {
	return TxDelegate{
		ValidatorAddress: validatorAddress,
		Amount:           amount,
		CubeBatch:        cubeBatch,
		Sig:              sig,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxDelegate) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxWithdraw struct {
	ValidatorAddress common.Address `json:"validator_address"`
	Amount           string         `json:"amount"`
}

func (tx TxWithdraw) ValidateBasic() error {
	return nil
}

func NewTxWithdraw(validatorAddress common.Address, amount string) sdk.Tx {
	return TxWithdraw{
		ValidatorAddress: validatorAddress,
		Amount:           amount,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxWithdraw) Wrap() sdk.Tx { return sdk.Tx{tx} }
