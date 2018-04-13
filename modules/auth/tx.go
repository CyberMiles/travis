/*
Package auth contains generic Signable implementations that can be used
by your application or tests to handle authentication needs.

It currently supports transaction data as opaque bytes and either single
or multiple private key signatures using straightforward algorithms.
It currently does not support N-of-M key share signing of other more
complex algorithms (although it would be great to add them).

You can create them with NewSig() and NewMultiSig(), and they fulfill
the keys.Signable interface. You can then .Wrap() them to create
a sdk.Tx.
*/
package auth

import (
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tendermint/go-wire/data"

	commons "github.com/CyberMiles/travis/commons"
	ttypes "github.com/CyberMiles/travis/types"
)

// nolint
const (
	// for signatures
	ByteSingleTx = 0x18
)

// nolint
const (
	// for signatures
	TypeSingleTx = "sigs/one"
)

// Signed holds one signature of the data
type Signed struct {
	tx *types.Transaction
}

/**** Registration ****/

func init() {
	sdk.TxMapper.
		RegisterImplementation(&OneSig{}, TypeSingleTx, ByteSingleTx)
}

/**** One Sig ****/

// OneSig lets us wrap arbitrary data with a go-crypto signature
type OneSig struct {
	Tx       sdk.Tx `json:"tx"`
	SignedTx []byte `json:"signature"`
}

var _ ttypes.Signable = &OneSig{}
var _ sdk.TxLayer = &OneSig{}

// NewSig wraps the tx with a Signable that accepts exactly one signature
func NewSig(tx sdk.Tx) *OneSig {
	return &OneSig{Tx: tx}
}

//nolint
func (s *OneSig) Wrap() sdk.Tx {
	return sdk.Tx{s}
}
func (s *OneSig) Next() sdk.Tx {
	return s.Tx
}
func (s *OneSig) ValidateBasic() error {
	return s.Tx.ValidateBasic()
}

// TxBytes returns the full data with signatures
func (s *OneSig) TxBytes() ([]byte, error) {
	return data.ToWire(s.Wrap())
}

// SignBytes returns the original data passed into `NewSig`
func (s *OneSig) SignBytes() []byte {
	res, err := data.ToWire(s.Tx)
	if err != nil {
		panic(err)
	}
	return res
}

// Sign will add transaction with signature
func (s *OneSig) Sign(tx *types.Transaction) error {
	// set the value once we are happy
	encodedTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		fmt.Errorf("Error encoding the transaction: %v", err)
	}

	s.SignedTx = encodedTx
	return nil
}

func (s *OneSig) Signers() (common.Address, error) {
	ntx := new(types.Transaction)
	rlp.DecodeBytes(s.SignedTx, ntx)

	var signer types.Signer = types.NewEIP155Signer(ntx.ChainId())

	// Make sure the transaction is signed properly
	from, err := types.Sender(signer, ntx)
	if err != nil {
		return common.Address{}, ErrInvalidSignature()
	}

	return from, nil
}

// Sign - sign the transaction with private key
func Sign(tx ttypes.Signable, address string, passphrase string) error {
	ethTx := types.NewTransaction(
		0,
		common.Address([20]byte{}),
		big.NewInt(0),
		big.NewInt(0),
		big.NewInt(0),
		tx.SignBytes(),
	)

	am, _, _ := commons.MakeAccountManager()
	addr := common.HexToAddress(address)
	_, err := commons.UnlockAccount(am, addr, passphrase, nil)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	account := accounts.Account{Address: addr}
	wallet, err := am.Find(account)
	signed, err := wallet.SignTx(account, ethTx, big.NewInt(15))
	if err != nil {
		fmt.Errorf("error")
	}

	return tx.Sign(signed)
}
