package api

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/spf13/cast"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	abci "github.com/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	ttypes "github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/sdk"
	"github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
)

// CmtRPCService offers cmt related RPC methods
type CmtRPCService struct {
	backend   *Backend
	am        *accounts.Manager
	nonceLock *AddrLocker
}

func NewCmtRPCService(b *Backend, nonceLock *AddrLocker) *CmtRPCService {
	return &CmtRPCService{
		backend:   b,
		am:        b.ethereum.AccountManager(),
		nonceLock: nonceLock,
	}
}

func (s *CmtRPCService) makeTravisTxArgs(tx sdk.Tx, address common.Address, nonce *hexutil.Uint64) (*SendTxArgs, error) {
	data, err := tx.MarshalJSON()
	if err != nil {
		return nil, err
	}

	zero := big.NewInt(0)
	return &SendTxArgs{
		address,
		nil,
		(*hexutil.Big)(zero),
		(*hexutil.Big)(zero),
		(*hexutil.Big)(zero),
		data,
		nonce,
	}, nil
}

// SendTransaction creates a transaction for the given argument, sign it and broadcast it to tendermint.
func (s *CmtRPCService) sendTransaction(args *SendTxArgs) (*ctypes.ResultBroadcastTxCommit, error) {

	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}

	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(s.backend); err != nil {
		return nil, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	wallet, err := s.am.Find(account)
	if err != nil {
		return nil, err
	}
	ethChainId := int64(s.backend.ethConfig.NetworkId)
	signed, err := wallet.SignTx(account, tx, big.NewInt(ethChainId))
	if err != nil {
		return nil, err
	}

	return s.backend.BroadcastTxCommit(signed)
}

// SendRawTx will broadcast the signed transaction to tendermint.
// The sender is responsible for signing the transaction and using the correct nonce.
func (s *CmtRPCService) SendRawTx(encodedTx hexutil.Bytes) (*ctypes.ResultBroadcastTxCommit, error) {
	tx := new(ethTypes.Transaction)
	if err := rlp.DecodeBytes(encodedTx, tx); err != nil {
		return nil, err
	}

	if utils.IsEthTx(tx) {
		result, err := s.backend.BroadcastTxSync(tx)
		if err != nil {
			return nil, err
		}
		if result.Code > 0 {
			return nil, errors.New(result.Log)
		}

		return &ctypes.ResultBroadcastTxCommit{
			Hash: ttypes.Tx(encodedTx).Hash(), //tx.Hash().Hex(),
		}, nil
	} else {
		return s.backend.BroadcastTxCommit(tx)
	}
}

// GetBlockByNumber returns the requested block by height.
func (s *CmtRPCService) GetBlockByNumber(height uint64) (*ctypes.ResultBlock, error) {
	h := cast.ToInt64(height)
	return s.backend.localClient.Block(&h)
}

// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockNumber      *hexutil.Big           `json:"blockNumber"`
	From             common.Address         `json:"from"`
	Gas              *hexutil.Big           `json:"gas"`
	GasPrice         *hexutil.Big           `json:"gasPrice"`
	Hash             common.Hash            `json:"hash"`
	CmtHash          cmn.HexBytes           `json:"cmt_hash"`
	Input            hexutil.Bytes          `json:"input"`
	CmtInput         interface{}            `json:"cmt_input"`
	Nonce            hexutil.Uint64         `json:"nonce"`
	To               *common.Address        `json:"to"`
	TransactionIndex hexutil.Uint           `json:"transactionIndex"`
	Value            *hexutil.Big           `json:"value"`
	V                *hexutil.Big           `json:"v"`
	R                *hexutil.Big           `json:"r"`
	S                *hexutil.Big           `json:"s"`
	TxResult         abci.ResponseDeliverTx `json:"tx_result"`
}

// newRPCTransaction returns a transaction that will serialize to the RPC representation.
func newRPCTransaction(res *ctypes.ResultTx) (*RPCTransaction, error) {
	tx := new(ethTypes.Transaction)
	rlpStream := rlp.NewStream(bytes.NewBuffer(res.Tx), 0)
	if err := tx.DecodeRLP(rlpStream); err != nil {
		return nil, err
	}

	var signer ethTypes.Signer = ethTypes.FrontierSigner{}
	if tx.Protected() {
		signer = ethTypes.NewEIP155Signer(tx.ChainId())
	}
	from, _ := ethTypes.Sender(signer, tx)
	v, r, s := tx.RawSignatureValues()

	var travisTx sdk.Tx
	if !utils.IsEthTx(tx) {
		if err := json.Unmarshal(tx.Data(), &travisTx); err != nil {
			return nil, err
		}
	}

	return &RPCTransaction{
		BlockNumber:      (*hexutil.Big)(big.NewInt(res.Height)),
		From:             from,
		Gas:              (*hexutil.Big)(tx.Gas()),
		GasPrice:         (*hexutil.Big)(tx.GasPrice()),
		Hash:             tx.Hash(),
		CmtHash:          res.Hash,
		Input:            hexutil.Bytes(tx.Data()),
		CmtInput:         travisTx,
		Nonce:            hexutil.Uint64(tx.Nonce()),
		To:               tx.To(),
		TransactionIndex: hexutil.Uint(res.Index),
		Value:            (*hexutil.Big)(tx.Value()),
		V:                (*hexutil.Big)(v),
		R:                (*hexutil.Big)(r),
		S:                (*hexutil.Big)(s),
		TxResult:         res.TxResult,
	}, nil
}

// GetTransactionFromBlock returns the transaction for the given block number and index.
func (s *CmtRPCService) GetTransactionFromBlock(height uint64, index uint64) (*RPCTransaction, error) {
	// get block
	h := cast.ToInt64(height)
	block, err := s.backend.localClient.Block(&h)
	if err != nil {
		return nil, err
	}
	// check index
	if cast.ToInt64(index) >= block.Block.NumTxs {
		return nil, errors.New(fmt.Sprintf("No transaction in block %d, index %d. ", height, index))
	}
	// get transaction by hash
	tx := ttypes.Tx(block.Block.Txs[index])
	return s.GetTransactionByHash(hex.EncodeToString(tx.Hash()))
}

// GetTransactionByHash returns the transaction for the given hash
func (s *CmtRPCService) GetTransactionByHash(hash string) (*RPCTransaction, error) {
	// bytes from hash string
	bkey, err := hex.DecodeString(cmn.StripHex(hash))
	if err != nil {
		return nil, err
	}
	// get transaction
	res, err := s.backend.localClient.Tx(bkey, false)
	if err != nil {
		return nil, err
	}

	return newRPCTransaction(res)
}

// DecodeRawTx returns the transaction from the raw tx string in the block data
func (s *CmtRPCService) DecodeRawTx(raw string) (*RPCTransaction, error) {
	tx, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	return newRPCTransaction(&ctypes.ResultTx{
		Tx: tx,
	})
}

type DeclareCandidacyArgs struct {
	Nonce       *hexutil.Uint64   `json:"nonce"`
	From        common.Address    `json:"from"`
	PubKey      string            `json:"pubKey"`
	MaxAmount   hexutil.Big       `json:"maxAmount"`
	CompRate    string            `json:"compRate"`
	Description stake.Description `json:"description"`
}

func (s *CmtRPCService) DeclareCandidacy(args DeclareCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	pubKey, err := types.GetPubKey(args.PubKey)
	if err != nil {
		return nil, err
	}
	tx := stake.NewTxDeclareCandidacy(pubKey, args.MaxAmount.ToInt().String(), args.CompRate, args.Description)

	txArgs, err := s.makeTravisTxArgs(tx, args.From, args.Nonce)
	if err != nil {
		return nil, err
	}

	return s.sendTransaction(txArgs)
}

type WithdrawCandidacyArgs struct {
	Nonce *hexutil.Uint64 `json:"nonce"`
	From  common.Address  `json:"from"`
}

func (s *CmtRPCService) WithdrawCandidacy(args WithdrawCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx := stake.NewTxWithdrawCandidacy()

	txArgs, err := s.makeTravisTxArgs(tx, args.From, args.Nonce)
	if err != nil {
		return nil, err
	}

	return s.sendTransaction(txArgs)
}

type UpdateCandidacyArgs struct {
	Nonce       *hexutil.Uint64   `json:"nonce"`
	From        common.Address    `json:"from"`
	MaxAmount   *hexutil.Big      `json:"maxAmount"`
	Description stake.Description `json:"description"`
}

func (s *CmtRPCService) UpdateCandidacy(args UpdateCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	maxAmount := ""
	if args.MaxAmount != nil {
		maxAmount = args.MaxAmount.ToInt().String()
	}
	tx := stake.NewTxUpdateCandidacy(maxAmount, args.Description)

	txArgs, err := s.makeTravisTxArgs(tx, args.From, args.Nonce)
	if err != nil {
		return nil, err
	}

	return s.sendTransaction(txArgs)
}

type VerifyCandidacyArgs struct {
	Nonce            *hexutil.Uint64 `json:"nonce"`
	From             common.Address  `json:"from"`
	CandidateAddress common.Address  `json:"candidateAddress"`
	Verified         bool            `json:"verified"`
}

func (s *CmtRPCService) VerifyCandidacy(args VerifyCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	if len(args.CandidateAddress) == 0 {
		return nil, fmt.Errorf("must provide new address")
	}
	tx := stake.NewTxVerifyCandidacy(args.CandidateAddress, args.Verified)

	txArgs, err := s.makeTravisTxArgs(tx, args.From, args.Nonce)
	if err != nil {
		return nil, err
	}

	return s.sendTransaction(txArgs)
}

type ActivateCandidacyArgs struct {
	Nonce *hexutil.Uint64 `json:"nonce"`
	From  common.Address  `json:"from"`
}

func (s *CmtRPCService) ActivateCandidacy(args ActivateCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx := stake.NewTxActivateCandidacy()

	txArgs, err := s.makeTravisTxArgs(tx, args.From, args.Nonce)
	if err != nil {
		return nil, err
	}

	return s.sendTransaction(txArgs)
}

type DelegateArgs struct {
	Nonce            *hexutil.Uint64 `json:"nonce"`
	From             common.Address  `json:"from"`
	ValidatorAddress common.Address  `json:"validatorAddress"`
	Amount           hexutil.Big     `json:"amount"`
}

func (s *CmtRPCService) Delegate(args DelegateArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	if len(args.ValidatorAddress) == 0 {
		return nil, fmt.Errorf("must provide validator address")
	}
	tx := stake.NewTxDelegate(args.ValidatorAddress, args.Amount.ToInt().String())

	txArgs, err := s.makeTravisTxArgs(tx, args.From, args.Nonce)
	if err != nil {
		return nil, err
	}

	return s.sendTransaction(txArgs)
}

type WithdrawArgs struct {
	Nonce            *hexutil.Uint64 `json:"nonce"`
	From             common.Address  `json:"from"`
	ValidatorAddress common.Address  `json:"validatorAddress"`
	Amount           hexutil.Big     `json:"amount"`
}

func (s *CmtRPCService) Withdraw(args WithdrawArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	if len(args.ValidatorAddress) == 0 {
		return nil, fmt.Errorf("must provide validator address")
	}
	tx := stake.NewTxWithdraw(args.ValidatorAddress, args.Amount.ToInt().String())

	txArgs, err := s.makeTravisTxArgs(tx, args.From, args.Nonce)
	if err != nil {
		return nil, err
	}

	return s.sendTransaction(txArgs)
}

type StakeQueryResult struct {
	Height int64       `json:"height"`
	Data   interface{} `json:"data"`
}

func (s *CmtRPCService) QueryValidators(height uint64) (*StakeQueryResult, error) {
	var candidates stake.Candidates
	//key := stack.PrefixedKey(stake.Name(), stake.CandidatesPubKeysKey)
	h, err := s.getParsed("/validators", []byte{0}, &candidates, height)
	if err != nil {
		return nil, err
	}

	return &StakeQueryResult{h, candidates}, nil
}

func (s *CmtRPCService) QueryValidator(address common.Address, height uint64) (*StakeQueryResult, error) {
	var candidate stake.Candidate
	h, err := s.getParsed("/validator", []byte(address.Hex()), &candidate, height)
	if err != nil {
		return nil, err
	}

	return &StakeQueryResult{h, &candidate}, nil
}

func (s *CmtRPCService) QueryDelegator(address common.Address, height uint64) (*StakeQueryResult, error) {
	var slotDelegates []*stake.Delegation
	h, err := s.getParsed("/delegator", []byte(address.Hex()), &slotDelegates, height)
	if err != nil {
		return nil, err
	}

	return &StakeQueryResult{h, slotDelegates}, nil
}

type GovernanceProposalArgs struct {
	Nonce        *hexutil.Uint64 `json:"nonce"`
	From         common.Address  `json:"from"`
	TransferFrom common.Address  `json:"transferFrom"`
	TransferTo   common.Address  `json:"transferTo"`
	Amount       hexutil.Big     `json:"amount"`
	Reason       string          `json:"reason"`
	Expire       uint64          `json:"expire"`
}

func (s *CmtRPCService) Propose(args GovernanceProposalArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx := governance.NewTxPropose(&args.From, &args.TransferFrom, &args.TransferTo, args.Amount.ToInt().String(), args.Reason, args.Expire)

	txArgs, err := s.makeTravisTxArgs(tx, args.From, args.Nonce)
	if err != err {
		return nil, err
	}

	return s.sendTransaction(txArgs)
}

type GovernanceVoteArgs struct {
	Nonce      *hexutil.Uint64 `json:"nonce"`
	Voter      common.Address  `json:"from"`
	ProposalId string          `json:"proposalId"`
	Answer     string          `json:"answer"`
}

func (s *CmtRPCService) Vote(args GovernanceVoteArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx := governance.NewTxVote(args.ProposalId, args.Voter, args.Answer)

	txArgs, err := s.makeTravisTxArgs(tx, args.Voter, args.Nonce)
	if err != err {
		return nil, err
	}

	return s.sendTransaction(txArgs)
}

func (s *CmtRPCService) QueryProposals() (*StakeQueryResult, error) {
	var proposals []*governance.Proposal
	//key := stack.PrefixedKey(stake.Name(), stake.CandidatesPubKeysKey)
	h, err := s.getParsed("/governance/proposals", []byte{0}, &proposals, 0)
	if err != nil {
		return nil, err
	}

	return &StakeQueryResult{h, proposals}, nil
}
