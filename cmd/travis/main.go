package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/CyberMiles/travis/sdk/client/commands/auto"
	basecmd "github.com/CyberMiles/travis/server/commands"
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
		attachCmd,
		versionCmd,

		lineBreak,
		auto.AutoCompleteCmd,
	)

	// prepare and add flags
	basecmd.SetUpRoot(TravisCmd)
	executor := cli.PrepareMainCmd(TravisCmd, "TR", os.ExpandEnv("$HOME/.travis-cli"))
	executor.Execute()
}
