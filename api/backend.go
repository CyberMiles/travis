package api

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	abciTypes "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
	tmn "github.com/tendermint/tendermint/node"
	rpcClient "github.com/tendermint/tendermint/rpc/client"

	"github.com/CyberMiles/travis/modules/nonce"
	"github.com/CyberMiles/travis/vm/ethereum"
	emtTypes "github.com/CyberMiles/travis/vm/types"
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
	txSub *event.TypeMuxSubscription

	// EthState
	es *ethereum.EthState

	// client for forwarding txs to Tendermint over http
	client *rpcClient.HTTP
	// local client for in-proc app to execute the rpc functions without the overhead of http
	localClient *rpcClient.Local

	// travis chain id
	chainID string

	pendingState *state.ManagedState
}

// NewBackend creates a new Backend
// #stable - 0.4.0
func NewBackend(ctx *node.ServiceContext, ethConfig *eth.Config,
	client *rpcClient.HTTP) (*Backend, error) {

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

	ethBackend := &Backend{
		ethereum:  ethereum,
		ethConfig: ethConfig,
		es:        es,
		client:    client,
	}
	ethBackend.ResetState()
	return ethBackend, nil
}

func (b *Backend) ResetState() {
	currentState, err := b.Ethereum().BlockChain().State()
	if err != nil {
		return
	}
	b.pendingState = state.ManageState(currentState)
}

func (b *Backend) PendingState() *state.ManagedState {
	return b.pendingState
}

// Ethereum returns the underlying the ethereum object.
// #stable
func (b *Backend) Ethereum() *eth.Ethereum {
	return b.ethereum
}

// Config returns the eth.Config.
// #stable
func (b *Backend) Config() *eth.Config {
	return b.ethConfig
}

func (b *Backend) SetTMNode(tmNode *tmn.Node) {
	b.chainID = tmNode.GenesisDoc().ChainID
	b.localClient = rpcClient.NewLocal(tmNode)
	// uncomment this for TxPool broadcast tx to tendermint directly,
	// the TxPool must has SetTMClient method when uncomment this
	b.ethereum.TxPool().SetTMClient(b.localClient)
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
func (b *Backend) AccumulateRewards(strategy *emtTypes.Strategy) {
	b.es.AccumulateRewards(strategy)
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
func (b *Backend) UpdateHeaderWithTimeInfo(tmHeader *abciTypes.Header) {
	b.es.UpdateHeaderWithTimeInfo(b.ethereum.ApiBackend.ChainConfig(), uint64(tmHeader.Time),
		uint64(tmHeader.GetNumTxs()))
}

// GasLimit returns the maximum gas per block
// #unstable
func (b *Backend) GasLimit() big.Int {
	return b.es.GasLimit()
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
	}...)

	retApis := []rpc.API{}
	for _, v := range apis {
		if v.Namespace == "net" {
			v.Service = NewNetRPCService(b.ethConfig.NetworkId)
		}
		if v.Namespace == "miner" {
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
	go b.txBroadcastLoop()
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
// #stable
func (b *Backend) Stop() error {
	b.txSub.Unsubscribe()
	b.ethereum.Stop() // nolint: errcheck
	return nil
}

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
// #stable
func (b *Backend) Protocols() []p2p.Protocol {
	return nil
}

func (b *Backend) GetSequence(signers []common.Address, sequence *uint64) error {
	// key := stack.PrefixedKey(nonce.NameNonce, nonce.GetSeqKey(signers))
	key := nonce.GetSeqKey(signers)
	result, err := b.localClient.ABCIQuery("/key", key)
	if err != nil {
		return err
	}

	if len(result.Response.Value) == 0 {
		return nil
	}
	return wire.ReadBinaryBytes(result.Response.Value, sequence)
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
	receipts ethTypes.Receipts, usedGas *big.Int) error {
	return nil
}
