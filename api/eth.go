package api

import (
	"bytes"
	"math/big"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/tendermint/tendermint/rpc/core"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
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

// GetTransactionCount returns the number of transactions sent from the given address.
// blockNr is useless, be compatible with eth call
func (s *EthRPCService) GetTransactionCount(address common.Address, blockNr rpc.BlockNumber) (*hexutil.Uint64, error) {
	state := s.backend.ManagedState()
	nonce := state.GetNonce(address)

	return (*hexutil.Uint64)(&nonce), nil
}

func newEthRPCTransaction(tx *types.Transaction, blockNumber uint64, index uint64) *RPCTransaction {
	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	from, _ := types.Sender(signer, tx)
	v, r, s := tx.RawSignatureValues()

	result := &RPCTransaction{
		From:     from,
		Gas:      hexutil.Uint64(tx.Gas()),
		GasPrice: (*hexutil.Big)(tx.GasPrice()),
		Hash:     tx.Hash(),
		Input:    hexutil.Bytes(tx.Data()),
		Nonce:    hexutil.Uint64(tx.Nonce()),
		To:       tx.To(),
		Value:    (*hexutil.Big)(tx.Value()),
		V:        (*hexutil.Big)(v),
		R:        (*hexutil.Big)(r),
		S:        (*hexutil.Big)(s),
	}
	if blockNumber > 0 {
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = hexutil.Uint(index)
	}
	return result
}

// GetTransactionByHash returns the transaction for the given hash
func (s *EthRPCService) GetTransactionByHash(hash common.Hash) *RPCTransaction {
	// Try to return an already finalized transaction
	if tx, _, blockNumber, index := rawdb.ReadTransaction(s.backend.Ethereum().ChainDb(), hash); tx != nil {
		return newEthRPCTransaction(tx, blockNumber, index)
	}

	// No finalized transaction, try to retrieve it from the pool
	unConfirmedTxs, err := core.UnconfirmedTxs(-1)
	if err != nil {
		return nil
	}

	for _, tx := range unConfirmedTxs.Txs {
		rpcTx, err := newRPCTransaction(&ctypes.ResultTx{Tx: ttypes.Tx(tx)})
		if err != nil {
			return nil
		}
		return rpcTx
	}
	// Transaction unknown, return as such
	return nil
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

// PrivateAccountAPI provides an API to access accounts managed by this node.
// It offers methods to create, (un)lock en list accounts. Some methods accept
// passwords and are therefore considered private by default.
type PrivateAccountAPI struct {
	am        *accounts.Manager
	nonceLock *AddrLocker
	backend   *Backend
}

// NewPrivateAccountAPI create a new PrivateAccountAPI.
func NewPrivateAccountAPI(b *Backend, nonceLock *AddrLocker) *PrivateAccountAPI {
	return &PrivateAccountAPI{
		am:        b.ethereum.AccountManager(),
		nonceLock: nonceLock,
		backend:   b,
	}
}

// sign tx and broardcast sync to tendermint.
func (s *PrivateAccountAPI) signAndBroadcastSync(args SendTxArgs, passwd string) (*types.Transaction, error) {
	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		// release noncelock after broadcast
		defer s.nonceLock.UnlockAddr(args.From)
	}

	signed, err := s.backend.signTransactionWithPassphrase(&args, passwd)
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

// SendTransaction is compatible with Ethereum, return eth transaction hash.
// It will create a transaction from the given arguments and try to sign it
// with the key associated with args.From. If the given passwd isn't
// able to decrypt the key it fails.
func (s *PrivateAccountAPI) SendTransaction(args SendTxArgs, passwd string) (common.Hash, error) {
	signed, err := s.signAndBroadcastSync(args, passwd)
	if err != nil {
		return common.Hash{}, err
	}

	return signed.Hash(), nil
}
