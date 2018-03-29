package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk"
	crypto "github.com/tendermint/go-crypto"
	"github.com/ethereum/go-ethereum/common"
)

// Tx
//--------------------------------------------------------------------------------

// register the tx type with its validation logic
// make sure to use the name of the handler as the prefix in the tx type,
// so it gets routed properly
const (
	ByteTxDeclare      = 0x55
	ByteTxWithdraw     = 0x56
	ByteTxProposeSlot  = 0x57
	ByteTxAcceptSlot   = 0x58
	ByteTxWithdrawSlot = 0x59
	ByteTxCancelSlot   = 0x60
	TypeTxDeclare      = stakingModuleName + "/declare"
	TypeTxWithdraw     = stakingModuleName + "/withdraw"
	TypeTxProposeSlot  = stakingModuleName + "/proposeSlot"
	TypeTxAcceptSlot   = stakingModuleName + "/acceptSlot"
	TypeTxWithdrawSlot = stakingModuleName + "/withdrawSlot"
	TypeTxCancelSlot   = stakingModuleName + "/cancelSlot"
)

func init() {
	sdk.TxMapper.RegisterImplementation(TxDeclare{}, TypeTxDeclare, ByteTxDeclare)
	sdk.TxMapper.RegisterImplementation(TxWithdraw{}, TypeTxWithdraw, ByteTxWithdraw)
	sdk.TxMapper.RegisterImplementation(TxProposeSlot{}, TypeTxProposeSlot, ByteTxProposeSlot)
	sdk.TxMapper.RegisterImplementation(TxAcceptSlot{}, TypeTxAcceptSlot, ByteTxAcceptSlot)
	sdk.TxMapper.RegisterImplementation(TxWithdrawSlot{}, TypeTxWithdrawSlot, ByteTxWithdrawSlot)
	sdk.TxMapper.RegisterImplementation(TxCancelSlot{}, TypeTxCancelSlot, ByteTxCancelSlot)
}

//Verify interface at compile time
var _, _ sdk.TxInner = &TxDeclare{}, &TxProposeSlot{}

type TxDeclare struct {
	PubKey crypto.PubKey `json:"pub_key"`
}

func (tx TxDeclare) ValidateBasic() error {
	if tx.PubKey.Empty() {
		return errCandidateEmpty
	}

	return nil
}

func NewTxDeclare(pubKey crypto.PubKey) sdk.Tx {
	return TxDeclare{
		PubKey: pubKey,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxDeclare) Wrap() sdk.Tx { return sdk.Tx{tx} }

type TxWithdraw struct {
	Address common.Address `json:"address"`
}

// ValidateBasic - Check for non-empty candidate, and valid coins
func (tx TxWithdraw) ValidateBasic() error {
	if len(tx.Address) == 0 {
		return errCandidateEmpty
	}

	return nil
}

func NewTxWithdraw(address common.Address) sdk.Tx {
	return TxWithdraw{
		Address: address,
	}.Wrap()
}

// Wrap - Wrap a Tx as a Basecoin Tx
func (tx TxWithdraw) Wrap() sdk.Tx { return sdk.Tx{tx} }

// TxProposeSlot - struct for propose slot
type TxProposeSlot struct {
	ValidatorAddress      	common.Address
	Amount      			int64
	ProposedRoi 			int64
}

// NewTxProposeSlot - new TxProposeSlot
func NewTxProposeSlot(validatorAddress common.Address, amount int64, proposedRoi int64) sdk.Tx {
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

type TxWithdrawSlot struct {
	SlotUpdate
}

func NewTxWithdrawSlot(amount int64, slotId string) sdk.Tx {
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