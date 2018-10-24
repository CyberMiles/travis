package api

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	tmn "github.com/tendermint/tendermint/node"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/CyberMiles/travis/vm/ethereum"
	emtTypes "github.com/CyberMiles/travis/vm/types"
	"github.com/ethereum/go-ethereum/params"
)

//----------------------------------------------------------------------
// Backend manages the underlying ethereum state for storage and processing,
// and maintains the connection to Tendermint for forwarding txs

// Backend handles the chain database and VM
// #stable - 0.4.0
type Backend struct {
	// backing ethereum structures
	ethereum  *eth.Ethereum
	ethConfig *eth.Config

	// txBroadcastLoop subscription
	//txSub *event.TypeMuxSubscription
	txsCh  chan core.NewTxsEvent
	txsSub event.Subscription

	// EthState
	es *ethereum.EthState

	// local client for in-proc app to execute the rpc functions without the overhead of http
	localClient *rpcClient.Local

	// travis chain id
	chainID string

	// moved from txpool.pendingState
	managedState *state.ManagedState
}

// NewBackend creates a new Backend
// #stable - 0.4.0
func NewBackend(ctx *node.ServiceContext, ethConfig *eth.Config) (*Backend, error) {

	// Create working ethereum state.
	es := ethereum.NewEthState()

	// eth.New takes a ServiceContext for the EventMux, the AccountManager,
	// and some basic functions around the DataDir.
	ethereum, err := eth.New(ctx, ethConfig, es)
	if err != nil {
		return nil, err
	}

	es.SetEthereum(ethereum)
	es.SetEthConfig(ethConfig)

	// send special event to go-ethereum to switch homestead=true.
	currentBlock := ethereum.BlockChain().CurrentBlock()
	ethereum.EventMux().Post(core.ChainHeadEvent{currentBlock}) // nolint: vet, errcheck

	// We don't need PoW/Uncle validation.
	ethereum.BlockChain().SetValidator(NullBlockProcessor{})

	ethereum.BlockChain().SetUmbrella(&EthUmbrella{})

	ethBackend := &Backend{
		ethereum:  ethereum,
		ethConfig: ethConfig,
		es:        es,
	}
	ethBackend.ResetState()
	return ethBackend, nil
}

func (b *Backend) ResetState() (*state.ManagedState, error) {
	currentState, err := b.Ethereum().BlockChain().State()
	if err != nil {
		return nil, err
	}
	b.managedState = state.ManageState(currentState)
	return b.managedState, nil
}

func (b *Backend) ManagedState() *state.ManagedState {
	return b.managedState
}

// Ethereum returns the underlying the ethereum object.
// #stable
func (b *Backend) Ethereum() *eth.Ethereum {
	return b.ethereum
}

func (b *Backend) DeliverTxState() *state.StateDB {
	return b.es.GetEthState()
}

// Config returns the eth.Config.
// #stable
func (b *Backend) Config() *eth.Config {
	return b.ethConfig
}

func (b *Backend) SetTMNode(tmNode *tmn.Node) {
	b.chainID = tmNode.GenesisDoc().ChainID
	b.localClient = rpcClient.NewLocal(tmNode)
	waitForServer(b.localClient)
}

func (b *Backend) PeerCount() int {
	var net *ctypes.ResultNetInfo
	net, _ = b.GetLocalClient().NetInfo()
	if net != nil {
		return len(net.Peers)
	}
	return 0
}

func (b *Backend) GetLocalClient() *rpcClient.Local {
	for b.localClient == nil {
		log.Info("Waiting for local client to set up...")
		time.Sleep(time.Second * 1)
	}
	return b.localClient
}

//----------------------------------------------------------------------
// Handle block processing

// DeliverTx appends a transaction to the current block
// #stable
func (b *Backend) DeliverTx(tx *ethTypes.Transaction) abciTypes.ResponseDeliverTx {
	return b.es.DeliverTx(tx)
}

// AccumulateRewards accumulates the rewards based on the given strategy
// #unstable
func (b *Backend) AccumulateRewards(config *params.ChainConfig, strategy *emtTypes.Strategy) {
	b.es.AccumulateRewards(config, strategy)
}

// Commit finalises the current block
// #unstable
func (b *Backend) Commit(receiver common.Address) (common.Hash, error) {
	return b.es.Commit(receiver)
}

func (b *Backend) EndBlock() {
	b.es.EndBlock()
}

// InitEthState initializes the EthState
// #unstable
func (b *Backend) InitEthState(receiver common.Address) error {
	return b.es.ResetWorkState(receiver)
}

// UpdateHeaderWithTimeInfo uses the tendermint header to update the ethereum header
// #unstable
func (b *Backend) UpdateHeaderWithTimeInfo(tmHeader abciTypes.Header) {
	b.es.UpdateHeaderWithTimeInfo(b.ethereum.APIBackend.ChainConfig(), uint64(tmHeader.Time),
		uint64(tmHeader.GetNumTxs()))
}

// GasLimit returns the maximum gas per block
// #unstable
func (b *Backend) GasLimit() uint64 {
	return b.es.GasLimit().Gas()
}

//----------------------------------------------------------------------
// Implements: node.Service

// APIs returns the collection of RPC services the ethereum package offers.
// #stable - 0.4.0
func (b *Backend) APIs() []rpc.API {
	nonceLock := new(AddrLocker)
	apis := b.Ethereum().APIs()
	// append cmt and stake api
	apis = append(apis, []rpc.API{
		{
			Namespace: "cmt",
			Version:   "1.0",
			Service:   NewCmtRPCService(b, nonceLock),
			Public:    true,
		},
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewEthRPCService(b, nonceLock),
			Public:    true,
		},
		{
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPrivateAccountAPI(b, nonceLock),
			Public:    true,
		},
	}...)

	retApis := []rpc.API{}
	for _, v := range apis {
		if v.Namespace == "net" {
			v.Service = NewNetRPCService(b)
		}
		if v.Namespace == "miner" || v.Namespace == "txpool" {
			continue
		}
		if _, ok := v.Service.(*eth.PublicMinerAPI); ok {
			continue
		}
		retApis = append(retApis, v)
	}
	return retApis
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
// #stable
func (b *Backend) Start(_ *p2p.Server) error {
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
// #stable
func (b *Backend) Stop() error {
	b.ethereum.Stop() // nolint: errcheck
	return nil
}

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
// #stable
func (b *Backend) Protocols() []p2p.Protocol {
	return nil
}

//----------------------------------------------------------------------
// We need a block processor that just ignores PoW and uncles and so on

// NullBlockProcessor does not validate anything
// #unstable
type NullBlockProcessor struct{}

// ValidateBody does not validate anything
// #unstable
func (NullBlockProcessor) ValidateBody(*ethTypes.Block) error { return nil }

// ValidateState does not validate anything
// #unstable
func (NullBlockProcessor) ValidateState(block, parent *ethTypes.Block, state *state.StateDB,
	receipts ethTypes.Receipts, usedGas uint64) error {
	return nil
}
