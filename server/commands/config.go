package commands

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/spf13/viper"

	"github.com/ethereum/go-ethereum/node"
	tmcfg "github.com/tendermint/tendermint/config"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/CyberMiles/travis/utils"
)

const (
	defaultConfigDir = "config"
	defaultDataDir   = "data"

	configFile = "config.toml"
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
	ChainId             uint   `mapstructure:"chainid"`
	RootDir             string `mapstructure:"home"`
	ABCIAddr            string `mapstructure:"abci_laddr"`
	ABCIProtocol        string `mapstructure:"abci_protocol"`
	RPCEnabledFlag      bool   `mapstructure:"rpc"`
	RPCListenAddrFlag   string `mapstructure:"rpcaddr"`
	RPCPortFlag         uint   `mapstructure:"rpcport"`
	RPCCORSDomainFlag   string `mapstructure:"rpccorsdomain"`
	RPCApiFlag          string `mapstructure:"rpcapi"`
	RPCVirtualHostsFlag string `mapstructure:"rpcvhosts"`
	WSEnabledFlag       bool   `mapstructure:"ws"`
	WSListenAddrFlag    string `mapstructure:"wsaddr"`
	WSPortFlag          uint   `mapstructure:"wsport"`
	WSApiFlag           string `mapstructure:"wsapi"`
	VerbosityFlag       uint   `mapstructure:"verbosity"`
	GCMode              string `mapstructure:"gcmode"`
}

func DefaultEthermintConfig() EthermintConfig {
	return EthermintConfig{
		ChainId:             utils.TestNet,
		ABCIAddr:            "tcp://0.0.0.0:8848",
		ABCIProtocol:        "socket",
		RPCEnabledFlag:      true,
		RPCListenAddrFlag:   "0.0.0.0",
		RPCPortFlag:         node.DefaultHTTPPort,
		RPCCORSDomainFlag:   "*",
		RPCApiFlag:          "cmt,eth,net,web3,personal",
		RPCVirtualHostsFlag: "localhost",
		WSEnabledFlag:       false,
		WSListenAddrFlag:    node.DefaultWSHost,
		WSPortFlag:          node.DefaultWSPort,
		WSApiFlag:           "",
		VerbosityFlag:       3,
		GCMode:              "full",
	}
}

// copied from tendermint/commands/root.go
// to call our revised EnsureRoot
func ParseConfig() (*TravisConfig, error) {
	conf := DefaultConfig()
	// set chainid as per --env
	switch viper.GetString(FlagENV) {
	case "staging":
		conf.EMConfig.ChainId = utils.Staging
	case "mainnet":
		conf.EMConfig.ChainId = utils.MainNet
	case "testnet":
		conf.EMConfig.ChainId = utils.TestNet
	default:
		conf.EMConfig.ChainId = utils.PrivateChain
	}

	err := viper.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}
	conf.TMConfig.SetRoot(conf.TMConfig.RootDir)
	// replace EnsureRoot of tendermint with our own
	ensureRoot(conf)

	return conf, nil
}

// copied from tendermint/config/toml.go
// modified to override some defaults and append vm configs
func ensureRoot(conf *TravisConfig) {
	rootDir := conf.TMConfig.RootDir

	if err := cmn.EnsureDir(rootDir, 0700); err != nil {
		cmn.PanicSanity(err.Error())
	}
	if err := cmn.EnsureDir(filepath.Join(rootDir, defaultConfigDir), 0700); err != nil {
		cmn.PanicSanity(err.Error())
	}
	if err := cmn.EnsureDir(filepath.Join(rootDir, defaultDataDir), 0700); err != nil {
		cmn.PanicSanity(err.Error())
	}

	configFilePath := path.Join(rootDir, defaultConfigDir, configFile)

	// Write default config file if missing.
	if !cmn.FileExists(configFilePath) {
		// override some defaults
		conf.TMConfig.Consensus.TimeoutCommit = 10000
		//conf.TMConfig.LogLevel = "app:debug,*:error"
		//conf.TMConfig.Consensus.MaxBlockSizeTxs = 50000
		// write config file
		tmcfg.WriteConfigFile(configFilePath, &conf.TMConfig)
		// append vm configs
		AppendVMConfig(configFilePath, conf)
	}
}

func AppendVMConfig(configFilePath string, conf *TravisConfig) {
	var configTemplate *template.Template
	var err error
	if configTemplate, err = template.New("vmConfigTemplate").Parse(defaultVmTemplate); err != nil {
		panic(err)
	}

	var buffer bytes.Buffer
	if err := configTemplate.Execute(&buffer, conf); err != nil {
		panic(err)
	}

	f, err := os.OpenFile(configFilePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.Write(buffer.Bytes()); err != nil {
		panic(err)
	}
}

var defaultVmTemplate = `
[vm]
chainid = {{ .EMConfig.ChainId }}
rpc = {{ .EMConfig.RPCEnabledFlag }}
rpcapi = "{{ .EMConfig.RPCApiFlag }}"
rpcaddr = "{{ .EMConfig.RPCListenAddrFlag }}"
rpcport = {{ .EMConfig.RPCPortFlag }}
rpccorsdomain = "{{ .EMConfig.RPCCORSDomainFlag }}"
rpcvhosts = "{{ .EMConfig.RPCVirtualHostsFlag }}"
ws = {{ .EMConfig.WSEnabledFlag }}
verbosity = "{{ .EMConfig.VerbosityFlag }}"
`
