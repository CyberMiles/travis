package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/CyberMiles/travis/modules/auth"
	"github.com/CyberMiles/travis/modules/governance"
	"github.com/CyberMiles/travis/modules/nonce"
	"github.com/CyberMiles/travis/modules/stake"
	ttypes "github.com/CyberMiles/travis/types"
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

func (s *CmtRPCService) getChainID() (string, error) {
	if s.backend.chainID == "" {
		return "", errors.New("Empty chain id. Please wait for tendermint to finish starting up. ")
	}

	return s.backend.chainID, nil
}

func (s *CmtRPCService) GetSequence(address string) (*uint64, error) {
	signers := []common.Address{getSigner(address)}
	var sequence uint64
	err := s.backend.GetSequence(signers, &sequence)
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
	return s.broadcastTx(tx)
}

func (s *CmtRPCService) SendRawTx(encodedTx hexutil.Bytes) (*ctypes.ResultBroadcastTxCommit, error) {
	var tx sdk.Tx
	err := data.FromWire(encodedTx, &tx)
	if err != nil {
		return nil, err
	}
	return s.broadcastTx(tx)
}

type DeclareCandidacyArgs struct {
	Sequence uint64 `json:"sequence"`
	From     string `json:"from"`
	PubKey   string `json:"pubKey"`
}

func (s *CmtRPCService) DeclareCandidacy(args DeclareCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareDeclareCandidacyTx(args)
	if err != nil {
		return nil, err
	}
	return s.broadcastTx(tx)
}

func (s *CmtRPCService) SignDeclareCandidacy(args DeclareCandidacyArgs) (hexutil.Bytes, error) {
	tx, err := s.prepareDeclareCandidacyTx(args)
	if err != nil {
		return nil, err
	}
	return data.ToWire(tx)
}

func (s *CmtRPCService) prepareDeclareCandidacyTx(args DeclareCandidacyArgs) (sdk.Tx, error) {
	pubKey, err := stake.GetPubKey(args.PubKey)
	if err != nil {
		return sdk.Tx{}, err
	}
	tx := stake.NewTxDeclareCandidacy(pubKey)
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
	return s.broadcastTx(tx)
}

func (s *CmtRPCService) prepareWithdrawCandidacyTx(args WithdrawCandidacyArgs) (sdk.Tx, error) {
	address := common.HexToAddress(args.From)
	tx := stake.NewTxWithdraw(address)
	return s.wrapAndSignTx(tx, args.From, args.Sequence)
}

type EditCandidacyArgs struct {
	Sequence   uint64 `json:"sequence"`
	From       string `json:"from"`
	NewAddress string `json:"newAddress"`
}

func (s *CmtRPCService) EditCandidacy(args EditCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareEditCandidacyTx(args)
	if err != nil {
		return nil, err
	}
	return s.broadcastTx(tx)
}

func (s *CmtRPCService) prepareEditCandidacyTx(args EditCandidacyArgs) (sdk.Tx, error) {
	if len(args.NewAddress) == 0 {
		return sdk.Tx{}, fmt.Errorf("must provide new address")
	}
	address := common.HexToAddress(args.NewAddress)
	tx := stake.NewTxEditCandidacy(address)
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
	return s.broadcastTx(tx)
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
	return s.broadcastTx(tx)
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
	return s.broadcastTx(tx)
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
	return s.broadcastTx(tx)
}

func (s *CmtRPCService) prepareCancelSlotTx(args CancelSlotArgs) (sdk.Tx, error) {
	address := common.HexToAddress(args.From)
	tx := stake.NewTxCancelSlot(address, args.SlotId)
	return s.wrapAndSignTx(tx, args.From, args.Sequence)
}

func (s *CmtRPCService) wrapAndSignTx(tx sdk.Tx, address string, sequence uint64) (sdk.Tx, error) {
	// wrap
	// only add the actual signer to the nonce
	signers := []common.Address{getSigner(address)}
	if sequence <= 0 {
		// calculate default sequence
		err := s.backend.GetSequence(signers, &sequence)
		if err != nil {
			return sdk.Tx{}, err
		}
		sequence = sequence + 1
	}
	tx = nonce.NewTx(sequence, signers, tx)

	/*
		chainID, err := s.getChainID()
		if err != nil {
			return sdk.Tx{}, err
		}
		tx = base.NewChainTx(chainID, 0, tx)
	*/
	tx = auth.NewSig(tx).Wrap()

	// sign
	err := s.signTx(tx, address)
	if err != nil {
		return sdk.Tx{}, err
	}
	return tx, err
}

// sign the transaction with private key
func (s *CmtRPCService) signTx(tx sdk.Tx, address string) error {
	// validate tx client-side
	err := tx.ValidateBasic()
	if err != nil {
		return err
	}

	if sign, ok := tx.Unwrap().(ttypes.Signable); ok {
		if address == "" {
			return errors.New("address is required to sign tx")
		}
		err := s.sign(sign, address)
		if err != nil {
			return err
		}
	}
	return err
}

func (s *CmtRPCService) sign(data ttypes.Signable, address string) error {
	ethTx := types.NewTransaction(
		0,
		common.Address([20]byte{}),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		data.SignBytes(),
	)

	addr := common.HexToAddress(address)
	account := accounts.Account{Address: addr}
	wallet, err := s.am.Find(account)
	if err != nil {
		return err
	}

	ethChainId := int64(s.backend.ethConfig.NetworkId)
	signed, err := wallet.SignTx(account, ethTx, big.NewInt(ethChainId))
	if err != nil {
		return err
	}

	return data.Sign(signed)
}

func (s *CmtRPCService) broadcastTx(tx sdk.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	txBytes := wire.BinaryBytes(tx)
	return s.backend.localClient.BroadcastTxCommit(txBytes)
}

func getSigner(address string) (res common.Address) {
	// this could be much cooler with multisig...
	res = common.HexToAddress(address)
	return res
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

func (s *CmtRPCService) QuerySlots(height uint64) (*StakeQueryResult, error) {
	var slots []*stake.Slot
	h, err := s.getParsed("/slots", []byte{0}, &slots, height)
	if err != nil {
		return nil, err
	}

	return &StakeQueryResult{h, slots}, nil
}

func (s *CmtRPCService) QuerySlot(slotId string, height uint64) (*StakeQueryResult, error) {
	var slot stake.Slot
	h, err := s.getParsed("/slot", []byte(slotId), &slot, height)
	if err != nil {
		return nil, err
	}

	return &StakeQueryResult{h, slot}, nil
}

func (s *CmtRPCService) QueryDelegator(address string, height uint64) (*StakeQueryResult, error) {
	var slotDelegates []*stake.SlotDelegate
	h, err := s.getParsed("/delegator", []byte(address), &slotDelegates, height)
	if err != nil {
		return nil, err
	}

	return &StakeQueryResult{h, slotDelegates}, nil
}

func (s *CmtRPCService) getParsed(path string, key []byte, data interface{}, height uint64) (int64, error) {
	bs, h, err := s.get(path, key, cast.ToInt64(height))
	if err != nil {
		return 0, err
	}
	if len(bs) == 0 {
		return h, client.ErrNoData()
	}
	err = wire.ReadBinaryBytes(bs, data)
	if err != nil {
		return 0, err
	}
	return h, nil
}

func (s *CmtRPCService) get(path string, key []byte, height int64) (data.Bytes, int64, error) {
	node := s.backend.localClient
	resp, err := node.ABCIQueryWithOptions(path, key,
		rpcclient.ABCIQueryOptions{Trusted: true, Height: int64(height)})
	if resp == nil {
		return nil, height, err
	}
	return data.Bytes(resp.Response.Value), resp.Response.Height, err
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

	return s.broadcastTx(tx)
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
	return s.broadcastTx(tx)
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
