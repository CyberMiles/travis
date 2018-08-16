package api

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	ttypes "github.com/tendermint/tendermint/types"
)

// CmtRPCService offers cmt related RPC methods
type EthRPCService struct {
	backend   *Backend
	am        *accounts.Manager
	nonceLock *AddrLocker
}

func NewEthRPCService(b *Backend, nonceLock *AddrLocker) *EthRPCService {
	return &EthRPCService{
		backend:   b,
		am:        b.ethereum.AccountManager(),
		nonceLock: nonceLock,
	}
}

// sign tx and broardcast sync to tendermint.
func (s *EthRPCService) signAndBroadcastSync(args SendTxArgs) (*types.Transaction, error) {
	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		// release noncelock after broadcast
		defer s.nonceLock.UnlockAddr(args.From)
	}

	signed, err := s.backend.signTransaction(&args)
	if err != nil {
		return nil, err
	}

	result, err := s.backend.BroadcastTxSync(signed)
	if err != nil {
		return nil, err
	}
	if result.Code > 0 {
		return nil, errors.New(result.Log)
	}

	return signed, nil
}

// SendTransaction is compatible with Ethereum, return eth transaction hash
func (s *EthRPCService) SendTransaction(args SendTxArgs) (common.Hash, error) {
	signed, err := s.signAndBroadcastSync(args)
	if err != nil {
		return common.Hash{}, err
	}

	return signed.Hash(), nil
}

// SendTx is same as SendTransaction, but return cmt transaction hash
func (s *EthRPCService) SendTx(args SendTxArgs) (string, error) {
	signed, err := s.signAndBroadcastSync(args)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err := signed.EncodeRLP(buf); err != nil {
		return "", err
	}

	return hexutil.Encode(ttypes.Tx(buf.Bytes()).Hash()), nil
}

// SendRawTransaction will broadcast the signed transaction to tendermint.
// The sender is responsible for signing the transaction and using the correct nonce.
func (s *EthRPCService) SendRawTransaction(encodedTx hexutil.Bytes) (string, error) {
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(encodedTx, tx); err != nil {
		return "", err
	}

	result, err := s.backend.BroadcastTxSync(tx)
	if err != nil {
		return "", err
	}
	if result.Code > 0 {
		return "", errors.New(result.Log)
	}
	return tx.Hash().Hex(), nil
}
