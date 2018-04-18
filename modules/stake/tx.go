package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk"
	crypto "github.com/tendermint/go-crypto"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with its validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	ByteTxDeclareCandidacy = 0x55
	ByteTxEditCandidacy    = 0x56
	ByteTxWithdraw         = 0x57
	ByteTxProposeSlot      = 0x58
	ByteTxAcceptSlot       = 0x59
	ByteTxWithdrawSlot     = 0x60
	ByteTxCancelSlot       = 0x61
	TypeTxDeclareCandidacy = stakingModuleName + "/declareCandidacy"
	TypeTxEditCandidacy    = stakingModuleName + "/editCandidacy"
	TypeTxWithdraw         = stakingModuleName + "/withdrawCandidacy"
	TypeTxProposeSlot      = stakingModuleName + "/proposeSlot"
	TypeTxAcceptSlot       = stakingModuleName + "/acceptSlot"
	TypeTxWithdrawSlot     = stakingModuleName + "/withdrawSlot"
	TypeTxCancelSlot       = stakingModuleName + "/cancelSlot"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxDeclareCandidacy{}, TypeTxDeclareCandidacy, ByteTxDeclareCandidacy)
	sdk.TxMapper.RegisterImplementation(TxEditCandidacy{}, TypeTxEditCandidacy, ByteTxEditCandidacy)
	sdk.TxMapper.RegisterImplementation(TxWithdrawCandidacy{}, TypeTxWithdraw, ByteTxWithdraw)
	sdk.TxMapper.RegisterImplementation(TxProposeSlot{}, TypeTxProposeSlot, ByteTxProposeSlot)
	sdk.TxMapper.RegisterImplementation(TxAcceptSlot{}, TypeTxAcceptSlot, ByteTxAcceptSlot)
	sdk.TxMapper.RegisterImplementation(TxWithdrawSlot{}, TypeTxWithdrawSlot, ByteTxWithdrawSlot)
	sdk.TxMapper.RegisterImplementation(TxCancelSlot{}, TypeTxCancelSlot, ByteTxCancelSlot)
}

//Verify interface at compile time
var _, _, _, _, _, _, _, _ sdk.TxInner = &TxDeclareCandidacy{}, &TxProposeSlot{}, &TxEditCandidacy{}, &TxWithdrawCandidacy{}, &TxProposeSlot{}, &TxAcceptSlot{}, &TxCancelSlot{}, &TxWithdrawSlot{}

type TxDeclareCandidacy struct {
	PubKey crypto.PubKey `json:"pub_key"`
}

func (tx TxDeclareCandidacy) ValidateBasic() error {
	if tx.PubKey.Empty() {
		return errCandidateEmpty
	}

	return nil
}

func NewTxDeclareCandidacy(pubKey crypto.PubKey) sdk.Tx {
	return TxDeclareCandidacy{
		PubKey: pubKey,
	}.Wrap()
}

func (tx TxDeclareCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxEditCandidacy struct {
	NewAddress common.Address `json:"new_address"`
}

func (tx TxEditCandidacy) ValidateBasic() error {
	if len(tx.NewAddress) == 0 {
		return errCandidateEmpty
	}

	return nil
}

func NewTxEditCandidacy(newAddress common.Address) sdk.Tx {
	return TxEditCandidacy{
		NewAddress: newAddress,
	}.Wrap()
}

func (tx TxEditCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxWithdrawCandidacy struct {
	Address common.Address `json:"address"`
}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (tx TxWithdrawCandidacy) ValidateBasic() error {
	if len(tx.Address) == 0 {
		return errCandidateEmpty
	}

	return nil
}

func NewTxWithdrawCandidacy(address common.Address) sdk.Tx {
	return TxWithdrawCandidacy{
		Address: address,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxWithdrawCandidacy) Wrap() sdk.Tx { return sdk.Tx{tx} }

// TxProposeSlot - struct for propose slot
type TxProposeSlot struct {
	ValidatorAddress      	common.Address
	Amount      			string
	ProposedRoi 			int64
}

// NewTxProposeSlot - new TxProposeSlot
func NewTxProposeSlot(validatorAddress common.Address, amount string, proposedRoi int64) sdk.Tx {
	return TxProposeSlot{
		ValidatorAddress:      	validatorAddress,
		Amount:      			amount,
		ProposedRoi: 			proposedRoi,
	}.Wrap()
}

// ValidateBasic - Check for non-empty candidate, positive shares
func (tx TxProposeSlot) ValidateBasic() error {
	if len(tx.ValidatorAddress) == 0 {
		return errCandidateEmpty
	}

	amount := new(big.Int)
	_, ok := amount.SetString(tx.Amount, 10)
	if !ok || amount.Cmp(big.NewInt(0)) <= 0 {
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
	Amount string
	SlotId string
}

func (tx SlotUpdate) ValidateBasic() error {
	return nil
}

type TxAcceptSlot struct {
	SlotUpdate
}

func NewTxAcceptSlot(amount string, slotId string) sdk.Tx {
	return TxAcceptSlot{ SlotUpdate{
			Amount: amount,
			SlotId: slotId,
		}}.Wrap()
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxAcceptSlot) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxWithdrawSlot struct {
	SlotUpdate
}

func NewTxWithdrawSlot(amount string, slotId string) sdk.Tx {
	return TxWithdrawSlot{ SlotUpdate{
		Amount: amount,
		SlotId: slotId,
	}}.Wrap()
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxWithdrawSlot) Wrap() sdk.Tx { return sdk.Tx{tx} }

// TxProposeSlot - struct for propose slot
type TxCancelSlot struct {
	ValidatorAddress	common.Address
	SlotId				string
}

// NewTxProposeSlot - new TxProposeSlot
func NewTxCancelSlot(validatorAddress common.Address, slotId string) sdk.Tx {
	return TxCancelSlot{
		ValidatorAddress: validatorAddress,
		SlotId:	slotId,
	}.Wrap()
}

// ValidateBasic - Check for non-empty candidate, positive shares
func (tx TxCancelSlot) ValidateBasic() error {
	if len(tx.ValidatorAddress) == 0 {
		return errCandidateEmpty
	}

	if tx.SlotId == "" {
		return fmt.Errorf("slot must be provided")
	}

	return nil
}

// Wrap - Wrap a Tx as a Travis Tx
func (tx TxCancelSlot) Wrap() sdk.Tx { return sdk.Tx{tx} }