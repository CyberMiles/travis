package main

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	txcmd "github.com/CyberMiles/travis/modules/txs"
	stakecmd "github.com/CyberMiles/travis/modules/stake/commands"
	authcmd "github.com/CyberMiles/travis/modules/auth/commands"
	basecmd "github.com/cosmos/cosmos-sdk/modules/base/commands"
	rolecmd "github.com/cosmos/cosmos-sdk/modules/roles/commands"
)

// clientCmd is the entry point for this binary
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Travis light client",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func prepareClientCommands() {
	commands.AddBasicFlags(clientCmd)

	query.RootCmd.AddCommand(
		stakecmd.CmdQueryCandidates,
		stakecmd.CmdQueryCandidate,
		stakecmd.CmdQueryDelegatorBond,
		stakecmd.CmdQueryDelegatorCandidates,

		stakecmd.CmdQueryValidator,
		stakecmd.CmdQueryDelegator,
	)

	// set up the middleware
	txcmd.Middleware = txcmd.Wrappers{
		rolecmd.RoleWrapper{},
		basecmd.ChainWrapper{},
		authcmd.SigWrapper{},
	}
	txcmd.Middleware.Register(txcmd.RootCmd.PersistentFlags())

	txcmd.RootCmd.AddCommand(
		stakecmd.CmdDeclareCandidacy,
		stakecmd.CmdEditCandidacy,
		stakecmd.CmdDelegate,
		stakecmd.CmdUnbond,

		stakecmd.CmdDeclare,
		stakecmd.CmdProposeSlot,
		stakecmd.CmdAcceptSlot,
		stakecmd.CmdCancelSlot,
	)

	clientCmd.AddCommand(
		txcmd.RootCmd,
		query.RootCmd,
		lineBreak,

		commands.InitCmd,
		commands.ResetCmd,
	)
}
