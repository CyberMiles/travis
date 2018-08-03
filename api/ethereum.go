package api

import (
	"bytes"
	"math/big"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	ttypes "github.com/tendermint/tendermint/types"

	"github.com/CyberMiles/travis/utils"
)

const defaultGas = 90000

// SendTxArgs represents the arguments to sumbit a new transaction
type SendTxArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *hexutil.Uint64 `json:"gas"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Value    *hexutil.Big    `json:"value"`
	Nonce    *hexutil.Uint64 `json:"nonce"`
	// We accept "data" and "input" for backwards-compatibility reasons. "input" is the
	// newer name and should be preferred by clients.
	Data  *hexutil.Bytes `json:"data"`
	Input *hexutil.Bytes `json:"input"`
}

// prepareSendTxArgs is a helper function that fills in default values for unspecified tx fields.
func (args *SendTxArgs) setDefaults(b *Backend) error {
	if args.Gas == nil {
		args.Gas = new(hexutil.Uint64)
		*(*uint64)(args.Gas) = defaultGas
	}

	if args.GasPrice == nil {
		price := utils.GetParams().GasPrice
		args.GasPrice = (*hexutil.Big)(new(big.Int).SetUint64(price))
	}
	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	if args.Nonce == nil {
		nonce := b.ManagedState().GetNonce(args.From)
		args.Nonce = (*hexutil.Uint64)(&nonce)
	}
	if args.Data != nil && args.Input != nil && !bytes.Equal(*args.Data, *args.Input) {
		return errors.New(`Both "data" and "input" are set and not equal. Please use "input" to pass transaction call data.`)
	}
	if args.To == nil {
		// Contract creation
		var input []byte
		if args.Data != nil {
			input = *args.Data
		} else if args.Input != nil {
			input = *args.Input
		}
		if len(input) == 0 {
			return errors.New(`contract creation without any data provided`)
		}
	}
	return nil
}

func (args *SendTxArgs) toTransaction() *types.Transaction {
	var input []byte
	if args.Data != nil {
		input = *args.Data
	} else if args.Input != nil {
		input = *args.Input
	}
	if args.To == nil {
		return types.NewContractCreation(uint64(*args.Nonce), (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
	}
	return types.NewTransaction(uint64(*args.Nonce), *args.To, (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
}

// signTransaction sets defaults and signs the given transaction
// NOTE: the caller needs to ensure that the nonceLock is held, and release it after use.
func (s *CmtRPCService) signTransaction(args SendTxArgs) (*types.Transaction, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}

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

	return signed, nil
}

// sign tx and broardcast sync to tendermint.
func (s *CmtRPCService) signAndBroadcastSync(args SendTxArgs) (*types.Transaction, error) {
	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		// release noncelock after broadcast
		defer s.nonceLock.UnlockAddr(args.From)
	}

	signed, err := s.signTransaction(args)
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
func (s *CmtRPCService) SendTransaction(args SendTxArgs) (common.Hash, error) {
	signed, err := s.signAndBroadcastSync(args)
	if err != nil {
		return common.Hash{}, err
	}

	return signed.Hash(), nil
}

// SendTx is same as SendTransaction, but return cmt transaction hash
func (s *CmtRPCService) SendTx(args SendTxArgs) (string, error) {
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
func (s *CmtRPCService) SendRawTransaction(encodedTx hexutil.Bytes) (string, error) {
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
