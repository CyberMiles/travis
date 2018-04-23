package txs

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk"
)

// nolint
const (
	FlagName    = "name"
	FlagNoSign  = "no-sign"
	FlagIn      = "in"
	FlagPrepare = "prepare"
	FlagAddress = "address"
	FlagType	= "type"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "tx",
	Short: "Post tx from json input",
	RunE:  doRawTx,
}

func init() {
	RootCmd.PersistentFlags().String(FlagName, "", "name to sign the tx")
	RootCmd.PersistentFlags().String(FlagAddress, "", "account address to sign the tx")
	RootCmd.PersistentFlags().Bool(FlagNoSign, false, "don't add a signature")
	RootCmd.PersistentFlags().String(FlagPrepare, "", "file to store prepared tx")
	RootCmd.Flags().String(FlagIn, "", "file with tx in json format")
	RootCmd.PersistentFlags().String(FlagType, "commit", "type(sync|commit) of broadcast tx to tendermint")
}

func doRawTx(cmd *cobra.Command, args []string) error {
	raw, err := readInput(viper.GetString(FlagIn))
	if err != nil {
		return err
	}

	// parse the input
	var tx sdk.Tx
	err = json.Unmarshal(raw, &tx)
	if err != nil {
		return errors.WithStack(err)
	}

	// sign it
	err = SignTx(tx)
	if err != nil {
		return err
	}

	commit := viper.GetString(FlagType)
	if commit == "commit" {
		// otherwise, post it and display response
		bres, err := PrepareOrPostTx(tx)
		if err != nil {
			return err
		}
		if bres == nil {
			return nil // successful prep, nothing left to do
		}
		return OutputTx(bres) // print response of the post
	} else {
		bres, err := PrepareOrPostTxSync(tx)
		if err != nil {
			return err
		}
		if bres == nil {
			return nil // successful prep, nothing left to do
		}
		return OutputTxSync(bres) // print response of the post
	}
}
