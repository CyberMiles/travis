package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	basecmd "github.com/cosmos/cosmos-sdk/server/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/auto"
)

// GaiaCmd is the entry point for this binary
var (
	GaiaCmd = &cobra.Command{
		Use:   "travis",
		Short: "The Travis Network delegation-game test",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	lineBreak = &cobra.Command{Run: func(*cobra.Command, []string) {}}
)

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// add commands
	prepareNodeCommands()
	prepareClientCommands()

	GaiaCmd.AddCommand(
		nodeCmd,
		clientCmd,

		lineBreak,
		auto.AutoCompleteCmd,
	)

	// prepare and add flags
	basecmd.SetUpRoot(GaiaCmd)
	executor := cli.PrepareMainCmd(GaiaCmd, "TR", os.ExpandEnv("$HOME/.travis"))
	executor.Execute()
}
