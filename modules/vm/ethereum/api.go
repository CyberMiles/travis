package ethereum

import (
	"errors"
	"fmt"
	//"math/big"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	//"github.com/ethereum/go-ethereum/core/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	//"github.com/ethereum/go-ethereum/accounts"
	"github.com/tendermint/go-wire"

	"github.com/CyberMiles/travis/modules/auth"
	"github.com/CyberMiles/travis/modules/coin"
	"github.com/CyberMiles/travis/modules/keys"
	"github.com/CyberMiles/travis/modules/nonce"
	"github.com/CyberMiles/travis/modules/stake"
)

// We must implement our own net service since we don't have access to `internal/ethapi`

// NetRPCService mirrors the implementation of `internal/ethapi`
// #unstable
type NetRPCService struct {
	networkVersion uint64
}

// NewNetRPCService creates a new net API instance.
// #unstable
func NewNetRPCService(networkVersion uint64) *NetRPCService {
	return &NetRPCService{networkVersion}
}

// Listening returns an indication if the node is listening for network connections.
// #unstable
func (n *NetRPCService) Listening() bool {
	return true // always listening
}

// PeerCount returns the number of connected peers
// #unstable
func (n *NetRPCService) PeerCount() hexutil.Uint {
	return hexutil.Uint(0)
}

// Version returns the current ethereum protocol version.
// #unstable
func (n *NetRPCService) Version() string {
	return fmt.Sprintf("%d", n.networkVersion)
}

// StakeRPCService offers stake related RPC methods
type StakeRPCService struct {
	backend *Backend
}

// NewStakeRPCAPI create a new StakeRPCAPI.
func NewStakeRPCService(b *Backend) *StakeRPCService {
	return &StakeRPCService{b}
}

type DeclareCandidacyArgs struct {
	Sequence    uint32            `json:"sequence"`
	From        string            `json:"from"`
	PubKey      string            `json:"pubKey"`
	Amount      coin.Coin         `json:"amount"`
	Description stake.Description `json:"description"`
}

func (n *StakeRPCService) DeclareCandidacy(args DeclareCandidacyArgs, password string) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := n.prepareDelegateTx(args, password)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return n.postTx(tx)
}

func (n *StakeRPCService) prepareDelegateTx(di DeclareCandidacyArgs, password string) (sdk.Tx, error) {
	pubKey, _ := stake.GetPubKey(di.PubKey)
	tx := stake.NewTxDeclareCandidacy(di.Amount, pubKey, di.Description)

	// only add the actual signer to the nonce
	signers := []sdk.Actor{getSignerAct(di.From)}
	//seq, _, err := doNonceQuery(signers)
	//if err != nil {
	//	return sdk.Tx{}, err
	//}
	////increase the sequence by 1!
	//seq++
	tx = nonce.NewTx(di.Sequence, signers, tx)
	if n.backend.tmNode == nil {
		return sdk.Tx{}, errors.New("waiting for tendermint to finish starting up")
	}
	tx = base.NewChainTx(n.backend.tmNode.GenesisDoc().ChainID, 0, tx)
	tx = auth.NewSig(tx).Wrap()

	err := signTx(tx, di.From, password)
	if err != nil {
		return sdk.Tx{}, err
	}

	return tx, nil
}

func getSignerAct(address string) (res sdk.Actor) {
	// this could be much cooler with multisig...
	signer := common.HexToAddress(address)
	res = auth.SigPerm(signer.Bytes())
	return res
}

func signTx(tx sdk.Tx, address string, password string) error {
	// validate tx client-side
	err := tx.ValidateBasic()
	if err != nil {
		return err
	}

	if sign, ok := tx.Unwrap().(keys.Signable); ok {
		if address == "" {
			return errors.New("address is required to sign tx")
		}
		/*
			ethTx := types.NewTransaction(
				0,
				common.Address([20]byte{}),
				big.NewInt(0),
				big.NewInt(0),
				big.NewInt(0),
				sign.SignBytes(),
			)

			am, _, _ := auth.MakeAccountManager()
			addr := common.HexToAddress(address)

			account := accounts.Account{Address: addr}
			wallet, err := am.Find(account)
			signed, err := wallet.SignTx(account, ethTx, big.NewInt(111))
			if err != nil {
				fmt.Errorf("error")
			}
			sign.Sign(signed)
		*/
		err := auth.Sign(sign, address, password)
		if err != nil {
			return err
		}
	}
	return err
}

func doNonceQuery(signers []sdk.Actor) (sequence uint32, height int64, err error) {
	key := stack.PrefixedKey(nonce.NameNonce, nonce.GetSeqKey(signers))
	height, err = query.GetParsed(key, &sequence, query.GetHeight(), false)
	if client.IsNoDataErr(err) {
		// no data, return sequence 0
		return 0, 0, nil
	}
	return
}

func (n *StakeRPCService) postTx(tx sdk.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	packet := wire.BinaryBytes(tx)
	result := new(ctypes.ResultBroadcastTxCommit)
	_, err := n.backend.client.Call("broadcast_tx_commit",
		map[string]interface{}{"tx": packet}, result)

	if err != nil {
		return nil, err
	}
	return result, nil
}
