package commands

import (
	"fmt"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
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
)

func init() {
	//Add Flags
	fsAddr := flag.NewFlagSet("", flag.ContinueOnError)
	fsAddr.String(FlagAddress, "", "account address")

	CmdQueryValidator.Flags().AddFlagSet(fsAddr)
	CmdQueryDelegator.Flags().AddFlagSet(fsAddr)
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
	address := viper.GetString(FlagAddress)
	if address == "" {
		return fmt.Errorf("please enter validator address using --address")
	}

	err := GetParsed("/validator", []byte(address), &candidate)
	if err != nil {
		return err
	}
	return Foutput(candidate)
}

func cmdQueryDelegator(cmd *cobra.Command, args []string) error {
	var delegation []*stake.Delegation
	address := viper.GetString(FlagAddress)
	err := GetParsed("/delegator", []byte(address), &delegation)
	if err != nil {
		return err
	}
	return Foutput(delegation)
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

	//err = json.Unmarshal(bs, data)
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
