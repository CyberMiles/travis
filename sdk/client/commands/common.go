/*
Package commands contains any general setup/helpers valid for all subcommands
*/
package commands

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/lite"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/CyberMiles/travis/sdk/client"
)

var (
	trustedProv lite.Provider
	sourceProv  lite.Provider
)

const (
	ChainFlag   = "chain-id"
	NodeFlag    = "node"
	CliHomeFlag = "cli-home"
)

// AddBasicFlags adds --node and --chain-id, which we need for everything
func AddBasicFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String(ChainFlag, "", "Chain ID of tendermint node")
	cmd.PersistentFlags().String(NodeFlag, "", "<host>:<port> to tendermint rpc interface for this chain")
	cmd.PersistentFlags().String(CliHomeFlag, os.ExpandEnv("$HOME/.travis-cli"), "directory for cli")
}

// GetChainID reads ChainID from the flags
func GetChainID() string {
	return viper.GetString(ChainFlag)
}

// GetNode prepares a simple rpc.Client from the flags
func GetNode() rpcclient.Client {
	return client.GetNode(viper.GetString(NodeFlag))
}

// GetSourceProvider returns a provider pointing to an rpc handler
func GetSourceProvider() lite.Provider {
	if sourceProv == nil {
		node := viper.GetString(NodeFlag)
		sourceProv = client.GetRPCProvider(node)
	}
	return sourceProv
}

// GetTrustedProvider returns a reference to a local store with cache
func GetTrustedProvider() lite.Provider {
	if trustedProv == nil {
		rootDir := viper.GetString(CliHomeFlag)
		trustedProv = client.GetLocalProvider(rootDir)
	}
	return trustedProv
}

// GetProviders creates a trusted (local) seed provider and a remote
// provider based on configuration.
func GetProviders() (trusted lite.Provider, source lite.Provider) {
	return GetTrustedProvider(), GetSourceProvider()
}
