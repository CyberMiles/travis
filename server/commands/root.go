package commands

import (
	"flag"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/urfave/cli.v1"

	ethUtils "github.com/ethereum/go-ethereum/cmd/utils"
	ethlog "github.com/ethereum/go-ethereum/log"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"

	emtUtils "github.com/CyberMiles/travis/vm/cmd/utils"
)

//nolint
const (
	FlagLogLevel = "log_level"
	FlagPassword = "password"
	FlagLightKDF = "lightkdf"

	defaultLogLevel = "error"
)

var (
	config  = DefaultConfig()
	context *cli.Context
	logger  = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "main")
)

// preRunSetup should be set as PersistentPreRunE on the root command to
// properly handle the logging and the tracer
func preRunSetup(cmd *cobra.Command, args []string) (err error) {
	config, err = ParseConfig()
	if err != nil {
		return err
	}
	level := viper.GetString(FlagLogLevel)
	logger, err = tmflags.ParseLogLevel(level, logger, defaultLogLevel)
	if err != nil {
		return err
	}
	if viper.GetBool(tmcli.TraceFlag) {
		logger = log.NewTracingLogger(logger)
	}
	setupEmtContext()
	return nil
}

// SetUpRoot - initialize the root command
func SetUpRoot(cmd *cobra.Command) {
	cmd.PersistentPreRunE = preRunSetup
	cmd.PersistentFlags().String(FlagLogLevel, defaultLogLevel, "Log level")
}

// copied from ethermint
var (
	// flags that configure the go-ethereum node
	nodeFlags = []cli.Flag{
		ethUtils.DataDirFlag,
		ethUtils.KeyStoreDirFlag,
		ethUtils.NoUSBFlag,
		// Performance tuning
		ethUtils.CacheFlag,
		ethUtils.TrieCacheGenFlag,
		// Account settings
		ethUtils.UnlockedAccountFlag,
		ethUtils.PasswordFileFlag,
		ethUtils.VMEnableDebugFlag,
		ethUtils.LightKDFFlag,
		// Logging and debug settings
		ethUtils.NoCompactionFlag,
		// Gas price oracle settings
		ethUtils.GpoBlocksFlag,
		ethUtils.GpoPercentileFlag,
		emtUtils.TargetGasLimitFlag,
		// Gas Price
		ethUtils.GasPriceFlag,
		// Network Id
		ethUtils.NetworkIdFlag,
		ethUtils.ListenPortFlag,
	}

	rpcFlags = []cli.Flag{
		ethUtils.RPCEnabledFlag,
		ethUtils.RPCListenAddrFlag,
		ethUtils.RPCPortFlag,
		ethUtils.RPCCORSDomainFlag,
		ethUtils.RPCApiFlag,
		ethUtils.RPCVirtualHostsFlag,
		ethUtils.IPCDisabledFlag,
		ethUtils.WSEnabledFlag,
		ethUtils.WSListenAddrFlag,
		ethUtils.WSPortFlag,
		ethUtils.WSApiFlag,
		ethUtils.WSAllowedOriginsFlag,
	}

	// flags that configure the ABCI app
	ethermintFlags = []cli.Flag{
		emtUtils.TendermintAddrFlag,
		emtUtils.ABCIAddrFlag,
		emtUtils.ABCIProtocolFlag,
		emtUtils.VerbosityFlag,
		emtUtils.ConfigFileFlag,
		emtUtils.WithTendermintFlag,
		ethUtils.GCModeFlag,
	}
)

func setupEmtContext() error {
	// create a new context to invoke ethermint
	a := cli.NewApp()
	a.Name = "travis"
	a.Flags = []cli.Flag{}
	a.Flags = append(a.Flags, nodeFlags...)
	a.Flags = append(a.Flags, rpcFlags...)
	a.Flags = append(a.Flags, ethermintFlags...)

	set, err := flagSet(a.Name, a.Flags)
	if err != nil {
		return err
	}

	context = cli.NewContext(a, set, nil)

	context.GlobalSet(ethUtils.DataDirFlag.Name, config.BaseConfig.RootDir)
	context.GlobalSet(ethUtils.NetworkIdFlag.Name, strconv.Itoa(int(config.EMConfig.ChainId)))
	context.GlobalSet(emtUtils.VerbosityFlag.Name, strconv.Itoa(int(config.EMConfig.VerbosityFlag)))

	context.GlobalSet(emtUtils.TendermintAddrFlag.Name, config.TMConfig.RPC.ListenAddress)

	context.GlobalSet(emtUtils.ABCIAddrFlag.Name, config.EMConfig.ABCIAddr)
	context.GlobalSet(emtUtils.ABCIProtocolFlag.Name, config.EMConfig.ABCIProtocol)

	context.GlobalSet(ethUtils.RPCEnabledFlag.Name, strconv.FormatBool(config.EMConfig.RPCEnabledFlag))
	context.GlobalSet(ethUtils.RPCApiFlag.Name, config.EMConfig.RPCApiFlag)
	context.GlobalSet(ethUtils.RPCVirtualHostsFlag.Name, config.EMConfig.RPCVirtualHostsFlag)

	context.GlobalSet(ethUtils.RPCListenAddrFlag.Name, config.EMConfig.RPCListenAddrFlag)
	context.GlobalSet(ethUtils.RPCPortFlag.Name, strconv.Itoa(int(config.EMConfig.RPCPortFlag)))
	context.GlobalSet(ethUtils.RPCCORSDomainFlag.Name, config.EMConfig.RPCCORSDomainFlag)

	context.GlobalSet(ethUtils.WSEnabledFlag.Name, strconv.FormatBool(config.EMConfig.WSEnabledFlag))
	context.GlobalSet(ethUtils.WSApiFlag.Name, config.EMConfig.WSApiFlag)

	context.GlobalSet(ethUtils.IPCDisabledFlag.Name, strconv.FormatBool(config.EMConfig.IPCDisabledFlag))
	if config.EMConfig.ListenPortFlag > 0 {
		context.GlobalSet(ethUtils.ListenPortFlag.Name, strconv.Itoa(int(config.EMConfig.ListenPortFlag)))
	}

	context.GlobalSet(ethUtils.LightKDFFlag.Name, strconv.FormatBool(config.EMConfig.LightKDFFlag))

	if err := emtUtils.Setup(context); err != nil {
		return err
	}
	return nil
}

func flagSet(name string, flags []cli.Flag) (*flag.FlagSet, error) {
	set := flag.NewFlagSet(name, flag.ContinueOnError)

	for _, f := range flags {
		f.Apply(set)
	}
	return set, nil
}

func SetupAccountContext() (*cli.Context, error) {
	err := setupEmtContext()
	if err != nil {
		return nil, err
	}

	// extra flag from command line args
	// log level
	lvl, err := ethlog.LvlFromString(viper.GetString(FlagLogLevel))
	if err != nil {
		return nil, err
	}
	context.GlobalSet(emtUtils.VerbosityFlag.Name, strconv.Itoa(int(lvl)))
	if err := emtUtils.Setup(context); err != nil {
		return nil, err
	}
	// password file
	context.GlobalSet(ethUtils.PasswordFileFlag.Name, viper.GetString(FlagPassword))

	// lightkdf
	context.GlobalSet(ethUtils.LightKDFFlag.Name, strconv.FormatBool(viper.GetBool(FlagLightKDF)))

	return context, nil
}
