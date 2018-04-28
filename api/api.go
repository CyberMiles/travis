package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tendermint/go-wire/data"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/modules/stake"
	ttypes "github.com/CyberMiles/travis/types"
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

func (s *CmtRPCService) GetSequence(address string) (*uint64, error) {
	signers := []common.Address{getSigner(address)}
	var sequence uint64
	err := s.getSequence(signers, &sequence)
	return &sequence, err
}

func (s *CmtRPCService) Test(encodedTx hexutil.Bytes) (*ctypes.ResultBroadcastTxCommit, error) {
	var tx sdk.Tx
	err := json.Unmarshal(encodedTx, &tx)
	if err != nil {
		return nil, err
	}
	//d, err := data.ToJSON(tx2)
	//if err != nil {
	//	return nil, err
	//}
	//fmt.Printf("%s\n", d)
	//
	return s.backend.broadcastSdkTx(tx)
}

func (s *CmtRPCService) SendRawTx(encodedTx hexutil.Bytes) (*ctypes.ResultBroadcastTxCommit, error) {
	var tx sdk.Tx
	err := data.FromWire(encodedTx, &tx)
	if err != nil {
		return nil, err
	}
	return s.backend.broadcastSdkTx(tx)
}

type DeclareCandidacyArgs struct {
	Sequence uint64 `json:"sequence"`
	From     string `json:"from"`
	PubKey   string `json:"pubKey"`
	MaxAmount   string            `json:max_amount`
	Cut         int64             `json:"cut"`
	Description stake.Description `json:"description"`
}

func (s *CmtRPCService) DeclareCandidacy(args DeclareCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareDeclareCandidacyTx(args)
	if err != nil {
		return nil, err
	}
	return s.backend.broadcastSdkTx(tx)
}

func (s *CmtRPCService) SignDeclareCandidacy(args DeclareCandidacyArgs) (hexutil.Bytes, error) {
	tx, err := s.prepareDeclareCandidacyTx(args)
	if err != nil {
		return nil, err
	}
	return data.ToWire(tx)
}

func (s *CmtRPCService) prepareDeclareCandidacyTx(args DeclareCandidacyArgs) (sdk.Tx, error) {
	pubKey, err := utils.GetPubKey(args.PubKey)
	if err != nil {
		return sdk.Tx{}, err
	}
	tx := stake.NewTxDeclareCandidacy(pubKey, args.MaxAmount, args.Cut, args.Description)
	return s.wrapAndSignTx(tx, args.From, args.Sequence)
}

type WithdrawCandidacyArgs struct {
	Sequence uint64 `json:"sequence"`
	From     string `json:"from"`
}

func (s *CmtRPCService) WithdrawCandidacy(args WithdrawCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareWithdrawCandidacyTx(args)
	if err != nil {
		return nil, err
	}
	return s.backend.broadcastSdkTx(tx)
}

func (s *CmtRPCService) prepareWithdrawCandidacyTx(args WithdrawCandidacyArgs) (sdk.Tx, error) {
	tx := stake.NewTxWithdrawCandidacy()
	return s.wrapAndSignTx(tx, args.From, args.Sequence)
}

type UpdateCandidacyArgs struct {
	Sequence    uint64            `json:"sequence"`
	From        string            `json:"from"`
	NewAddress  string            `json:"newAddress"`
	MaxAmount   string            `json:"max_amount"`
	Description stake.Description `json:"description"`
}

func (s *CmtRPCService) UpdateCandidacy(args UpdateCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareUpdatteCandidacyTx(args)
	if err != nil {
		return nil, err
	}
	return s.backend.broadcastSdkTx(tx)
}

func (s *CmtRPCService) prepareUpdateCandidacyTx(args UpdateCandidacyArgs) (sdk.Tx, error) {
	if len(args.NewAddress) == 0 {
		return sdk.Tx{}, fmt.Errorf("must provide new address")
	}
	address := common.HexToAddress(args.NewAddress)
	tx := stake.NewTxEditCandidacy(address, args.MaxAmount, args.Description)
	return s.wrapAndSignTx(tx, args.From, args.Sequence)
}

type ProposeSlotArgs struct {
	Sequence    uint64 `json:"sequence"`
	From        string `json:"from"`
	Amount      int64  `json:"amount"`
	ProposedRoi int64  `json:"proposedRoi"`
}

func (s *CmtRPCService) ProposeSlot(args ProposeSlotArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareProposeSlotTx(args)
	if err != nil {
		return nil, err
	}
	return s.backend.broadcastSdkTx(tx)
}

func (s *CmtRPCService) prepareProposeSlotTx(args ProposeSlotArgs) (sdk.Tx, error) {
	address := common.HexToAddress(args.From)
	tx := stake.NewTxProposeSlot(address, args.Amount, args.ProposedRoi)
	return s.wrapAndSignTx(tx, args.From, args.Sequence)
}

type AcceptSlotArgs struct {
	Sequence uint64 `json:"sequence"`
	From     string `json:"from"`
	Amount   int64  `json:"amount"`
	SlotId   string `json:"slotId"`
}

func (s *CmtRPCService) AcceptSlot(args AcceptSlotArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareAcceptSlotTx(args)
	if err != nil {
		return nil, err
	}
	return s.backend.broadcastSdkTx(tx)
}

func (s *CmtRPCService) prepareAcceptSlotTx(args AcceptSlotArgs) (sdk.Tx, error) {
	tx := stake.NewTxAcceptSlot(args.Amount, args.SlotId)
	return s.wrapAndSignTx(tx, args.From, args.Sequence)
}

type WithdrawSlotArgs struct {
	Sequence uint64 `json:"sequence"`
	From     string `json:"from"`
	Amount   int64  `json:"amount"`
	SlotId   string `json:"slotId"`
}

func (s *CmtRPCService) WithdrawSlot(args WithdrawSlotArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareWithdrawSlotTx(args)
	if err != nil {
		return nil, err
	}
	return s.backend.broadcastSdkTx(tx)
}

func (s *CmtRPCService) prepareWithdrawSlotTx(args WithdrawSlotArgs) (sdk.Tx, error) {
	tx := stake.NewTxWithdrawSlot(args.Amount, args.SlotId)
	return s.wrapAndSignTx(tx, args.From, args.Sequence)
}

type CancelSlotArgs struct {
	Sequence uint64 `json:"sequence"`
	From     string `json:"from"`
	SlotId   string `json:"slotId"`
}

func (s *CmtRPCService) CancelSlot(args CancelSlotArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareCancelSlotTx(args)
	if err != nil {
		return nil, err
	}
	return s.backend.broadcastSdkTx(tx)
}

func (s *CmtRPCService) prepareCancelSlotTx(args CancelSlotArgs) (sdk.Tx, error) {
	address := common.HexToAddress(args.From)
	tx := stake.NewTxCancelSlot(address, args.SlotId)
	tx := stake.NewTxUpdateCandidacy(address, args.MaxAmount, args.Description)
	return s.wrapAndSignTx(tx, args.From, args.Sequence)
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

func (s *CmtRPCService) QueryValidator(address string, height uint64) (*StakeQueryResult, error) {
	var candidate stake.Candidate
	h, err := s.getParsed("/validator", []byte(address), &candidate, height)
	if err != nil {
		return nil, err
	}

	return &StakeQueryResult{h, candidate}, nil
}

func (s *CmtRPCService) QueryDelegator(address string, height uint64) (*StakeQueryResult, error) {
	var delegation *stake.Delegation
	h, err := s.getParsed("/delegator", []byte(address), &delegation, height)
	if err != nil {
		return nil, err
	}

	return &StakeQueryResult{h, delegation}, nil
}

type GovernanceProposalArgs struct {
	Sequence uint64          `json:"sequence"`
	Proposer *common.Address `json:"from"`
	From     *common.Address `json:"transferFrom"`
	To       *common.Address `json:"transferTo"`
	Amount   string          `json:"amount"`
	Reason   string          `json:"reason"`
}

func (s *CmtRPCService) Propose(args GovernanceProposalArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx := governance.NewTxPropose(args.Proposer, args.From, args.To, args.Amount, args.Reason)
	tx, err := s.wrapAndSignTx(tx, args.Proposer.String(), args.Sequence)

	if err != err {
		return nil, err
	}

	return s.backend.broadcastSdkTx(tx)
}

type GovernanceVoteArgs struct {
	Sequence   uint64 `json:"sequence"`
	ProposalId string `json:"proposalId"`
	Voter      string `json:"from"`
	Answer     string `json:"answer"`
}

func (s *CmtRPCService) Vote(args GovernanceVoteArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	voter := common.HexToAddress(args.Voter)

	tx := governance.NewTxVote(args.ProposalId, voter, args.Answer)
	tx, err := s.wrapAndSignTx(tx, args.Voter, args.Sequence)

	if err != err {
		return nil, err
	}
	return s.backend.broadcastSdkTx(tx)
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
