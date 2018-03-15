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
	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/go-wire"
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

	FlagDelegatorAddress = "address"
)

func init() {
	//Add Flags
	fsPk := flag.NewFlagSet("", flag.ContinueOnError)
	fsPk.String(FlagPubKey, "", "PubKey of the validator-candidate")
	fsAddr := flag.NewFlagSet("", flag.ContinueOnError)
	fsAddr.String(FlagDelegatorAddress, "", "Delegator Hex Address")
	fsSlot := flag.NewFlagSet("", flag.ContinueOnError)
	fsSlot.String(FlagSlotId, "", "Slot ID")

	CmdQueryValidator.Flags().AddFlagSet(fsPk)
	CmdQueryDelegator.Flags().AddFlagSet(fsAddr)
	CmdQuerySlot.Flags().AddFlagSet(fsSlot)
}

func cmdQueryValidators(cmd *cobra.Command, args []string) error {

	var pks []crypto.PubKey

	prove := !viper.GetBool(commands.FlagTrustNode)
	key := stack.PrefixedKey(stake.Name(), stake.CandidatesPubKeysKey)
	height, err := query.GetParsed(key, &pks, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(pks, height)
}

func cmdQueryValidator(cmd *cobra.Command, args []string) error {

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

func cmdQueryDelegator(cmd *cobra.Command, args []string) error {

	//var bond stake.DelegatorBond
	//
	//pk, err := GetPubKey(viper.GetString(FlagPubKey))
	//if err != nil {
	//	return err
	//}
	//
	//delegatorAddr := viper.GetString(FlagDelegatorAddress)
	//delegator, err := commands.ParseActor(delegatorAddr)
	//if err != nil {
	//	return err
	//}
	//delegator = coin.ChainAddr(delegator)
	//
	//prove := !viper.GetBool(commands.FlagTrustNode)
	//key := stack.PrefixedKey(stake.Name(), stake.GetDelegatorBondKey(delegator, pk))
	//height, err := query.GetParsed(key, &bond, query.GetHeight(), prove)
	//if err != nil {
	//	return err
	//}
	//
	//return query.OutputProof(bond, height)
	return nil
}

func cmdQuerySlot(cmd *cobra.Command, args []string) error {
	var slot stake.Slot
	slotId := viper.GetString(FlagSlotId)
	if slotId == "" {
		return fmt.Errorf("please enter slot ID using --slot-id")
	}

	err := GetParsed("slot", []byte(slotId), &slot)
	if err != nil {
		return err
	}

	return Foutput(slot)
}

func cmdQuerySlots(cmd *cobra.Command, args []string) error {

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

func Get(path string, params []byte) (data.Bytes, error) {
	node := commands.GetNode()
	resp, err := node.ABCIQuery(path, params)
	if resp == nil {
		return nil, err
	}
	return data.Bytes(resp.Response.Value), err
}

func GetParsed(path string, params []byte, data interface{}) error {
	bs, err := Get(path, params)
	if err != nil {
		return err
	}

	err = wire.ReadBinaryBytes(bs, data)
	if err != nil {
		return err
	}
	return nil
}

func Foutput(v interface{}) error {
	blob, err := data.ToJSON(v)
	if err != nil {
		return err
	}
	fmt.Sprintf( "%s\n", blob)
	return nil
}