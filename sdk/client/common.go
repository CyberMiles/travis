package client

import (
	"github.com/tendermint/tendermint/lite"
	certclient "github.com/tendermint/tendermint/lite/client"
	"github.com/tendermint/tendermint/lite/files"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// GetNode prepares a simple rpc.Client for the given endpoint
func GetNode(url string) rpcclient.Client {
	return rpcclient.NewHTTP(url, "/websocket")
}

// GetRPCProvider retuns a certifier compatible data source using
// tendermint RPC
func GetRPCProvider(url string) lite.Provider {
	return certclient.NewHTTPProvider(url)
}

// GetLocalProvider returns a reference to a file store of headers
// wrapped with an in-memory cache
func GetLocalProvider(dir string) lite.Provider {
	return lite.NewCacheProvider(
		lite.NewMemStoreProvider(),
		files.NewProvider(dir),
	)
}
