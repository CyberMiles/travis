package commands

import (
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"

	"github.com/ethereum/go-ethereum/node"
	tmcfg "github.com/tendermint/tendermint/config"
	cmn "github.com/tendermint/tmlibs/common"
)

const (
	configFile        = "config.toml"
	defaultEthChainId = 111
)

type TravisConfig struct {
	BaseConfig BaseConfig      `mapstructure:",squash"`
	TMConfig   tmcfg.Config    `mapstructure:",squash"`
	EMConfig   EthermintConfig `mapstructure:"vm"`
}

func DefaultConfig() *TravisConfig {
	return &TravisConfig{
		BaseConfig: DefaultBaseConfig(),
		TMConfig:   *tmcfg.DefaultConfig(),
		EMConfig:   DefaultEthermintConfig(),
	}
}

type BaseConfig struct {
	// The root directory for all data.
	// This should be set in viper so it can unmarshal into this struct
	RootDir string `mapstructure:"home"`
}

func DefaultBaseConfig() BaseConfig {
	return BaseConfig{}
}

type EthermintConfig struct {
	EthChainId        uint   `mapstructure:"eth_chain_id"`
	RootDir           string `mapstructure:"home"`
	ABCIAddr          string `mapstructure:"abci_laddr"`
	ABCIProtocol      string `mapstructure:"abci_protocol"`
	RPCEnabledFlag    bool   `mapstructure:"rpc"`
	RPCListenAddrFlag string `mapstructure:"rpcaddr"`
	RPCPortFlag       uint   `mapstructure:"rpcport"`
	RPCCORSDomainFlag string `mapstructure:"rpccorsdomain"`
	RPCApiFlag        string `mapstructure:"rpcapi"`
	WSEnabledFlag     bool   `mapstructure:"ws"`
	WSListenAddrFlag  string `mapstructure:"wsaddr"`
	WSPortFlag        uint   `mapstructure:"wsport"`
	WSApiFlag         string `mapstructure:"wsapi"`
	VerbosityFlag     uint   `mapstructure:"verbosity"`
}

func DefaultEthermintConfig() EthermintConfig {
	return EthermintConfig{
		EthChainId:        defaultEthChainId,
		ABCIAddr:          "tcp://0.0.0.0:8848",
		ABCIProtocol:      "socket",
		RPCEnabledFlag:    true,
		RPCListenAddrFlag: node.DefaultHTTPHost,
		RPCPortFlag:       node.DefaultHTTPPort,
		RPCApiFlag:        "eth,net,web3,personal,admin",
		WSEnabledFlag:     true,
		WSListenAddrFlag:  node.DefaultWSHost,
		WSPortFlag:        node.DefaultWSPort,
		WSApiFlag:         "",
		VerbosityFlag:     3,
	}
}

// ParseConfig retrieves the default environment configuration,
// sets up the Tendermint root and ensures that the root exists
func ParseConfig() (*TravisConfig, error) {
	conf := DefaultConfig()
	err := viper.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}
	conf.TMConfig.SetRoot(conf.BaseConfig.RootDir)
	ensureRoot(conf.BaseConfig.RootDir)

	return conf, err
}

func ensureRoot(rootDir string) {
	if err := cmn.EnsureDir(rootDir, 0700); err != nil {
		cmn.PanicSanity(err.Error())
	}
	if err := cmn.EnsureDir(rootDir+"/data", 0700); err != nil {
		cmn.PanicSanity(err.Error())
	}

	configFilePath := path.Join(rootDir, configFile)

	// Write default config file if missing.
	if !cmn.FileExists(configFilePath) {
		cmn.MustWriteFile(configFilePath, []byte(defaultConfig(defaultMoniker)), 0644)
	}
}

func defaultConfig(moniker string) string {
	return strings.Replace(defaultConfigTmpl, "__MONIKER__", moniker, -1)
}

var defaultConfigTmpl = `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

moniker = "__MONIKER__"
fast_sync = true
db_backend = "leveldb"
log_level = "state:info,*:error"

[rpc]
laddr = "tcp://0.0.0.0:46657"

[p2p]
laddr = "tcp://0.0.0.0:46656"
seeds = ""

[vm]
rpc = true
rpcapi = "cmt,eth,net,web3,personal,admin"
rpcaddr = "0.0.0.0"
rpcport = 8545
ws = false
verbosity = 3


[consensus]
timeout_commit = 10000
`

var defaultMoniker = getDefaultMoniker()

// getDefaultMoniker returns a default moniker, which is the host name. If runtime
// fails to get the host name, "anonymous" will be returned.
func getDefaultMoniker() string {
	moniker, err := os.Hostname()
	if err != nil {
		moniker = "anonymous"
	}
	return moniker
}
