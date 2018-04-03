/*
Package nonce - This module allows replay protection to be added to process stack.
This is achieved through the use of a sequence number for each unique set of signers.
Note that the sequence number for the single signing account "foo" will be unique
from the sequence number for a multi-sig account {"foo", "bar"} which would also be
unique from a different multi-sig account {"foo", "soup"}
*/
package nonce

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
	"github.com/CyberMiles/travis/types"
	"github.com/ethereum/go-ethereum/common"
)

// nolint
const (
	ByteNonce = 0x69 //TODO overhaul byte assign system don't make no sense!
	TypeNonce = "nonce"
)

func init() {
	sdk.TxMapper.RegisterImplementation(Tx{}, TypeNonce, ByteNonce)
}

// Tx - Nonce transaction structure, contains list of signers and current sequence number
type Tx struct {
	Sequence uint32      		`json:"sequence"`
	Signers  []common.Address 	`json:"signers"`
	Tx       sdk.Tx      		`json:"tx"`
}

var _ sdk.TxInner = &Tx{}

// NewTx wraps the tx with a signable nonce
func NewTx(sequence uint32, signers []common.Address, tx sdk.Tx) sdk.Tx {
	return (Tx{
		Sequence: sequence,
		Signers:  signers,
		Tx:       tx,
	}).Wrap()
}

//nolint
func (n Tx) Wrap() sdk.Tx {
	return sdk.Tx{n}
}
func (n Tx) ValidateBasic() error {
	switch {
	case n.Tx.Empty():
		return ErrTxEmpty()
	case n.Sequence == 0:
		return ErrZeroSequence()
	case len(n.Signers) == 0:
		return ErrNoSigners()
	}
	return n.Tx.ValidateBasic()
}
func (n Tx) Next() sdk.Tx {
	return n.Tx
}

// CheckIncrementSeq - Check that the sequence number is one more than the state sequence number
// and further increment the sequence number
// NOTE It is okay to modify the sequence before running the wrapped TX because if the
// wrapped Tx fails, the state changes are not applied
func (n Tx) CheckIncrementSeq(ctx types.Context, store state.SimpleDB) error {

	seqKey := n.getSeqKey()

	// check the current state
	cur, err := getSeq(store, seqKey)
	if err != nil {
		return err
	}
	if n.Sequence != cur+1 {
		return ErrBadNonce(n.Sequence, cur+1)
	}

	// make sure they all signed
	for _, s := range n.Signers {
		if !ctx.HasSigner(s) {
			return ErrNotMember()
		}
	}

	// increment the sequence by 1
	err = setSeq(store, seqKey, cur+1)
	if err != nil {
		return err
	}

	return nil
}

func (n Tx) getSeqKey() (seqKey []byte) {
	return GetSeqKey(n.Signers)
}

// GetSeqKey - Generate the sequence key as the concatenated list of signers, sorted by address.
func GetSeqKey(signers []common.Address) (seqKey []byte) {

	// First copy the list of signers to sort as sort is done in place
	signers2sort := make([]common.Address, len(signers))
	copy(signers2sort, signers)
	sort.Sort(types.ByAll(signers))

	for _, signer := range signers {
		seqKey = append(seqKey, signer.Bytes()...)
	}

	return
}
