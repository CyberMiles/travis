/*
Package commands contains any general setup/helpers valid for all subcommands
*/
package commands

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/lite"
	"github.com/tendermint/tendermint/libs/cli"
	cmn "github.com/tendermint/tendermint/libs/common"

	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/CyberMiles/travis/sdk/client"
	"github.com/ethereum/go-ethereum/common"
)

var (
	trustedProv lite.Provider
	sourceProv  lite.Provider
)

const (
	ChainFlag = "chain-id"
	NodeFlag  = "node"
)

// AddBasicFlags adds --node and --chain-id, which we need for everything
func AddBasicFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String(ChainFlag, "", "Chain ID of tendermint node")
	cmd.PersistentFlags().String(NodeFlag, "", "<host>:<port> to tendermint rpc interface for this chain")
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
		rootDir := viper.GetString(cli.HomeFlag)
		trustedProv = client.GetLocalProvider(rootDir)
	}
	return trustedProv
}

// GetProviders creates a trusted (local) seed provider and a remote
// provider based on configuration.
func GetProviders() (trusted lite.Provider, source lite.Provider) {
	return GetTrustedProvider(), GetSourceProvider()
}

// ParseActor parses an address of form:
// [<chain>:][<app>:]<hex address>
// into a sdk.Actor.
// If app is not specified or "", then assume auth.NameSigs
func ParseActor(input string) (res common.Address, err error) {
	input = strings.TrimSpace(input)
	addr, err := hex.DecodeString(cmn.StripHex(input))
	if err != nil {
		return res, errors.Errorf("Address is invalid hex: %v\n", err)
	}
	res = common.BytesToAddress(addr)
	return
}

// ParseActors takes a comma-separated list of actors and parses them into
// a slice
func ParseActors(key string) (signers []common.Address, err error) {
	var act common.Address
	for _, k := range strings.Split(key, ",") {
		act, err = ParseActor(k)
		if err != nil {
			return
		}
		signers = append(signers, act)
	}
	return
}

// GetOneArg makes sure there is exactly one positional argument
func GetOneArg(args []string, argname string) (string, error) {
	if len(args) == 0 {
		return "", errors.Errorf("Missing required argument [%s]", argname)
	}
	if len(args) > 1 {
		return "", errors.Errorf("Only accepts one argument [%s]", argname)
	}
	return args[0], nil
}

// ParseHexFlag takes a flag name and parses the viper contents as hex
func ParseHexFlag(flag string) ([]byte, error) {
	arg := viper.GetString(flag)
	if arg == "" {
		return nil, errors.Errorf("No such flag: %s", flag)
	}
	value, err := hex.DecodeString(cmn.StripHex(arg))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Cannot parse %s", flag))
	}
	return value, nil

}
