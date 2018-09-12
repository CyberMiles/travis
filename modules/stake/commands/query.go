package commands

import (
	"fmt"
	"github.com/CyberMiles/travis/sdk/client/commands"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"os"
)

/**
The stake/query/validator is to query the current stake status of the validator. Not signed.

* Validator address

The stake/query/delegator is to query the current stake status of a delegator. Not signed.

* Delegator address
*/

//nolint
const (
	FlagHeight = "height"
)

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

	CmdQueryAwardInfo = &cobra.Command{
		Use:   "awardInfo",
		RunE:  cmdQueryAwardInfo,
		Short: "Query the award info of a block",
	}
)

func init() {
	//Add Flags
	fsAddr := flag.NewFlagSet("", flag.ContinueOnError)
	fsAddr.String(FlagAddress, "", "account address")

	CmdQueryValidator.Flags().AddFlagSet(fsAddr)
	CmdQueryDelegator.Flags().AddFlagSet(fsAddr)

	//fsHeight := flag.NewFlagSet("", flag.ContinueOnError)
	//fsHeight.String(FlagHeight, "", "block height")

	//CmdQueryAwardInfo.Flags().AddFlagSet(fsHeight)
}

func cmdQueryValidators(cmd *cobra.Command, args []string) error {
	b, err := Get("/validators", []byte{0})
	if err != nil {
		return err
	}
	return Foutput(b)
}

func cmdQueryValidator(cmd *cobra.Command, args []string) error {
	address := viper.GetString(FlagAddress)
	if address == "" {
		return fmt.Errorf("please enter validator address using --address")
	}

	b, err := Get("/validator", []byte(address))
	if err != nil {
		return err
	}
	return Foutput(b)
}

func cmdQueryDelegator(cmd *cobra.Command, args []string) error {
	address := viper.GetString(FlagAddress)
	b, err := Get("/delegator", []byte(address))
	if err != nil {
		return err
	}
	return Foutput(b)
}

func cmdQueryAwardInfo(cmd *cobra.Command, args []string) error {
	height := viper.GetString(FlagHeight)
	b, err := GetByHeight("/awardInfo", []byte(height), int64(viper.GetInt(FlagHeight)))
	if err != nil {
		return err
	}
	return Foutput(b)
}

func Get(path string, params []byte) ([]byte, error) {
	node := commands.GetNode()
	resp, err := node.ABCIQuery(path, params)
	if resp == nil {
		return nil, err
	}
	return resp.Response.Value, err
}

func GetByHeight(path string, params []byte, height int64) ([]byte, error) {
	node := commands.GetNode()
	resp, err := node.ABCIQueryWithOptions(path, params, rpcclient.ABCIQueryOptions{Trusted: true, Height: int64(height)})
	if resp == nil {
		return nil, err
	}
	return resp.Response.Value, err
}

func Foutput(b []byte) error {
	_, err := fmt.Fprintf(os.Stdout, "%s\n", b)
	return err
}
