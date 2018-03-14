package commands

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/CyberMiles/travis/modules/coin"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/CyberMiles/travis/modules/stake"
	"fmt"
)

/**
The stake/query/validator is to query the current stake status of the validator. Not signed.

* Validator address

The stake/query/delegator is to query the current stake status of a delegator. Not signed.

* Delegator address
 */

//nolint
var (
	//CmdQueryCandidates = &cobra.Command{
	//	Use:   "candidates",
	//	Short: "Query for the set of validator-candidates pubkeys",
	//	RunE:  cmdQueryCandidates,
	//}
	//
	//CmdQueryCandidate = &cobra.Command{
	//	Use:   "candidate",
	//	Short: "Query a validator-candidate account",
	//	RunE:  cmdQueryCandidate,
	//}
	//
	//CmdQueryDelegatorBond = &cobra.Command{
	//	Use:   "delegator-bond",
	//	Short: "Query a delegators bond based on address and candidate pubkey",
	//	RunE:  cmdQueryDelegatorBond,
	//}
	//
	//CmdQueryDelegatorCandidates = &cobra.Command{
	//	Use:   "delegator-candidates",
	//	RunE:  cmdQueryDelegatorCandidates,
	//	Short: "Query all delegators candidates' pubkeys based on address",
	//}

	CmdQueryValidator = &cobra.Command{
		Use:   "validator",
		RunE:  cmdQueryValidator,
		Short: "Query the current stake status of a validator",
	}

	CmdQueryValidators = &cobra.Command{
		Use:   "validators",
		RunE:  cmdQueryValidators,
		Short: "Query a list of all current validators and validator candidates",
	}

	CmdQueryDelegator = &cobra.Command{
		Use:   "delegator",
		RunE:  cmdQueryDelegator,
		Short: "Query the current stake status of a delegator",
	}

	CmdQuerySlot = &cobra.Command{
		Use:   "slot",
		RunE:  cmdQuerySlot,
		Short: "Query the current status of a slot",
	}

	CmdQuerySlots = &cobra.Command{
		Use:   "slots",
		RunE:  cmdQuerySlots,
		Short: "Query all open and close slots for staking",
	}

	FlagDelegatorAddress = "delegator-address"
)

func init() {
	//Add Flags
	fsPk := flag.NewFlagSet("", flag.ContinueOnError)
	fsPk.String(FlagPubKey, "", "PubKey of the validator-candidate")
	fsAddr := flag.NewFlagSet("", flag.ContinueOnError)
	fsAddr.String(FlagDelegatorAddress, "", "Delegator Hex Address")

	CmdQueryCandidate.Flags().AddFlagSet(fsPk)
	CmdQueryDelegatorBond.Flags().AddFlagSet(fsPk)
	CmdQueryDelegatorBond.Flags().AddFlagSet(fsAddr)
	CmdQueryDelegatorCandidates.Flags().AddFlagSet(fsAddr)
}

func cmdQueryCandidates(cmd *cobra.Command, args []string) error {

	var pks []crypto.PubKey

	prove := !viper.GetBool(commands.FlagTrustNode)
	h := viper.GetInt64("height")
	fmt.Printf("height", h)
	key := stack.PrefixedKey(stake.Name(), stake.CandidatesPubKeysKey)
	fmt.Printf("cmdQueryCandidats, key: %v", key)
	//height, err := query.GetParsed(key, &pks, query.GetHeight(), prove)
	h = query.GetHeight()
	height, err := query.GetParsed(key, &pks, h, prove)
	if err != nil {
		return err
	}

	return query.OutputProof(pks, height)
}

func cmdQueryCandidate(cmd *cobra.Command, args []string) error {

	var candidate stake.Candidate

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	prove := !viper.GetBool(commands.FlagTrustNode)
	key := stack.PrefixedKey(stake.Name(), stake.GetCandidateKey(pk))
	height, err := query.GetParsed(key, &candidate, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(candidate, height)
}

func cmdQueryDelegatorBond(cmd *cobra.Command, args []string) error {

	var bond stake.DelegatorBond

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	delegatorAddr := viper.GetString(FlagDelegatorAddress)
	delegator, err := commands.ParseActor(delegatorAddr)
	if err != nil {
		return err
	}
	delegator = coin.ChainAddr(delegator)

	prove := !viper.GetBool(commands.FlagTrustNode)
	key := stack.PrefixedKey(stake.Name(), stake.GetDelegatorBondKey(delegator, pk))
	height, err := query.GetParsed(key, &bond, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(bond, height)
}

func cmdQueryDelegatorCandidates(cmd *cobra.Command, args []string) error {

	delegatorAddr := viper.GetString(FlagDelegatorAddress)
	delegator, err := commands.ParseActor(delegatorAddr)
	if err != nil {
		return err
	}
	delegator = coin.ChainAddr(delegator)

	prove := !viper.GetBool(commands.FlagTrustNode)
	key := stack.PrefixedKey(stake.Name(), stake.GetDelegatorBondsKey(delegator))
	var candidates []crypto.PubKey
	height, err := query.GetParsed(key, &candidates, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(candidates, height)
}

func cmdQueryValidator(cmd *cobra.Command, args []string) error {
	// todo

	return nil
}

func cmdQueryValidators(cmd *cobra.Command, args []string) error {
	// todo

	return nil
}

func cmdQueryDelegator(cmd *cobra.Command, args []string) error {
	// todo

	return nil
}

func cmdQuerySlot(cmd *cobra.Command, args []string) error {
	// todo

	return nil
}

func cmdQuerySlots(cmd *cobra.Command, args []string) error {
	// todo

	return nil
}