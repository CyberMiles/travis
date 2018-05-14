package commands

import (
	"os"
	"path"

	"github.com/spf13/viper"

	"github.com/ethereum/go-ethereum/node"
	tmcfg "github.com/tendermint/tendermint/config"
	cmn "github.com/tendermint/tmlibs/common"
)

const (
	defaultConfigDir  = "config"
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

	rootDir := conf.TMConfig.RootDir
	conf.TMConfig.SetRoot(rootDir)

	configFilePath := path.Join(rootDir, defaultConfigDir, configFile)
	if !cmn.FileExists(configFilePath) {
		tmcfg.EnsureRoot(rootDir)
		// append vm configs
		AppendVMConfig(configFilePath)
	}

	return conf, nil
}

func AppendVMConfig(configFilePath string) {
	f, err := os.OpenFile(configFilePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := f.Write([]byte(defaultVm)); err != nil {
		panic(err)
	}
}

var defaultVm = `
[vm]
rpc = true
rpcapi = "cmt,eth,net,web3,personal,admin"
rpcaddr = "0.0.0.0"
rpcport = 8545
ws = false
verbosity = 1
`
