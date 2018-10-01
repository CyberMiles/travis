package commands

import (
	"fmt"
	"math/big"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/ethereum/go-ethereum/common"

	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/sdk"
	txcmd "github.com/CyberMiles/travis/sdk/client/commands/txs"
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
	FlagPubKey                 = "pubkey"
	FlagAmount                 = "amount"
	FlagMaxAmount              = "max-amount"
	FlagCompRate               = "comp-rate"
	FlagAddress                = "address"
	FlagCandidateAddress       = "candidate-address"
	FlagName                   = "name"
	FlagEmail                  = "email"
	FlagWebsite                = "website"
	FlagLocation               = "location"
	FlagProfile                = "profile"
	FlagVerified               = "verified"
	FlagCubeBatch              = "cube-batch"
	FlagSig                    = "sig"
	FlagDelegatorAddress       = "delegator-address"
	FlagNewCandidateAddress    = "new-candidate-address"
	FlagAccountUpdateRequestId = "account-update-request-id"
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
		Short: "Allows a validator to activate itself",
		RunE:  cmdActivateCandidacy,
	}
	CmdUpdateCandidacyAccount = &cobra.Command{
		Use:   "update-candidacy-account",
		Short: "Allows a validator to update its account",
		RunE:  cmdUpdateCandidacyAccount,
	}
	CmdAcceptCandidacyAccountUpdate = &cobra.Command{
		Use:   "accept-candidacy-account-update",
		Short: "Accept the candidate's account update request and become a candidate",
		RunE:  cmdAcceptCandidacyAccountUpdate,
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
	CmdSetCompRate = &cobra.Command{
		Use:   "set-comprate",
		Short: "Set the compensation rate for a certain delegator",
		RunE:  cmdSetCompRate,
	}
)

func init() {

	// define the flags
	fsPk := flag.NewFlagSet("", flag.ContinueOnError)
	fsPk.String(FlagPubKey, "", "PubKey of the validator-candidate")

	fsAmount := flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount.String(FlagAmount, "", "Amount of CMTs")

	fsCandidate := flag.NewFlagSet("", flag.ContinueOnError)
	fsCandidate.String(FlagMaxAmount, "", "Max amount of CMTs to be staked")
	fsCandidate.String(FlagName, "", "name")
	fsCandidate.String(FlagWebsite, "", "website")
	fsCandidate.String(FlagLocation, "", "location")
	fsCandidate.String(FlagEmail, "", "email")
	fsCandidate.String(FlagProfile, "", "profile")

	fsCompRate := flag.NewFlagSet("", flag.ContinueOnError)
	fsCompRate.String(FlagCompRate, "0", "The compensation percentage of block awards to be distributed to the validator")

	fsAddr := flag.NewFlagSet("", flag.ContinueOnError)
	fsAddr.String(FlagAddress, "", "Account address")

	fsVerified := flag.NewFlagSet("", flag.ContinueOnError)
	fsVerified.String(FlagVerified, "false", "true or false")

	fsValidatorAddress := flag.NewFlagSet("", flag.ContinueOnError)
	fsValidatorAddress.String(FlagCandidateAddress, "", "validator address")

	fsNewValidatorAddress := flag.NewFlagSet("", flag.ContinueOnError)
	fsNewValidatorAddress.String(FlagNewCandidateAddress, "", "new validator address")

	fsCubeBatch := flag.NewFlagSet("", flag.ContinueOnError)
	fsCubeBatch.String(FlagCubeBatch, "", "cube batch number")

	fsSig := flag.NewFlagSet("", flag.ContinueOnError)
	fsSig.String(FlagSig, "", "cube signature")

	fsDelegatorAddress := flag.NewFlagSet("", flag.ContinueOnError)
	fsDelegatorAddress.String(FlagDelegatorAddress, "", "delegator address")

	fsAccountUpdateRequestId := flag.NewFlagSet("", flag.ContinueOnError)
	fsAccountUpdateRequestId.Int64(FlagAccountUpdateRequestId, 0, "account update request ID")

	// add the flags
	CmdDeclareCandidacy.Flags().AddFlagSet(fsPk)
	CmdDeclareCandidacy.Flags().AddFlagSet(fsCandidate)
	CmdDeclareCandidacy.Flags().AddFlagSet(fsCompRate)

	CmdUpdateCandidacy.Flags().AddFlagSet(fsCandidate)

	CmdVerifyCandidacy.Flags().AddFlagSet(fsValidatorAddress)
	CmdVerifyCandidacy.Flags().AddFlagSet(fsVerified)

	CmdDelegate.Flags().AddFlagSet(fsValidatorAddress)
	CmdDelegate.Flags().AddFlagSet(fsAmount)
	CmdDelegate.Flags().AddFlagSet(fsCubeBatch)
	CmdDelegate.Flags().AddFlagSet(fsSig)

	CmdWithdraw.Flags().AddFlagSet(fsValidatorAddress)
	CmdWithdraw.Flags().AddFlagSet(fsAmount)

	CmdSetCompRate.Flags().AddFlagSet(fsCompRate)
	CmdSetCompRate.Flags().AddFlagSet(fsDelegatorAddress)

	CmdUpdateCandidacyAccount.Flags().AddFlagSet(fsNewValidatorAddress)
	CmdAcceptCandidacyAccountUpdate.Flags().AddFlagSet(fsAccountUpdateRequestId)
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

	c, ok := sdk.NewRatFromString(viper.GetString(FlagCompRate))
	if !ok || c.LTE(sdk.ZeroRat) || c.GTE(sdk.OneRat) {
		return fmt.Errorf("comp-rate must between 0 and 1")
	}

	description := stake.Description{
		Name:     viper.GetString(FlagName),
		Email:    viper.GetString(FlagEmail),
		Website:  viper.GetString(FlagWebsite),
		Location: viper.GetString(FlagLocation),
		Profile:  viper.GetString(FlagProfile),
	}

	tx := stake.NewTxDeclareCandidacy(pk, maxAmount, c, description)
	return txcmd.DoTx(tx)
}

func cmdUpdateCandidacy(cmd *cobra.Command, args []string) error {
	maxAmount := viper.GetString(FlagMaxAmount)
	if maxAmount != "" {
		v := new(big.Int)
		_, ok := v.SetString(maxAmount, 10)
		if !ok || v.Cmp(big.NewInt(0)) <= 0 {
			return fmt.Errorf("max-amount must be positive interger")
		}
	}

	newCandidateAddress := common.HexToAddress(viper.GetString(FlagNewCandidateAddress))
	if newCandidateAddress.String() == "" {
		return fmt.Errorf("please enter new candidate address using --new-candidate-address")
	}

	description := stake.Description{
		Name:     viper.GetString(FlagName),
		Email:    viper.GetString(FlagEmail),
		Website:  viper.GetString(FlagWebsite),
		Location: viper.GetString(FlagLocation),
		Profile:  viper.GetString(FlagProfile),
	}

	tx := stake.NewTxUpdateCandidacy(maxAmount, description)
	return txcmd.DoTx(tx)
}

func cmdWithdrawCandidacy(cmd *cobra.Command, args []string) error {
	tx := stake.NewTxWithdrawCandidacy()
	return txcmd.DoTx(tx)
}

func cmdVerifyCandidacy(cmd *cobra.Command, args []string) error {
	candidateAddress := common.HexToAddress(viper.GetString(FlagCandidateAddress))
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

	validatorAddress := common.HexToAddress(viper.GetString(FlagCandidateAddress))
	if validatorAddress.String() == "" {
		return fmt.Errorf("please enter validator address using --validator-address")
	}

	cubeBatch := viper.GetString(FlagCubeBatch)
	if cubeBatch == "" {
		return fmt.Errorf("please enter cube's batch number using --cube-batch")
	}

	sig := viper.GetString(FlagSig)
	if sig == "" {
		return fmt.Errorf("please enter signature using --sig")
	}

	tx := stake.NewTxDelegate(validatorAddress, amount, cubeBatch, sig)
	return txcmd.DoTx(tx)
}

func cmdWithdraw(cmd *cobra.Command, args []string) error {
	validatorAddress := common.HexToAddress(viper.GetString(FlagCandidateAddress))
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

func cmdSetCompRate(cmd *cobra.Command, args []string) error {
	delegatorAddress := common.HexToAddress(viper.GetString(FlagDelegatorAddress))
	if delegatorAddress.String() == "" {
		return fmt.Errorf("please enter delegator address using --delegator-address")
	}

	c, ok := sdk.NewRatFromString(viper.GetString(FlagCompRate))
	if !ok || c.LTE(sdk.ZeroRat) || c.GTE(sdk.OneRat) {
		return fmt.Errorf("comp-rate must between 0 and 1")
	}

	tx := stake.NewTxSetCompRate(delegatorAddress, c)
	return txcmd.DoTx(tx)
}

func cmdUpdateCandidacyAccount(cmd *cobra.Command, args []string) error {
	newCandidateAddress := common.HexToAddress(viper.GetString(FlagNewCandidateAddress))
	if newCandidateAddress.String() == "" {
		return fmt.Errorf("please enter new candidate address using --new-candidate-address")
	}

	tx := stake.NewTxUpdateCandidacyAccount(newCandidateAddress)
	return txcmd.DoTx(tx)
}

func cmdAcceptCandidacyAccountUpdate(cmd *cobra.Command, args []string) error {
	updateAccountRequestId := viper.GetInt64(FlagAccountUpdateRequestId)
	if updateAccountRequestId == 0 {
		return fmt.Errorf("account-update-request-id must be present")
	}

	tx := stake.NewTxAcceptCandidacyAccountUpdate(updateAccountRequestId)
	return txcmd.DoTx(tx)
}
