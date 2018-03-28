package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	txcmd "github.com/CyberMiles/travis/client/commands/txs"
	"github.com/CyberMiles/travis/modules/stake"
)

/*
The stake/declare tx allows a potential validator to declare its candidacy. Signed by the validator.

* Validator address

The stake/slot/propose tx allows a potential validator to offer a slot of CMTs and corresponding ROI. It returns a tx ID. Signed by the validator.

* Validator address
* CMT amount
* Proposed ROI

The stake/slot/accept tx is used by a delegator to accept and stake CMTs for an ID. Signed by the user.

* Slot ID
* CMT amount
* Delegator address

The stake/slot/cancel tx is to cancel all remianing amounts from an unaccepted slot by its creator using the ID. Signed by the validator.

* Slot ID
* Validator address
 */

// nolint
const (
	FlagPubKey = "pubkey"
	FlagAmount = "amount"
	FlagProposedRoi = "proposed-roi"
	FlagSlotId = "slot-id"
	FlagAddress = "address"
)

// nolint
var (
	CmdDeclare = &cobra.Command{
		Use:   "declare",
		Short: "Allows a potential validator to declare its candidacy",
		RunE:  cmdDeclare,
	}
	CmdWithdraw = &cobra.Command{
		Use:   "withdraw",
		Short: "Allows a validator to withdraw",
		RunE:  cmdWithdraw,
	}
	CmdProposeSlot = &cobra.Command{
		Use:   "propose-slot",
		Short: "Allows a potential validator to offer a slot of CMTs and corresponding ROI",
		RunE:  cmdProposeSlot,
	}
	CmdAcceptSlot = &cobra.Command{
		Use:   "accept-slot",
		Short: "Accept and stake CMTs for an Slot ID",
		RunE:  cmdAcceptSlot,
	}
	CmdWithdrawSlot = &cobra.Command{
		Use:   "withdraw-slot",
		Short: "Withdraw staked CMTs from a validator",
		RunE:  cmdWithdrawSlot,
	}
	CmdCancelSlot = &cobra.Command{
		Use:   "cancel-slot",
		Short: "Cancel all remaining amounts from an unaccepted slot by its creator using the Slot ID",
		RunE:  cmdCancelSlot,
	}
)

func init() {

	// define the flags
	fsPk := flag.NewFlagSet("", flag.ContinueOnError)
	fsPk.String(FlagPubKey, "", "PubKey of the validator-candidate")

	fsAmount := flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount.Int64(FlagAmount, 0, "Amount of CMT")

	fsProposeSlot := flag.NewFlagSet("", flag.ContinueOnError)
	fsProposeSlot.Float64(FlagProposedRoi, 0, "corresponding ROI")

	fsSlot := flag.NewFlagSet("", flag.ContinueOnError)
	fsSlot.String(FlagSlotId, "", "Slot ID")

	// add the flags
	CmdDeclare.Flags().AddFlagSet(fsPk)

	CmdProposeSlot.Flags().AddFlagSet(fsAmount)
	CmdProposeSlot.Flags().AddFlagSet(fsProposeSlot)

	CmdAcceptSlot.Flags().AddFlagSet(fsSlot)
	CmdAcceptSlot.Flags().AddFlagSet(fsAmount)

	CmdWithdrawSlot.Flags().AddFlagSet(fsSlot)
	CmdWithdrawSlot.Flags().AddFlagSet(fsAmount)

	CmdCancelSlot.Flags().AddFlagSet(fsSlot)
}

func cmdDeclare(cmd *cobra.Command, args []string) error {
	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	tx := stake.NewTxDeclare(pk)
	return txcmd.DoTx(tx)
}

func cmdWithdraw(cmd *cobra.Command, args []string) error {
	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	tx := stake.NewTxWithdraw(pk)
	return txcmd.DoTx(tx)
}

// GetPubKey - create the pubkey from a pubkey string
func GetPubKey(pubKeyStr string) (pk crypto.PubKey, err error) {

	if len(pubKeyStr) == 0 {
		err = fmt.Errorf("must use --pubkey flag")
		return
	}
	if len(pubKeyStr) != 64 { //if len(pkBytes) != 32 {
		err = fmt.Errorf("pubkey must be Ed25519 hex encoded string which is 64 characters long")
		return
	}
	var pkBytes []byte
	pkBytes, err = hex.DecodeString(pubKeyStr)
	if err != nil {
		return
	}
	var pkEd crypto.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	pk = pkEd.Wrap()
	return
}

func cmdProposeSlot(cmd *cobra.Command, args []string) error {
	address := viper.GetString(FlagAddress)

	amount := viper.GetInt64(FlagAmount)
	if amount <= 0 {
		return fmt.Errorf("amount must be positive interger")
	}

	proposedRoi := viper.GetInt64(FlagProposedRoi)
	if proposedRoi <= 0 {
		return fmt.Errorf("proposed ROI must be positive interger")
	}

	tx := stake.NewTxProposeSlot(address, amount, proposedRoi)
	return txcmd.DoTx(tx)
}

func cmdAcceptSlot(cmd *cobra.Command, args []string) error {
	amount := viper.GetInt64(FlagAmount)
	if amount <= 0 {
		return fmt.Errorf("Amount must be positive interger")
	}

	slotId := viper.GetString(FlagSlotId)
	if slotId == "" {
		return fmt.Errorf("please enter slot ID using --slot-id")
	}

	tx := stake.NewTxAcceptSlot(amount, slotId)
	return txcmd.DoTx(tx)
}

func cmdWithdrawSlot(cmd *cobra.Command, args []string) error {
	amount := viper.GetInt64(FlagAmount)
	if amount <= 0 {
		return fmt.Errorf("Amount must be positive interger")
	}

	slotId := viper.GetString(FlagSlotId)
	if slotId == "" {
		return fmt.Errorf("please enter slot ID using --slot-id")
	}

	tx := stake.NewTxAcceptSlot(amount, slotId)
	return txcmd.DoTx(tx)
}

func cmdCancelSlot(cmd *cobra.Command, args []string) error {
	address := viper.GetString(FlagAddress)
	slotId := viper.GetString(FlagSlotId)
	if slotId == "" {
		return fmt.Errorf("please enter slot ID using --slot-id")
	}

	tx := stake.NewTxCancelSlot(address, slotId)
	return txcmd.DoTx(tx)
}