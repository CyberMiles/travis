package api

import (
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/modules/stake"
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

// SendTransaction creates a transaction for the given argument, sign it and broardcast it to tendermint.
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

func (s *CmtRPCService) SendRawTx(encodedTx hexutil.Bytes) (*ctypes.ResultBroadcastTxCommit, error) {
	tx := new(ethTypes.Transaction)
	if err := rlp.DecodeBytes(encodedTx, tx); err != nil {
		return nil, err
	}

	return s.backend.BroadcastTxCommit(tx)
}

/*
func (s *CmtRPCService) GetBlock(height uint64) (*ctypes.ResultBlock, error) {
	h := cast.ToInt64(height)
	return s.backend.localClient.Block(&h)
}
func (s *CmtRPCService) GetTransaction(hash string) (*ctypes.ResultTx, error) {
	bkey, err := hex.DecodeString(cmn.StripHex(hash))
	if err != nil {
		return nil, err
	}
	return s.backend.localClient.Tx(bkey, false)
}
func (s *CmtRPCService) GetTransactionFromBlock(height uint64, index int64) (*ctypes.ResultTx, error) {
	h := cast.ToInt64(height)
	block, err := s.backend.localClient.Block(&h)
	if err != nil {
		return nil, err
	}
	if index >= block.Block.NumTxs {
		return nil, errors.New(fmt.Sprintf("No transaction in block %d, index %d. ", height, index))
	}
	hash := block.Block.Txs[index].Hash()
	return s.GetTransaction(hex.EncodeToString(hash))
}
*/

type DeclareCandidacyArgs struct {
	Nonce       *hexutil.Uint64   `json:"nonce"`
	From        common.Address    `json:"from"`
	PubKey      string            `json:"pubKey"`
	MaxAmount   string            `json:"maxAmount"`
	CompRate    string            `json:"compRate"`
	Description stake.Description `json:"description"`
}

func (s *CmtRPCService) DeclareCandidacy(args DeclareCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	pubKey, err := utils.GetPubKey(args.PubKey)
	if err != nil {
		return nil, err
	}
	tx := stake.NewTxDeclareCandidacy(pubKey, args.MaxAmount, args.CompRate, args.Description)

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
	MaxAmount   string            `json:"maxAmount"`
	Description stake.Description `json:"description"`
}

func (s *CmtRPCService) UpdateCandidacy(args UpdateCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx := stake.NewTxUpdateCandidacy(args.MaxAmount, args.Description)

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

type DelegateArgs struct {
	Nonce            *hexutil.Uint64 `json:"nonce"`
	From             common.Address  `json:"from"`
	ValidatorAddress common.Address  `json:"validatorAddress"`
	Amount           string          `json:"amount"`
}

func (s *CmtRPCService) Delegate(args DelegateArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	if len(args.ValidatorAddress) == 0 {
		return nil, fmt.Errorf("must provide validator address")
	}
	tx := stake.NewTxDelegate(args.ValidatorAddress, args.Amount)

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
	Amount           string          `json:"amount"`
}

func (s *CmtRPCService) Withdraw(args WithdrawArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	if len(args.ValidatorAddress) == 0 {
		return nil, fmt.Errorf("must provide validator address")
	}
	tx := stake.NewTxWithdraw(args.ValidatorAddress, args.Amount)

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

	return &StakeQueryResult{h, candidate}, nil
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
	Amount       string          `json:"amount"`
	Reason       string          `json:"reason"`
	Expire       uint64          `json:"expire"`
}

func (s *CmtRPCService) Propose(args GovernanceProposalArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx := governance.NewTxPropose(&args.From, &args.TransferFrom, &args.TransferTo, args.Amount, args.Reason, args.Expire)

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
