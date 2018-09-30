package stake

import (
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
	ByteTxDeclareCandidacy             = 0x55
	ByteTxUpdateCandidacy              = 0x56
	ByteTxWithdrawCandidacy            = 0x57
	ByteTxVerifyCandidacy              = 0x58
	ByteTxActivateCandidacy            = 0x59
	ByteTxDelegate                     = 0x60
	ByteTxWithdraw                     = 0x61
	ByteTxSetCompRate                  = 0x62
	ByteTxUpdateCandidacyAccount       = 0x63
	ByteTxAcceptCandidacyAccountUpdate = 0x64
	TypeTxDeclareCandidacy             = "stake/declareCandidacy"
	TypeTxUpdateCandidacy              = "stake/updateCandidacy"
	TypeTxVerifyCandidacy              = "stake/verifyCandidacy"
	TypeTxWithdrawCandidacy            = "stake/withdrawCandidacy"
	TypeTxActivateCandidacy            = "stake/activateCandidacy"
	TypeTxDelegate                     = "stake/delegate"
	TypeTxWithdraw                     = "stake/withdraw"
	TypeTxSetCompRate                  = "stake/setCompRate"
	TypeTxUpdateCandidacyAccount       = "stake/updateCandidacyAccount"
	TypeTxAcceptCandidacyAccountUpdate = "stake/acceptCandidacyAccountUpdate"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxDeclareCandidacy{}, TypeTxDeclareCandidacy, ByteTxDeclareCandidacy)
	sdk.TxMapper.RegisterImplementation(TxUpdateCandidacy{}, TypeTxUpdateCandidacy, ByteTxUpdateCandidacy)
	sdk.TxMapper.RegisterImplementation(TxWithdrawCandidacy{}, TypeTxWithdrawCandidacy, ByteTxWithdrawCandidacy)
	sdk.TxMapper.RegisterImplementation(TxVerifyCandidacy{}, TypeTxVerifyCandidacy, ByteTxVerifyCandidacy)
	sdk.TxMapper.RegisterImplementation(TxActivateCandidacy{}, TypeTxActivateCandidacy, ByteTxActivateCandidacy)
	sdk.TxMapper.RegisterImplementation(TxDelegate{}, TypeTxDelegate, ByteTxDelegate)
	sdk.TxMapper.RegisterImplementation(TxWithdraw{}, TypeTxWithdraw, ByteTxWithdraw)
	sdk.TxMapper.RegisterImplementation(TxSetCompRate{}, TypeTxSetCompRate, ByteTxSetCompRate)
	sdk.TxMapper.RegisterImplementation(TxUpdateCandidacyAccount{}, TypeTxUpdateCandidacyAccount, ByteTxUpdateCandidacyAccount)
	sdk.TxMapper.RegisterImplementation(TxAcceptCandidacyAccountUpdate{}, TypeTxAcceptCandidacyAccountUpdate, ByteTxAcceptCandidacyAccountUpdate)
}

//Verify interface at compile time
var _, _, _, _, _, _, _, _, _, _ sdk.TxInner = &TxDeclareCandidacy{}, &TxUpdateCandidacy{}, &TxWithdrawCandidacy{}, TxVerifyCandidacy{}, &TxActivateCandidacy{}, &TxDelegate{}, &TxWithdraw{}, &TxSetCompRate{}, &TxUpdateCandidacyAccount{}, &TxAcceptCandidacyAccountUpdate{}

type TxDeclareCandidacy struct {
	PubKey      string      `json:"pub_key"`
	MaxAmount   string      `json:"max_amount"`
	CompRate    sdk.Rat     `json:"comp_rate"`
	Description Description `json:"description"`
}

func (tx TxDeclareCandidacy) ValidateBasic() error {
	return nil
}

func (tx TxDeclareCandidacy) SelfStakingAmount(ssr sdk.Rat) (res sdk.Int) {
	maxAmount, _ := sdk.NewIntFromString(tx.MaxAmount)
	res = maxAmount.MulRat(ssr)
	return
}

func NewTxDeclareCandidacy(pubKey types.PubKey, maxAmount string, compRate sdk.Rat, descrpition Description) sdk.Tx {
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

type TxSetCompRate struct {
	DelegatorAddress common.Address `json:"delegator_address"`
	CompRate         sdk.Rat        `json:"comp_rate"`
}

func (tx TxSetCompRate) ValidateBasic() error {
	return nil
}

func NewTxSetCompRate(delegatorAddress common.Address, compRate sdk.Rat) sdk.Tx {
	return TxSetCompRate{
		DelegatorAddress: delegatorAddress,
		CompRate:         compRate,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxSetCompRate) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxUpdateCandidacyAccount struct {
	NewCandidateAddress common.Address `json:"new_candidate_account"`
}

func (tx TxUpdateCandidacyAccount) ValidateBasic() error {
	return nil
}

func NewTxUpdateCandidacyAccount(newCandidateAddress common.Address) sdk.Tx {
	return TxUpdateCandidacyAccount{
		NewCandidateAddress: newCandidateAddress,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxUpdateCandidacyAccount) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxAcceptCandidacyAccountUpdate struct {
	AccountUpdateRequestId int64 `json:"account_update_request_id"`
}

func (tx TxAcceptCandidacyAccountUpdate) ValidateBasic() error {
	return nil
}

func NewTxAcceptCandidacyAccountUpdate(accountUpdateRequestId int64) sdk.Tx {
	return TxAcceptCandidacyAccountUpdate{
		accountUpdateRequestId,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxAcceptCandidacyAccountUpdate) Wrap() sdk.Tx { return sdk.Tx{tx} }
