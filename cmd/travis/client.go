package main

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	txcmd "github.com/CyberMiles/travis/client/commands/txs"
	stakecmd "github.com/CyberMiles/travis/modules/stake/commands"
	authcmd "github.com/CyberMiles/travis/modules/auth/commands"
	noncecmd "github.com/CyberMiles/travis/modules/nonce/commands"
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
		noncecmd.NonceQueryCmd,
		stakecmd.CmdQueryValidator,
		stakecmd.CmdQueryValidators,
		stakecmd.CmdQueryDelegator,
		stakecmd.CmdQuerySlot,
		stakecmd.CmdQuerySlots,
	)

	// set up the middleware
	txcmd.Middleware = txcmd.Wrappers{
		noncecmd.NonceWrapper{},
		authcmd.SigWrapper{},
	}
	txcmd.Middleware.Register(txcmd.RootCmd.PersistentFlags())

	txcmd.RootCmd.AddCommand(
		stakecmd.CmdDeclareCandidacy,
		stakecmd.CmdEditCandidacy,
		stakecmd.CmdWithdraw,
		stakecmd.CmdProposeSlot,
		stakecmd.CmdAcceptSlot,
		stakecmd.CmdWithdrawSlot,
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
