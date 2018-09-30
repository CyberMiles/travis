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
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	abciTypes "github.com/tendermint/tendermint/abci/types"

	"github.com/CyberMiles/travis/commons"
	"github.com/CyberMiles/travis/errors"
	gov "github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/sdk"
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
	chainConfig := es.ethereum.APIBackend.ChainConfig()
	blockHash := common.Hash{}
	return es.work.deliverTx(blockchain, es.ethConfig, chainConfig, blockHash, tx)
}

// Accumulate validator rewards.
func (es *EthState) AccumulateRewards(config *params.ChainConfig, strategy *emtTypes.Strategy) {
	es.mtx.Lock()
	defer es.mtx.Unlock()

	es.work.accumulateRewards(config, strategy)
}

// Commit and reset the work.
func (es *EthState) Commit(receiver common.Address) (common.Hash, error) {
	es.mtx.Lock()
	defer es.mtx.Unlock()

	blockHash, err := es.work.commit(es.ethereum.BlockChain(), es.ethereum.ChainDb(), receiver)
	es.resetWorkState(receiver)

	return blockHash, err
}

func (es *EthState) EndBlock() {
	utils.BlockGasFee = big.NewInt(0).Add(utils.BlockGasFee, es.work.totalUsedGasFee)
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
		es:              es,
		header:          ethHeader,
		parent:          currentBlock,
		state:           state,
		travisTxIndex:   0,
		txIndex:         0,
		totalUsedGas:    new(uint64),
		totalUsedGasFee: big.NewInt(0),
		gp:              new(core.GasPool).AddGas(ethHeader.GasLimit),
	}
	utils.StateChangeQueue = make([]utils.StateChangeObject, 0)
	return nil
}

func (es *EthState) GetEthState() *state.StateDB {
	return es.work.state
}

func (es *EthState) UpdateHeaderWithTimeInfo(
	config *params.ChainConfig, parentTime uint64, numTx uint64) {

	es.mtx.Lock()
	defer es.mtx.Unlock()

	es.work.updateHeaderWithTimeInfo(config, parentTime, numTx)
}

func (es *EthState) GasLimit() *core.GasPool {
	return es.work.gp
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
	es            *EthState
	header        *ethTypes.Header
	parent        *ethTypes.Block
	state         *state.StateDB
	travisTxIndex int //coped StateChangeObject index in the queue

	txIndex      int
	transactions []*ethTypes.Transaction
	receipts     ethTypes.Receipts
	allLogs      []*ethTypes.Log

	totalUsedGas    *uint64
	totalUsedGasFee *big.Int
	gp              *core.GasPool
}

// nolint: unparam
func (ws *workState) accumulateRewards(config *params.ChainConfig, strategy *emtTypes.Strategy) {

	ethash.AccumulateRewards(config, ws.state, ws.header, []*ethTypes.Header{})
	ws.header.GasUsed = *ws.totalUsedGas
}

// Runs ApplyTransaction against the ethereum blockchain, fetches any logs,
// and appends the tx, receipt, and logs.
func (ws *workState) deliverTx(blockchain *core.BlockChain, config *eth.Config,
	chainConfig *params.ChainConfig, blockHash common.Hash,
	tx *ethTypes.Transaction) abciTypes.ResponseDeliverTx {

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

	usedGasFee := big.NewInt(0).Mul(new(big.Int).SetUint64(usedGas), tx.GasPrice())
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
func (ws *workState) commit(blockchain *core.BlockChain, db ethdb.Database, receiver common.Address) (common.Hash, error) {
	currentHeight := ws.header.Number.Int64()

	proposalIds := utils.PendingProposal.ReachMin(ws.parent.Time().Int64(), currentHeight)
	for _, pid := range proposalIds {
		proposal := gov.GetProposalById(pid)

		switch proposal.Type {
		case gov.TRANSFER_FUND_PROPOSAL:
			amount, _ := sdk.NewIntFromString(proposal.Detail["amount"].(string))
			switch gov.CheckProposal(pid, nil) {
			case "approved":
				commons.TransferWithReactor(utils.GovHoldAccount, *proposal.Detail["to"].(*common.Address), amount, gov.ProposalReactor{proposal.Id, currentHeight, "Approved"})
			case "rejected":
				commons.TransferWithReactor(utils.GovHoldAccount, *proposal.Detail["from"].(*common.Address), amount, gov.ProposalReactor{proposal.Id, currentHeight, "Rejected"})
			default:
				commons.TransferWithReactor(utils.GovHoldAccount, *proposal.Detail["from"].(*common.Address), amount, gov.ProposalReactor{proposal.Id, currentHeight, "Expired"})
			}
		case gov.CHANGE_PARAM_PROPOSAL:
			switch gov.CheckProposal(pid, nil) {
			case "approved":
				utils.SetParam(proposal.Detail["name"].(string), proposal.Detail["value"].(string))
				gov.ProposalReactor{proposal.Id, currentHeight, "Approved"}.React("success", "")
			case "rejected":
				gov.ProposalReactor{proposal.Id, currentHeight, "Rejected"}.React("success", "")
			default:
				gov.ProposalReactor{proposal.Id, currentHeight, "Expired"}.React("success", "")
			}
		case gov.DEPLOY_LIBENI_PROPOSAL:
			if proposal.Result == "Approved" {
				if proposal.Detail["status"] != "ready" {
					gov.CancelDownload(proposal, true)
				} else {
					gov.RegisterLibEni(proposal)
					gov.UpdateDeployLibEniStatus(proposal.Id, "deployed")
				}
			} else {
				switch gov.CheckProposal(pid, nil) {
				case "approved":
					if proposal.Detail["status"] != "ready" {
						gov.CancelDownload(proposal, true)
					} else {
						gov.RegisterLibEni(proposal)
						gov.UpdateDeployLibEniStatus(proposal.Id, "deployed")
					}
					gov.ProposalReactor{proposal.Id, currentHeight, "Approved"}.React("success", "")
				case "rejected":
					if proposal.Detail["status"] != "ready" {
						gov.CancelDownload(proposal, false)
					}
					gov.ProposalReactor{proposal.Id, currentHeight, "Rejected"}.React("success", "")
				default:
					if proposal.Detail["status"] != "ready" {
						gov.CancelDownload(proposal, false)
					}
					gov.ProposalReactor{proposal.Id, currentHeight, "Expired"}.React("success", "")
				}
			}
		case gov.RETIRE_PROGRAM_PROPOSAL:
			if proposal.Result == "Approved" {
				// process will be killed at next block
				utils.RetiringProposalId = pid
			} else {
				switch gov.CheckProposal(pid, nil) {
					case "approved":
						// process will be killed at next block
						utils.RetiringProposalId = pid
						gov.ProposalReactor{proposal.Id, currentHeight, "Approved"}.React("success", "")
					case "rejected":
						gov.ProposalReactor{proposal.Id, currentHeight, "Rejected"}.React("success", "")
					default:
						gov.ProposalReactor{proposal.Id, currentHeight, "Expired"}.React("success", "")
				}
			}
		case gov.UPGRADE_PROGRAM_PROPOSAL:
			if proposal.Result == "Approved" {
				// Upgrade program command to new version
				gov.UpgradeProgramCmd(proposal)
			} else {
				switch gov.CheckProposal(pid, nil) {
				case "approved":
					// Upgrade program command to new version
					gov.UpgradeProgramCmd(proposal)
					gov.ProposalReactor{proposal.Id, currentHeight, "Approved"}.React("success", "")
				case "rejected":
					gov.ProposalReactor{proposal.Id, currentHeight, "Rejected"}.React("success", "")
				default:
					gov.ProposalReactor{proposal.Id, currentHeight, "Expired"}.React("success", "")
				}
			}
		}

		utils.PendingProposal.Del(pid)
	}

	ws.handleStateChangeQueue()

	// Commit ethereum state and update the header.
	hashArray, err := ws.state.Commit(false) // XXX: ugh hardforks
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
		log.Info("Error inserting ethereum block in chain", "err", err)

		ws.es.resetWorkState(receiver)

		pt := ws.parent.Time()
		pt = pt.Add(pt, big.NewInt(1))
		config := ws.es.ethereum.APIBackend.ChainConfig()
		ws.updateHeaderWithTimeInfo(config, pt.Uint64(), 0)

		hashArray, er := ws.state.Commit(false) // XXX: ugh hardforks
		if er != nil {
			return common.Hash{}, er
		}
		ws.header.Root = hashArray

		for _, log := range ws.allLogs {
			log.BlockHash = hashArray
		}

		// Create block object and compute final commit hash (hash of the ethereum
		// block).
		block = ethTypes.NewBlock(ws.header, ws.transactions, nil, ws.receipts)
		blockHash = block.Hash()
		_, er = blockchain.InsertChain([]*ethTypes.Block{block})
		if er != nil {
			return blockHash, er
		}
	}
	return blockHash, err
}

func (ws *workState) handleStateChangeQueue() {
	// Iterate to add/sub balance from state
	// ws.travisTxIndex used for recording handled index of queue
	for i := ws.travisTxIndex; i < len(utils.StateChangeQueue); i++ {
		scObj := utils.StateChangeQueue[i]
		if bytes.Compare(scObj.From.Bytes(), utils.MintAccount.Bytes()) == 0 {
			if bytes.Compare(scObj.To.Bytes(), utils.MintAccount.Bytes()) != 0 {
				ws.state.AddBalance(scObj.To, scObj.Amount.Int)
				if scObj.Reactor != nil {
					scObj.Reactor.React("success", "")
				}
			}
		} else {
			if ws.state.GetBalance(scObj.From).Cmp(scObj.Amount.Int) >= 0 {
				ws.state.SubBalance(scObj.From, scObj.Amount.Int)
				if bytes.Compare(scObj.To.Bytes(), utils.MintAccount.Bytes()) != 0 {
					ws.state.AddBalance(scObj.To, scObj.Amount.Int)
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
		GasLimit: calcGasLimit(prevBlock),
		Coinbase: receiver,
	}
}

// CalcGasLimit computes the gas limit of the next block after parent.
// The result may be modified by the caller.
// This is miner strategy, not consensus protocol.
func calcGasLimit(parent *types.Block) uint64 {
	// Ethereum average block gasLimit * 1000
	var gl uint64 = 8192000000 // 8192m
	return gl
}
