package commands

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/go-wire"
	"fmt"
	"os"
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

	var candidates stake.Candidates

	err := GetParsed("/validators", []byte{0}, &candidates)
	if err != nil {
		return err
	}

	return Foutput(candidates)
}

func cmdQueryValidator(cmd *cobra.Command, args []string) error {

	var candidate stake.Candidate

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	err = GetParsed("/validator", []byte(pk.KeyString()), &candidate)
	if err != nil {
		return err
	}

	return Foutput(candidate)
}

func cmdQueryDelegator(cmd *cobra.Command, args []string) error {
	//var candidate stake.Candidate
	//
	//pk, err := GetPubKey(viper.GetString(FlagPubKey))
	//if err != nil {
	//	return err
	//}
	//
	//err = GetParsed("/validator", pk.Address(), &candidate)
	//if err != nil {
	//	return err
	//}
	//
	//return Foutput(candidate)
	return nil
}

func cmdQuerySlot(cmd *cobra.Command, args []string) error {
	var slot stake.Slot
	slotId := viper.GetString(FlagSlotId)
	if slotId == "" {
		return fmt.Errorf("please enter slot ID using --slot-id")
	}

	err := GetParsed("/slot", []byte(slotId), &slot)
	if err != nil {
		return err
	}

	return Foutput(slot)
}

func cmdQuerySlots(cmd *cobra.Command, args []string) error {
	var slots []*stake.Slot

	err := GetParsed("/slots", []byte{0}, &slots)
	if err != nil {
		return err
	}

	return Foutput(slots)
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
	_, err = fmt.Fprintf(os.Stdout, "%s\n", blob)
	return nil
}