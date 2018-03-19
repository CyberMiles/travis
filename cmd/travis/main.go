package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	basecmd "github.com/CyberMiles/travis/server/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/auto"
)

// TravisCmd is the entry point for this binary
var (
	TravisCmd = &cobra.Command{
		Use:   "travis",
		Short: "The Travis Network",
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

	TravisCmd.AddCommand(
		nodeCmd,
		clientCmd,

		lineBreak,
		auto.AutoCompleteCmd,
	)

	// prepare and add flags
	basecmd.SetUpRoot(TravisCmd)
	executor := cli.PrepareMainCmd(TravisCmd, "TR", os.ExpandEnv("$HOME/.travis-cli"))
	executor.Execute()
}
