package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk"
	crypto "github.com/tendermint/go-crypto"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with its validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	ByteTxDeclareCandidacy = 0x55
	ByteTxProposeSlot 	   = 0x56
	ByteTxAcceptSlot       = 0x57
	ByteTxWidthdrawSlot    = 0x58
	ByteTxCancelSlot       = 0x59
	TypeTxDeclareCandidacy = stakingModuleName + "/declareCandidacy"
	TypeTxProposeSlot      = stakingModuleName + "/proposeSlot"
	TypeTxAcceptSlot       = stakingModuleName + "/acceptSlot"
	TypeTxWidthdrawSlot    = stakingModuleName + "/widthdrawSlot"
	TypeTxCancelSlot       = stakingModuleName + "/cancelSlot"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxDeclareCandidacy{}, TypeTxDeclareCandidacy, ByteTxDeclareCandidacy)
	sdk.TxMapper.RegisterImplementation(TxProposeSlot{}, TypeTxProposeSlot, ByteTxProposeSlot)
	sdk.TxMapper.RegisterImplementation(TxAcceptSlot{}, TypeTxAcceptSlot, ByteTxAcceptSlot)
	sdk.TxMapper.RegisterImplementation(TxWidthdrawSlot{}, TypeTxWidthdrawSlot, ByteTxWidthdrawSlot)
	sdk.TxMapper.RegisterImplementation(TxCancelSlot{}, TypeTxCancelSlot, ByteTxCancelSlot)
}

//Verify interface at compile time
var _, _ sdk.TxInner = &TxDeclareCandidacy{}, &TxProposeSlot{}

// TxDeclareCandidacy - struct for unbonding transactions
type TxDeclareCandidacy struct {
	PubKey crypto.PubKey `json:"pub_key"`
}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (tx TxDeclareCandidacy) ValidateBasic() error {
	if tx.PubKey.Empty() {
		return errCandidateEmpty
	}

	return nil
}

// NewTxDeclareCandidacy - new TxDeclareCandidacy
func NewTxDeclareCandidacy(pubKey crypto.PubKey) sdk.Tx {
	return TxDeclareCandidacy{
		PubKey: pubKey,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxDeclareCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

// TxProposeSlot - struct for propose slot
type TxProposeSlot struct {
	PubKey      crypto.PubKey
	Amount      int64
	ProposedRoi int64
}

// NewTxProposeSlot - new TxProposeSlot
func NewTxProposeSlot(pubKey crypto.PubKey, amount int64, proposedRoi int64) sdk.Tx {
	return TxProposeSlot{
		PubKey:      pubKey,
		Amount:      amount,
		ProposedRoi: proposedRoi,
	}.Wrap()
}

// ValidateBasic - Check for non-empty candidate, positive shares
func (tx TxProposeSlot) ValidateBasic() error {
	if tx.PubKey.Empty() {
		return errCandidateEmpty
	}

	if tx.Amount <= 0 {
		return fmt.Errorf("amount must be positive interger")
	}

	if tx.ProposedRoi <= 0 {
		return fmt.Errorf("proposed ROI must be positive interger")
	}
	return nil
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxProposeSlot) Wrap() sdk.Tx { return sdk.Tx{tx} }

// SlotUpdate - struct for bonding or unbonding transactions
type SlotUpdate struct {
	Amount int64
	SlotId string
}

func (tx SlotUpdate) ValidateBasic() error {
	return nil
}

type TxAcceptSlot struct {
	SlotUpdate
}

func NewTxAcceptSlot(amount int64, slotId string) sdk.Tx {
	return TxAcceptSlot{ SlotUpdate{
			Amount: amount,
			SlotId: slotId,
		}}.Wrap()
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxAcceptSlot) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxWidthdrawSlot struct {
	SlotUpdate
}

func NewTxWidthdrawSlot(amount int64, slotId string) sdk.Tx {
	return TxWidthdrawSlot{ SlotUpdate{
		Amount: amount,
		SlotId: slotId,
	}}.Wrap()
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxWidthdrawSlot) Wrap() sdk.Tx { return sdk.Tx{tx} }

// TxProposeSlot - struct for propose slot
type TxCancelSlot struct {
	PubKey      crypto.PubKey
	SlotId		string
}

// NewTxProposeSlot - new TxProposeSlot
func NewTxCancelSlot(pubKey crypto.PubKey, slotId string) sdk.Tx {
	return TxCancelSlot{
		PubKey: pubKey,
		SlotId:	slotId,
	}.Wrap()
}

// ValidateBasic - Check for non-empty candidate, positive shares
func (tx TxCancelSlot) ValidateBasic() error {
	if tx.PubKey.Empty() {
		return errCandidateEmpty
	}

	if tx.SlotId == "" {
		return fmt.Errorf("slot must be provided")
	}

	return nil
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxCancelSlot) Wrap() sdk.Tx { return sdk.Tx{tx} }