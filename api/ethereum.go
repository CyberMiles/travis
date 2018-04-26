package api

import (
	"math/big"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// SendTxArgs represents the arguments to sumbit a new transaction
type SendTxArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *hexutil.Big    `json:"gas"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Value    *hexutil.Big    `json:"value"`
	Data     hexutil.Bytes   `json:"data"`
	Nonce    *hexutil.Uint64 `json:"nonce"`
}

// prepareSendTxArgs is a helper function that fills in default values for unspecified tx fields.
func (args *SendTxArgs) setDefaults(b *Backend) error {
	if args.Gas == nil {
		args.Gas = (*hexutil.Big)(big.NewInt(90000))
	}

	if args.GasPrice == nil {
		price, err := b.Ethereum().ApiBackend.SuggestPrice(nil)
		if err != nil {
			return err
		}
		args.GasPrice = (*hexutil.Big)(price)
	}
	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	if args.Nonce == nil {
		nonce := b.ManagedState().GetNonce(args.From)
		args.Nonce = (*hexutil.Uint64)(&nonce)
	}
	return nil
}

func (args *SendTxArgs) toTransaction() *types.Transaction {
	if args.To == nil {
		return types.NewContractCreation(uint64(*args.Nonce), (*big.Int)(args.Value), (*big.Int)(args.Gas), (*big.Int)(args.GasPrice), args.Data)
	}
	return types.NewTransaction(uint64(*args.Nonce), *args.To, (*big.Int)(args.Value), (*big.Int)(args.Gas), (*big.Int)(args.GasPrice), args.Data)
}

// SendTransaction creates a transaction for the given argument, sign it and broardcast it to tendermint.
func (s *CmtRPCService) SendTransaction(args SendTxArgs) (common.Hash, error) {

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
		return common.Hash{}, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	wallet, err := s.am.Find(account)
	if err != nil {
		return common.Hash{}, err
	}
	ethChainId := int64(s.backend.ethConfig.NetworkId)
	signed, err := wallet.SignTx(account, tx, big.NewInt(ethChainId))
	if err != nil {
		return common.Hash{}, err
	}

	result, err := s.backend.BroadcastEthTx(signed)
	if err != nil {
		return common.Hash{}, err
	}
	if result.Code > 0 {
		return common.Hash{}, errors.New(result.Log)
	}
	return signed.Hash(), nil
}

// SendRawTransaction will broadcast the signed transaction to tendermint.
// The sender is responsible for signing the transaction and using the correct nonce.
func (s *CmtRPCService) SendRawTransaction(encodedTx hexutil.Bytes) (string, error) {
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(encodedTx, tx); err != nil {
		return "", err
	}

	result, err := s.backend.BroadcastEthTx(tx)
	if err != nil {
		return "", err
	}
	if result.Code > 0 {
		return "", errors.New(result.Log)
	}
	return tx.Hash().Hex(), nil
}
