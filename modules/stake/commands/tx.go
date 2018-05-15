package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	txcmd "github.com/CyberMiles/travis/client/commands/txs"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/CyberMiles/travis/types"
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
	FlagPubKey           = "pubkey"
	FlagAmount           = "amount"
	FlagMaxAmount        = "max-amount"
	FlagCut              = "cut"
	FlagAddress          = "address"
	FlagNewAddress       = "new-address"
	FlagValidatorAddress = "validator-address"
	FlagWebsite          = "website"
	FlagLocation         = "location"
	FlagDetails          = "details"
	FlagVerified         = "verified"
)

// nolint
var (
	CmdDeclareCandidacy = &cobra.Command{
		Use:   "declare-candidacy",
		Short: "Allows a potential validator to declare its candidacy",
		RunE:  cmdDeclareCandidacy,
	}
	CmdUpdateCandidacy = &cobra.Command{
		Use:   "update-candidacy",
		Short: "Allows a validator candidate to change its candidacy",
		RunE:  cmdUpdateCandidacy,
	}
	CmdWithdrawCandidacy = &cobra.Command{
		Use:   "withdraw-candidacy",
		Short: "Allows a validator/candidate to withdraw",
		RunE:  cmdWithdrawCandidacy,
	}
	CmdVerifyCandidacy = &cobra.Command{
		Use:   "verify-candidacy",
		Short: "Allows the foundation to verify a validator/candidate's information",
		RunE:  cmdVerifyCandidacy,
	}
	CmdActivateCandidacy = &cobra.Command{
		Use:   "activate-candidacy",
		Short: "Allows a validator activate itself",
		RunE:  cmdActivateCandidacy,
	}
	CmdDelegate = &cobra.Command{
		Use:   "delegate",
		Short: "Delegate coins to an existing validator/candidate",
		RunE:  cmdDelegate,
	}
	CmdWithdraw = &cobra.Command{
		Use:   "withdraw",
		Short: "Withdraw coins from a validator/candidate",
		RunE:  cmdWithdraw,
	}
)

func init() {

	// define the flags
	fsPk := flag.NewFlagSet("", flag.ContinueOnError)
	fsPk.String(FlagPubKey, "", "PubKey of the validator-candidate")

	fsAmount := flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount.String(FlagAmount, "0", "Amount of CMTs")

	fsCandidate := flag.NewFlagSet("", flag.ContinueOnError)
	fsCandidate.String(FlagMaxAmount, "", "Max amount of CMTs to be staked")
	fsCandidate.String(FlagWebsite, "", "optional website")
	fsCandidate.String(FlagLocation, "", "optional location")
	fsCandidate.String(FlagDetails, "", "optional detailed description")

	fsCut := flag.NewFlagSet("", flag.ContinueOnError)
	fsCut.Int64(FlagCut, 0, "The percentage of block awards to be distributed back to the delegators")

	fsAddr := flag.NewFlagSet("", flag.ContinueOnError)
	fsAddr.String(FlagAddress, "", "Account address")

	fsNewAddress := flag.NewFlagSet("", flag.ContinueOnError)
	fsNewAddress.String(FlagNewAddress, "", "New account address")

	fsVerified := flag.NewFlagSet("", flag.ContinueOnError)
	fsVerified.String(FlagVerified, "false", "true or false")

	fsValidatorAddress := flag.NewFlagSet("", flag.ContinueOnError)
	fsValidatorAddress.String(FlagValidatorAddress, "", "validator address")

	// add the flags
	CmdDeclareCandidacy.Flags().AddFlagSet(fsPk)
	CmdDeclareCandidacy.Flags().AddFlagSet(fsCandidate)
	CmdDeclareCandidacy.Flags().AddFlagSet(fsCut)

	CmdUpdateCandidacy.Flags().AddFlagSet(fsNewAddress)
	CmdUpdateCandidacy.Flags().AddFlagSet(fsCandidate)

	CmdVerifyCandidacy.Flags().AddFlagSet(fsValidatorAddress)
	CmdVerifyCandidacy.Flags().AddFlagSet(fsVerified)

	CmdDelegate.Flags().AddFlagSet(fsValidatorAddress)
	CmdDelegate.Flags().AddFlagSet(fsAmount)

	CmdWithdraw.Flags().AddFlagSet(fsValidatorAddress)
	CmdWithdraw.Flags().AddFlagSet(fsAmount)
}

func cmdDeclareCandidacy(cmd *cobra.Command, args []string) error {
	pk, err := types.GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	maxAmount := viper.GetString(FlagMaxAmount)
	v := new(big.Int)
	_, ok := v.SetString(maxAmount, 10)
	if !ok || v.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("max-amount must be positive interger")
	}

	cut := viper.GetInt64(FlagCut)
	if cut <= 0 || cut >= 10000 {
		return fmt.Errorf("cut must between 0 and 1")
	}

	description := stake.Description{
		Website:  viper.GetString(FlagWebsite),
		Location: viper.GetString(FlagLocation),
		Details:  viper.GetString(FlagDetails),
	}

	tx := stake.NewTxDeclareCandidacy(pk, maxAmount, cut, description)
	return txcmd.DoTx(tx)
}

func cmdUpdateCandidacy(cmd *cobra.Command, args []string) error {
	newAddress := common.HexToAddress(viper.GetString(FlagNewAddress))
	maxAmount := viper.GetString(FlagMaxAmount)
	if maxAmount != "" {
		v := new(big.Int)
		_, ok := v.SetString(maxAmount, 10)
		if !ok || v.Cmp(big.NewInt(0)) <= 0 {
			return fmt.Errorf("max-amount must be positive interger")
		}
	}

	description := stake.Description{
		Website:  viper.GetString(FlagWebsite),
		Location: viper.GetString(FlagLocation),
		Details:  viper.GetString(FlagDetails),
	}

	tx := stake.NewTxUpdateCandidacy(newAddress, maxAmount, description)
	return txcmd.DoTx(tx)
}

func cmdWithdrawCandidacy(cmd *cobra.Command, args []string) error {
	tx := stake.NewTxWithdrawCandidacy()
	return txcmd.DoTx(tx)
}

func cmdVerifyCandidacy(cmd *cobra.Command, args []string) error {
	candidateAddress := common.HexToAddress(viper.GetString(FlagValidatorAddress))
	if candidateAddress.String() == "" {
		return fmt.Errorf("please enter candidate address using --validator-address")
	}

	verified := viper.GetBool(FlagVerified)
	tx := stake.NewTxVerifyCandidacy(candidateAddress, verified)
	return txcmd.DoTx(tx)
}

func cmdActivateCandidacy(cmd *cobra.Command, args []string) error {
	tx := stake.NewTxActivateCandidacy()
	return txcmd.DoTx(tx)
}

func cmdDelegate(cmd *cobra.Command, args []string) error {
	amount := viper.GetString(FlagAmount)
	v := new(big.Int)
	_, ok := v.SetString(amount, 10)
	if !ok || v.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("amount must be positive interger")
	}

	validatorAddress := common.HexToAddress(viper.GetString(FlagValidatorAddress))
	if validatorAddress.String() == "" {
		return fmt.Errorf("please enter validator address using --validator-address")
	}

	tx := stake.NewTxDelegate(validatorAddress, amount)
	return txcmd.DoTx(tx)
}

func cmdWithdraw(cmd *cobra.Command, args []string) error {
	validatorAddress := common.HexToAddress(viper.GetString(FlagValidatorAddress))
	if validatorAddress.String() == "" {
		return fmt.Errorf("please enter validator address using --validator-address")
	}

	amount := viper.GetString(FlagAmount)
	v := new(big.Int)
	_, ok := v.SetString(amount, 10)
	if !ok || v.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("amount must be positive interger")
	}

	tx := stake.NewTxWithdraw(validatorAddress, amount)
	return txcmd.DoTx(tx)
}
