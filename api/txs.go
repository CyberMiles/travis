package api

import (
	"bytes"
	"math/big"
	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

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

func (args *SendTxArgs) toTransaction() *ethTypes.Transaction {
	var input []byte
	if args.Data != nil {
		input = *args.Data
	} else if args.Input != nil {
		input = *args.Input
	}
	if args.To == nil {
		return ethTypes.NewContractCreation(uint64(*args.Nonce), (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
	}
	return ethTypes.NewTransaction(uint64(*args.Nonce), *args.To, (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
}

// BroadcastTx broadcasts a transaction to tendermint core
// #unstable
func (b *Backend) BroadcastTxSync(tx *ethTypes.Transaction) (*ctypes.ResultBroadcastTx, error) {
	buf := new(bytes.Buffer)
	if err := tx.EncodeRLP(buf); err != nil {
		return nil, err
	}

	return b.GetLocalClient().BroadcastTxSync(buf.Bytes())
}

func (b *Backend) BroadcastTxCommit(tx *ethTypes.Transaction) (*ctypes.ResultBroadcastTxCommit, error) {
	buf := new(bytes.Buffer)
	if err := tx.EncodeRLP(buf); err != nil {
		return nil, err
	}

	return b.GetLocalClient().BroadcastTxCommit(buf.Bytes())
}

// signTransaction sets defaults and signs the given transaction
// NOTE: the caller needs to ensure that the nonceLock is held, and release it after use.
func (b *Backend) signTransaction(args *SendTxArgs) (*ethTypes.Transaction, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(b); err != nil {
		return nil, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	wallet, err := b.ethereum.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	ethChainId := int64(b.ethConfig.NetworkId)
	signed, err := wallet.SignTx(account, tx, big.NewInt(ethChainId))
	if err != nil {
		return nil, err
	}

	return signed, nil
}

// signTransaction sets defaults and signs the given transaction
// NOTE: the caller needs to ensure that the nonceLock is held, and release it after use.
func (b *Backend) signTransactionWithPassphrase(args *SendTxArgs, passwd string) (*ethTypes.Transaction, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(b); err != nil {
		return nil, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	wallet, err := b.ethereum.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	ethChainId := int64(b.ethConfig.NetworkId)
	signed, err := wallet.SignTxWithPassphrase(account, passwd, tx, big.NewInt(ethChainId))
	if err != nil {
		return nil, err
	}

	return signed, nil
}

//----------------------------------------------------------------------
// wait for Tendermint to open the socket and run http endpoint

func waitForServer(c *rpcClient.Local) {
	for {
		_, err := c.Status()
		if err == nil {
			break
		}

		log.Info("Waiting for tendermint endpoint to start", "err", err)
		time.Sleep(time.Second * 3)
	}
}
