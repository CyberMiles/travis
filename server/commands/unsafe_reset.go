package commands

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/privval"
)

// RemoveAddrBookCmd removes the address book file.
var RemoveAddrBookCmd = &cobra.Command{
	Use:   "unsafe_remove_addrbook",
	Short: "(unsafe) Remove address book file",
	Run:   removeAddrBook,
}

// ResetPrivValidatorCmd resets the private validator files.
var ResetPrivValidatorCmd = &cobra.Command{
	Use:   "unsafe_reset_priv_validator",
	Short: "(unsafe) Reset this node's validator to genesis state",
	Run:   resetPrivValidator,
}

// XXX: this is totally unsafe.
// it's only suitable for testnets.
func removeAddrBook(cmd *cobra.Command, args []string) {
	addrBookFile := config.TMConfig.P2P.AddrBookFile()
	if err := os.Remove(addrBookFile); err == nil {
		logger.Info("Removed existing address book", "file", addrBookFile)
	} else if !os.IsNotExist(err) {
		logger.Info("Error removing address book", "file", addrBookFile, "err", err)
	}
}

// XXX: this is totally unsafe.
// it's only suitable for testnets.
func resetPrivValidator(cmd *cobra.Command, args []string) {
	resetFilePV(config.TMConfig.PrivValidatorFile(), logger)
}

func resetFilePV(privValFile string, logger log.Logger) {
	if _, err := os.Stat(privValFile); err == nil {
		pv := privval.LoadFilePV(privValFile)
		pv.Reset()
		logger.Info("Reset private validator file to genesis state", "file", privValFile)
	} else {
		pv := privval.GenFilePV(privValFile)
		pv.Save()
		logger.Info("Generated private validator file", "file", privValFile)
	}
}
