package ethereum

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

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
func (s *NetRPCService) Listening() bool {
	return true // always listening
}

// PeerCount returns the number of connected peers
// #unstable
func (s *NetRPCService) PeerCount() hexutil.Uint {
	return hexutil.Uint(0)
}

// Version returns the current ethereum protocol version.
// #unstable
func (s *NetRPCService) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}

// StakeRPCService offers stake related RPC methods
type StakeRPCService struct {
	backend *Backend
	am      *accounts.Manager
	chainID string
}

// NewStakeRPCAPI create a new StakeRPCAPI.
func NewStakeRPCService(b *Backend) *StakeRPCService {
	return &StakeRPCService{
		backend: b,
		am:      b.ethereum.AccountManager(),
		chainID: "",
	}
}

func (s *StakeRPCService) getChainID() (string, error) {
	if s.chainID == "" {
		if s.backend.tmNode == nil {
			return "", errors.New("waiting for tendermint to finish starting up")
		}
		s.chainID = s.backend.tmNode.GenesisDoc().ChainID
		return s.chainID, nil
	}

	return s.chainID, nil
}

// copied from ethapi/api.go
func (s *StakeRPCService) UnlockAccount(addr common.Address, password string, duration *uint64) (bool, error) {
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return false, errors.New("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	err := fetchKeystore(s.am).TimedUnlock(accounts.Account{Address: addr}, password, d)
	return err == nil, err
}

// fetchKeystore retrives the encrypted keystore from the account manager.
func fetchKeystore(am *accounts.Manager) *keystore.KeyStore {
	return am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}

type DeclareCandidacyArgs struct {
	Sequence    uint32            `json:"sequence"`
	From        string            `json:"from"`
	PubKey      string            `json:"pubKey"`
	Bond        coin.Coin         `json:"bond"`
	Description stake.Description `json:"description"`
}

func (s *StakeRPCService) DeclareCandidacy(di DeclareCandidacyArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareDeclareCandidacyTx(di)
	if err != nil {
		return nil, err
	}
	return s.postTx(tx)
}

func (s *StakeRPCService) prepareDeclareCandidacyTx(di DeclareCandidacyArgs) (sdk.Tx, error) {
	pubKey, err := stake.GetPubKey(di.PubKey)
	if err != nil {
		return sdk.Tx{}, err
	}
	tx := stake.NewTxDeclareCandidacy(di.Bond, pubKey, di.Description)
	return s.wrapAndSignTx(tx, di.From, di.Sequence)
}

type DelegateArgs struct {
	Sequence uint32    `json:"sequence"`
	From     string    `json:"from"`
	PubKey   string    `json:"pubKey"`
	Bond     coin.Coin `json:"bond"`
}

func (s *StakeRPCService) Delegate(di DelegateArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareDelegateTx(di)
	if err != nil {
		return nil, err
	}
	return s.postTx(tx)
}

func (s *StakeRPCService) prepareDelegateTx(di DelegateArgs) (sdk.Tx, error) {
	pubKey, err := stake.GetPubKey(di.PubKey)
	if err != nil {
		return sdk.Tx{}, err
	}
	tx := stake.NewTxDelegate(di.Bond, pubKey)
	return s.wrapAndSignTx(tx, di.From, di.Sequence)
}

type UnbondArgs struct {
	Sequence uint32 `json:"sequence"`
	From     string `json:"from"`
	PubKey   string `json:"pubKey"`
	Amount   uint64 `json:"amount"`
}

func (s *StakeRPCService) Unbond(di UnbondArgs) (*ctypes.ResultBroadcastTxCommit, error) {
	tx, err := s.prepareUnbondTx(di)
	if err != nil {
		return nil, err
	}
	return s.postTx(tx)
}

func (s *StakeRPCService) prepareUnbondTx(di UnbondArgs) (sdk.Tx, error) {
	pubKey, err := stake.GetPubKey(di.PubKey)
	if err != nil {
		return sdk.Tx{}, err
	}
	tx := stake.NewTxUnbond(di.Amount, pubKey)
	return s.wrapAndSignTx(tx, di.From, di.Sequence)
}

func (s *StakeRPCService) wrapAndSignTx(tx sdk.Tx, address string, sequence uint32) (sdk.Tx, error) {
	// only add the actual signer to the nonce
	signers := []sdk.Actor{getSignerAct(address)}
	tx = nonce.NewTx(sequence, signers, tx)
	chainID, err := s.getChainID()
	if err != nil {
		return sdk.Tx{}, err
	}
	tx = base.NewChainTx(chainID, 0, tx)
	tx = auth.NewSig(tx).Wrap()
	// sign
	err = s.signTx(tx, address)
	if err != nil {
		return sdk.Tx{}, err
	}
	return tx, err
}

func (s *StakeRPCService) signTx(tx sdk.Tx, address string) error {
	// validate tx client-side
	err := tx.ValidateBasic()
	if err != nil {
		return err
	}

	if sign, ok := tx.Unwrap().(keys.Signable); ok {
		if address == "" {
			return errors.New("address is required to sign tx")
		}
		err := s.Sign(sign, address)
		if err != nil {
			return err
		}
	}
	return err
}

// Sign - sign the transaction with private key
func (s *StakeRPCService) Sign(tx keys.Signable, address string) error {
	ethTx := types.NewTransaction(
		0,
		common.Address([20]byte{}),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		tx.SignBytes(),
	)

	addr := common.HexToAddress(address)
	account := accounts.Account{Address: addr}
	wallet, err := s.am.Find(account)
	signed, err := wallet.SignTx(account, ethTx, big.NewInt(15))
	if err != nil {
		return err
	}

	return tx.Sign(signed)
}

func (s *StakeRPCService) postTx(tx sdk.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	packet := wire.BinaryBytes(tx)
	result := new(ctypes.ResultBroadcastTxCommit)
	_, err := s.backend.client.Call("broadcast_tx_commit",
		map[string]interface{}{"tx": packet}, result)

	if err != nil {
		return nil, err
	}
	return result, nil
}

func getSignerAct(address string) (res sdk.Actor) {
	// this could be much cooler with multisig...
	signer := common.HexToAddress(address)
	res = auth.SigPerm(signer.Bytes())
	return res
}
