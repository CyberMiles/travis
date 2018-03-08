package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/CyberMiles/travis/modules/coin"
	crypto "github.com/tendermint/go-crypto"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with its validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	ByteTxDeclareCandidacy = 0x55
	ByteTxEditCandidacy    = 0x56
	ByteTxDelegate         = 0x57
	ByteTxUnbond           = 0x58
	ByteTxProposeSlot      = 0x59
	TypeTxDeclareCandidacy = stakingModuleName + "/declareCandidacy"
	TypeTxEditCandidacy    = stakingModuleName + "/editCandidacy"
	TypeTxDelegate         = stakingModuleName + "/delegate"
	TypeTxUnbond           = stakingModuleName + "/unbond"
	TypeTxProposeSlot      = stakingModuleName + "/proposeSlot"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxDeclareCandidacy{}, TypeTxDeclareCandidacy, ByteTxDeclareCandidacy)
	sdk.TxMapper.RegisterImplementation(TxEditCandidacy{}, TypeTxEditCandidacy, ByteTxEditCandidacy)
	sdk.TxMapper.RegisterImplementation(TxDelegate{}, TypeTxDelegate, ByteTxDelegate)
	sdk.TxMapper.RegisterImplementation(TxUnbond{}, TypeTxUnbond, ByteTxUnbond)
	sdk.TxMapper.RegisterImplementation(TxProposeSlot{}, TypeTxProposeSlot, ByteTxProposeSlot)
}

//Verify interface at compile time
var _, _, _, _, _ sdk.TxInner = &TxDeclareCandidacy{}, &TxEditCandidacy{}, &TxDelegate{}, &TxUnbond{}, &TxProposeSlot{}

// BondUpdate - struct for bonding or unbonding transactions
type BondUpdate struct {
	PubKey crypto.PubKey `json:"pub_key"`
	Bond   coin.Coin     `json:"amount"`
}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (tx BondUpdate) ValidateBasic() error {
	if tx.PubKey.Empty() {
		return errCandidateEmpty
	}

	coins := coin.Coins{tx.Bond}
	if !coins.IsValid() {
		return coin.ErrInvalidCoins()
	}
	if !coins.IsPositive() {
		return fmt.Errorf("Amount must be > 0")
	}
	return nil
}

// TxDeclareCandidacy - struct for unbonding transactions
type TxDeclareCandidacy struct {
	BondUpdate
	Description
}

// NewTxDeclareCandidacy - new TxDeclareCandidacy
func NewTxDeclareCandidacy(bond coin.Coin, pubKey crypto.PubKey, description Description) sdk.Tx {
	return TxDeclareCandidacy{
		BondUpdate{
			PubKey: pubKey,
			Bond:   bond,
		},
		description,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxDeclareCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

// TxProposeSlot - struct for propose slot
type TxProposeSlot struct {
	PubKey crypto.PubKey
	OfferAmount int64
	Roi float64
}

// NewTxProposeSlot - new TxProposeSlot
func NewTxProposeSlot(pubKey crypto.PubKey, offerAmount int64, roi float64) sdk.Tx {
	return TxProposeSlot{
		PubKey: pubKey,
		OfferAmount: offerAmount,
		Roi: roi,
	}.Wrap()
}

// ValidateBasic - Check for non-empty candidate, positive shares
func (tx TxProposeSlot) ValidateBasic() error {
	if tx.PubKey.Empty() {
		return errCandidateEmpty
	}

	if tx.OfferAmount <= 0 {
		return fmt.Errorf("Offer amount must be positive interger")
	}

	if tx.Roi > 1 || tx.Roi <= 0 {
		return fmt.Errorf("ROI must between 0 and 1")
	}
	return nil
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxProposeSlot) Wrap() sdk.Tx { return sdk.Tx{tx} }

// TxEditCandidacy - struct for editing a candidate
type TxEditCandidacy struct {
	PubKey crypto.PubKey `json:"pub_key"`
	Description
}

// NewTxEditCandidacy - new TxEditCandidacy
func NewTxEditCandidacy(pubKey crypto.PubKey, description Description) sdk.Tx {
	return TxEditCandidacy{
		PubKey:      pubKey,
		Description: description,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxEditCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

// ValidateBasic - Check for non-empty candidate,
func (tx TxEditCandidacy) ValidateBasic() error {
	if tx.PubKey.Empty() {
		return errCandidateEmpty
	}

	empty := Description{}
	if tx.Description == empty {
		return fmt.Errorf("Transaction must include some information to modify")
	}
	return nil
}

// TxDelegate - struct for bonding transactions
type TxDelegate struct{ BondUpdate }

// NewTxDelegate - new TxDelegate
func NewTxDelegate(bond coin.Coin, pubKey crypto.PubKey) sdk.Tx {
	return TxDelegate{BondUpdate{
		PubKey: pubKey,
		Bond:   bond,
	}}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxDelegate) Wrap() sdk.Tx { return sdk.Tx{tx} }

// TxUnbond - struct for unbonding transactions
type TxUnbond struct {
	PubKey crypto.PubKey `json:"pub_key"`
	Shares uint64        `json:"amount"`
}

// NewTxUnbond - new TxUnbond
func NewTxUnbond(shares uint64, pubKey crypto.PubKey) sdk.Tx {
	return TxUnbond{
		PubKey: pubKey,
		Shares: shares,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxUnbond) Wrap() sdk.Tx { return sdk.Tx{tx} }

// ValidateBasic - Check for non-empty candidate, positive shares
func (tx TxUnbond) ValidateBasic() error {
	if tx.PubKey.Empty() {
		return errCandidateEmpty
	}

	if tx.Shares == 0 {
		return fmt.Errorf("Shares must be > 0")
	}
	return nil
}
