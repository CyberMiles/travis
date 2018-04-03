package main

import (
	"github.com/spf13/cobra"
	abci "github.com/tendermint/abci/types"
	sdk "github.com/cosmos/cosmos-sdk"
	basecmd "github.com/CyberMiles/travis/server/commands"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/CyberMiles/travis/modules/stake"
)

// nodeCmd is the entry point for this binary
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "The Travis Network",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func prepareNodeCommands() {
	nodeCmd.AddCommand(
		basecmd.InitCmd,
		basecmd.GetTickStartCmd(sdk.TickerFunc(tickFn)),
	)
}

// Tick - Called every block even if no transaction, process all queues,
// validator rewards, and calculate the validator set difference
func tickFn(ctx sdk.Context, store state.SimpleDB) (change []*abci.Validator, err error) {
	// first need to prefix the store, at this point it's a global store
	store = stack.PrefixedStore("stake", store)

	// execute Tick
	change, err = stake.UpdateValidatorSet(store)
	return
}