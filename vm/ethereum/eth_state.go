package ethereum

import (
	"bytes"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth"
	//"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	abciTypes "github.com/tendermint/abci/types"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/errors"
	gov "github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/utils"
	emtTypes "github.com/CyberMiles/travis/vm/types"
)

//----------------------------------------------------------------------
// EthState manages concurrent access to the intermediate workState object
// The ethereum tx pool fires TxPreEvent in a go-routine,
// and the miner subscribes to this in another go-routine and processes the tx onto
// an intermediate state. We used to use `unsafe` to overwrite the miner, but this
// didn't work because it didn't affect the already launched go-routines.
// So instead we introduce the Pending API in a small commit in go-ethereum
// so we don't even start the miner there, and instead manage the intermediate state from here.
// In the same commit we also fire the TxPreEvent synchronously so the order is preserved,
// instead of using a go-routine.

type EthState struct {
	ethereum  *eth.Ethereum
	ethConfig *eth.Config

	mtx  sync.Mutex
	work workState // latest working state
}

// After NewEthState, call SetEthereum and SetEthConfig.
func NewEthState() *EthState {
	return &EthState{
		ethereum:  nil, // set with SetEthereum
		ethConfig: nil, // set with SetEthConfig
	}
}

func (es *EthState) SetEthereum(ethereum *eth.Ethereum) {
	es.ethereum = ethereum
}

func (es *EthState) SetEthConfig(ethConfig *eth.Config) {
	es.ethConfig = ethConfig
}

// Execute the transaction.
func (es *EthState) DeliverTx(tx *ethTypes.Transaction) abciTypes.ResponseDeliverTx {
	es.mtx.Lock()
	defer es.mtx.Unlock()

	blockchain := es.ethereum.BlockChain()
	chainConfig := es.ethereum.ApiBackend.ChainConfig()
	blockHash := common.Hash{}
	return es.work.deliverTx(blockchain, es.ethConfig, chainConfig, blockHash, tx)
}

// called by travis tx only in deliver_tx
func (es *EthState) AddNonce(addr common.Address) {
	es.work.state.SetNonce(addr, es.work.state.GetNonce(addr)+1)
}

// Accumulate validator rewards.
func (es *EthState) AccumulateRewards(strategy *emtTypes.Strategy) {
	es.mtx.Lock()
	defer es.mtx.Unlock()

	es.work.accumulateRewards(strategy)
}

// Commit and reset the work.
func (es *EthState) Commit(receiver common.Address) (common.Hash, error) {
	es.mtx.Lock()
	defer es.mtx.Unlock()

	blockHash, err := es.work.commit(es.ethereum.BlockChain(), es.ethereum.ChainDb())
	if err != nil {
		return common.Hash{}, err
	}

	err = es.resetWorkState(receiver)
	if err != nil {
		return common.Hash{}, err
	}

	return blockHash, err
}

func (es *EthState) EndBlock() {
	utils.BlockGasFee.Set(es.work.totalUsedGasFee)
}

func (es *EthState) ResetWorkState(receiver common.Address) error {
	es.mtx.Lock()
	defer es.mtx.Unlock()

	return es.resetWorkState(receiver)
}

func (es *EthState) resetWorkState(receiver common.Address) error {

	blockchain := es.ethereum.BlockChain()
	state, err := blockchain.State()
	if err != nil {
		return err
	}

	currentBlock := blockchain.CurrentBlock()
	ethHeader := newBlockHeader(receiver, currentBlock)

	es.work = workState{
		header:          ethHeader,
		parent:          currentBlock,
		state:           state,
		travisTxIndex:   0,
		txIndex:         0,
		totalUsedGas:    big.NewInt(0),
		totalUsedGasFee: big.NewInt(0),
		gp:              new(core.GasPool).AddGas(ethHeader.GasLimit),
	}
	utils.BlockGasFee = big.NewInt(0)
	utils.StateChangeQueue = make([]utils.StateChangeObject, 0)
	utils.TravisTxAddrs = make([]*common.Address, 0)
	return nil
}

func (es *EthState) UpdateHeaderWithTimeInfo(
	config *params.ChainConfig, parentTime uint64, numTx uint64) {

	es.mtx.Lock()
	defer es.mtx.Unlock()

	es.work.updateHeaderWithTimeInfo(config, parentTime, numTx)
}

func (es *EthState) GasLimit() big.Int {
	return big.Int(*es.work.gp)
}

//----------------------------------------------------------------------
// Implements: miner.Pending API (our custom patch to go-ethereum)

// Return a new block and a copy of the state from the latest work.
// #unstable
func (es *EthState) Pending() (*ethTypes.Block, *state.StateDB) {
	es.mtx.Lock()
	defer es.mtx.Unlock()

	return ethTypes.NewBlock(
		es.work.header,
		es.work.transactions,
		nil,
		es.work.receipts,
	), es.work.state.Copy()
}

//----------------------------------------------------------------------
//

// The work struct handles block processing.
// It's updated with each DeliverTx and reset on Commit.
type workState struct {
	header        *ethTypes.Header
	parent        *ethTypes.Block
	state         *state.StateDB
	travisTxIndex int //coped StateChangeObject index in the queue

	txIndex      int
	transactions []*ethTypes.Transaction
	receipts     ethTypes.Receipts
	allLogs      []*ethTypes.Log

	totalUsedGas    *big.Int
	totalUsedGasFee *big.Int
	gp              *core.GasPool
}

// nolint: unparam
func (ws *workState) accumulateRewards(strategy *emtTypes.Strategy) {

	ethash.AccumulateRewards(ws.state, ws.header, []*ethTypes.Header{})
	ws.header.GasUsed = ws.totalUsedGas
}

// Runs ApplyTransaction against the ethereum blockchain, fetches any logs,
// and appends the tx, receipt, and logs.
func (ws *workState) deliverTx(blockchain *core.BlockChain, config *eth.Config,
	chainConfig *params.ChainConfig, blockHash common.Hash,
	tx *ethTypes.Transaction) abciTypes.ResponseDeliverTx {

	delete(utils.NonceCheckedTx, tx.Hash())

	ws.handleStateChangeQueue()
	ws.travisTxIndex = len(utils.StateChangeQueue)

	ws.state.Prepare(tx.Hash(), blockHash, ws.txIndex)
	receipt, usedGas, err := core.ApplyTransaction(
		chainConfig,
		blockchain,
		nil, // defaults to address of the author of the header
		ws.gp,
		ws.state,
		ws.header,
		tx,
		ws.totalUsedGas,
		vm.Config{EnablePreimageRecording: config.EnablePreimageRecording},
	)
	if err != nil {
		return abciTypes.ResponseDeliverTx{Code: errors.CodeTypeInternalErr, Log: err.Error()}
	}

	usedGasFee := big.NewInt(0).Mul(usedGas, tx.GasPrice())
	ws.totalUsedGasFee.Add(ws.totalUsedGasFee, usedGasFee)

	logs := ws.state.GetLogs(tx.Hash())

	ws.txIndex++

	// The slices are allocated in updateHeaderWithTimeInfo
	ws.transactions = append(ws.transactions, tx)
	ws.receipts = append(ws.receipts, receipt)
	ws.allLogs = append(ws.allLogs, logs...)

	return abciTypes.ResponseDeliverTx{Code: abciTypes.CodeTypeOK}
}

// Commit the ethereum state, update the header, make a new block and add it to
// the ethereum blockchain. The application root hash is the hash of the
// ethereum block.
func (ws *workState) commit(blockchain *core.BlockChain, db ethdb.Database) (common.Hash, error) {
	currentHeight := ws.header.Number.Uint64()

	proposalIds := utils.PendingProposal.ReachMin(currentHeight)
	for _, pid := range proposalIds {
		proposal := gov.GetProposalById(pid)
		amount := new(big.Int)
		amount.SetString(proposal.Amount, 10)

		switch gov.CheckProposal(pid) {
		case "approved":
			commons.TransferWithReactor(utils.EmptyAddress, *proposal.To, amount, gov.ProposalReactor{proposal.Id, currentHeight, "Approved"})
		case "rejected":
			commons.TransferWithReactor(utils.EmptyAddress, *proposal.From, amount, gov.ProposalReactor{proposal.Id, currentHeight, "Rejected"})
		default:
			commons.TransferWithReactor(utils.EmptyAddress, *proposal.From, amount, gov.ProposalReactor{proposal.Id, currentHeight, "Expired"})
		}
		utils.PendingProposal.Del(pid)
	}

	ws.handleStateChangeQueue()

	// Commit ethereum state and update the header.
	hashArray, err := ws.state.CommitTo(db, false) // XXX: ugh hardforks
	if err != nil {
		return common.Hash{}, err
	}
	ws.header.Root = hashArray

	for _, log := range ws.allLogs {
		log.BlockHash = hashArray
	}

	// Create block object and compute final commit hash (hash of the ethereum
	// block).
	block := ethTypes.NewBlock(ws.header, ws.transactions, nil, ws.receipts)
	blockHash := block.Hash()

	// Save the block to disk.
	// log.Info("Committing block", "stateHash", hashArray, "blockHash", blockHash)
	_, err = blockchain.InsertChain([]*ethTypes.Block{block})
	if err != nil {
		// log.Info("Error inserting ethereum block in chain", "err", err)
		return common.Hash{}, err
	}
	return blockHash, err
}

func (ws *workState) handleStateChangeQueue() {
	// Iterate to add/sub balance from state
	// ws.travisTxIndex used for recording handled index of queue
	for i := ws.travisTxIndex; i < len(utils.StateChangeQueue); i++ {
		scObj := utils.StateChangeQueue[i]
		if bytes.Compare(scObj.From.Bytes(), utils.EmptyAddress.Bytes()) == 0 {
			if bytes.Compare(scObj.To.Bytes(), utils.EmptyAddress.Bytes()) != 0 {
				ws.state.AddBalance(scObj.To, scObj.Amount)
				if scObj.Reactor != nil {
					scObj.Reactor.React("success", "")
				}
			}
		} else {
			if ws.state.GetBalance(scObj.From).Cmp(scObj.Amount) >= 0 {
				ws.state.SubBalance(scObj.From, scObj.Amount)
				if bytes.Compare(scObj.To.Bytes(), utils.EmptyAddress.Bytes()) != 0 {
					ws.state.AddBalance(scObj.To, scObj.Amount)
				}
				if scObj.Reactor != nil {
					scObj.Reactor.React("success", "")
				}
			} else {
				if scObj.Reactor != nil {
					scObj.Reactor.React("fail", "Insufficient balance")
				}
			}
		}
	}
}

func (ws *workState) updateHeaderWithTimeInfo(
	config *params.ChainConfig, parentTime uint64, numTx uint64) {

	lastBlock := ws.parent
	parentHeader := &ethTypes.Header{
		Difficulty: lastBlock.Difficulty(),
		Number:     lastBlock.Number(),
		Time:       lastBlock.Time(),
	}
	ws.header.Time = new(big.Int).SetUint64(parentTime)
	ws.header.Difficulty = ethash.CalcDifficulty(config, parentTime, parentHeader)
	ws.transactions = make([]*ethTypes.Transaction, 0, numTx)
	ws.receipts = make([]*ethTypes.Receipt, 0, numTx)
	ws.allLogs = make([]*ethTypes.Log, 0, numTx)
}

//----------------------------------------------------------------------

// Create a new block header from the previous block.
func newBlockHeader(receiver common.Address, prevBlock *ethTypes.Block) *ethTypes.Header {
	return &ethTypes.Header{
		Number:     prevBlock.Number().Add(prevBlock.Number(), big.NewInt(1)),
		ParentHash: prevBlock.Hash(),
		//GasLimit:   core.CalcGasLimit(prevBlock),
		GasLimit: calcGasLimit(prevBlock),
		Coinbase: receiver,
	}
}

// CalcGasLimit computes the gas limit of the next block after parent.
// The result may be modified by the caller.
// This is miner strategy, not consensus protocol.
func calcGasLimit(parent *types.Block) *big.Int {
	// 0xF00000000 = 64424509440
	gl := big.NewInt(64424509440)

	return gl
}
