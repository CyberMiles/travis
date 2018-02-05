package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	txcmd "github.com/CyberMiles/travis/modules/txs"
	"github.com/cosmos/cosmos-sdk/modules/coin"

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
	FlagShares = "shares"

	FlagMoniker  = "moniker"
	FlagIdentity = "keybase-sig"
	FlagWebsite  = "website"
	FlagDetails  = "details"
)

// nolint
var (
	CmdDeclareCandidacy = &cobra.Command{
		Use:   "declare-candidacy",
		Short: "create new validator-candidate account and delegate some coins to it",
		RunE:  cmdDeclareCandidacy,
	}
	CmdEditCandidacy = &cobra.Command{
		Use:   "edit-candidacy",
		Short: "edit and existing validator-candidate account",
		RunE:  cmdEditCandidacy,
	}
	CmdDelegate = &cobra.Command{
		Use:   "delegate",
		Short: "delegate coins to an existing validator/candidate",
		RunE:  cmdDelegate,
	}
	CmdUnbond = &cobra.Command{
		Use:   "unbond",
		Short: "unbond coins from a validator/candidate",
		RunE:  cmdUnbond,
	}

	CmdDeclare = &cobra.Command{
		Use:   "declare",
		Short: "Allows a potential validator to declare its candidacy",
		RunE:  cmdDeclare,
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
	CmdCancelSlot = &cobra.Command{
		Use:   "cancel-slot",
		Short: "Cancel all remianing amounts from an unaccepted slot by its creator using the Slot ID",
		RunE:  cmdCancelSlot,
	}
)

func init() {

	// define the flags
	fsPk := flag.NewFlagSet("", flag.ContinueOnError)
	fsPk.String(FlagPubKey, "", "PubKey of the validator-candidate")

	fsAmount := flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount.String(FlagAmount, "1fermion", "Amount of coins to bond")

	fsShares := flag.NewFlagSet("", flag.ContinueOnError)
	fsShares.Int64(FlagShares, 0, "Amount of shares to unbond")

	fsCandidate := flag.NewFlagSet("", flag.ContinueOnError)
	fsCandidate.String(FlagMoniker, "", "validator-candidate name")
	fsCandidate.String(FlagIdentity, "", "optional keybase signature")
	fsCandidate.String(FlagWebsite, "", "optional website")
	fsCandidate.String(FlagDetails, "", "optional detailed description space")

	// add the flags
	CmdDelegate.Flags().AddFlagSet(fsPk)
	CmdDelegate.Flags().AddFlagSet(fsAmount)

	CmdUnbond.Flags().AddFlagSet(fsPk)
	CmdUnbond.Flags().AddFlagSet(fsShares)

	CmdDeclareCandidacy.Flags().AddFlagSet(fsPk)
	CmdDeclareCandidacy.Flags().AddFlagSet(fsAmount)
	CmdDeclareCandidacy.Flags().AddFlagSet(fsCandidate)

	CmdEditCandidacy.Flags().AddFlagSet(fsPk)
	CmdEditCandidacy.Flags().AddFlagSet(fsCandidate)


	CmdDeclare.Flags().AddFlagSet(fsPk)
	CmdDeclare.Flags().AddFlagSet(fsAmount)
	CmdDeclare.Flags().AddFlagSet(fsCandidate)
}

func cmdDeclareCandidacy(cmd *cobra.Command, args []string) error {
	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	if viper.GetString(FlagMoniker) == "" {
		return fmt.Errorf("please enter a moniker for the validator-candidate using --moniker")
	}

	description := stake.Description{
		Moniker:  viper.GetString(FlagMoniker),
		Identity: viper.GetString(FlagIdentity),
		Website:  viper.GetString(FlagWebsite),
		Details:  viper.GetString(FlagDetails),
	}

	tx := stake.NewTxDeclareCandidacy(amount, pk, description)
	return txcmd.DoTx(tx)
}

func cmdEditCandidacy(cmd *cobra.Command, args []string) error {

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	description := stake.Description{
		Moniker:  viper.GetString(FlagMoniker),
		Identity: viper.GetString(FlagIdentity),
		Website:  viper.GetString(FlagWebsite),
		Details:  viper.GetString(FlagDetails),
	}

	tx := stake.NewTxEditCandidacy(pk, description)
	return txcmd.DoTx(tx)
}

func cmdDelegate(cmd *cobra.Command, args []string) error {
	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	tx := stake.NewTxDelegate(amount, pk)
	return txcmd.DoTx(tx)
}

func cmdUnbond(cmd *cobra.Command, args []string) error {

	sharesRaw := viper.GetInt64(FlagShares)
	if sharesRaw <= 0 {
		return fmt.Errorf("shares must be positive interger")
	}
	shares := uint64(sharesRaw)

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	tx := stake.NewTxUnbond(shares, pk)
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

func cmdDeclare(cmd *cobra.Command, args []string) error {
	amount, err := coin.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	if viper.GetString(FlagMoniker) == "" {
		return fmt.Errorf("please enter a moniker for the validator-candidate using --moniker")
	}

	description := stake.Description{
		Moniker:  viper.GetString(FlagMoniker),
		Identity: viper.GetString(FlagIdentity),
		Website:  viper.GetString(FlagWebsite),
		Details:  viper.GetString(FlagDetails),
	}

	tx := stake.NewTxDeclareCandidacy(amount, pk, description)
	return txcmd.DoTx(tx)
}

func cmdProposeSlot(cmd *cobra.Command, args []string) error {
	// todo

	return nil
}

func cmdAcceptSlot(cmd *cobra.Command, args []string) error {
	// todo

	return nil
}

func cmdCancelSlot(cmd *cobra.Command, args []string) error {
	// todo

	return nil
}