package keys

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/common"
)

// Signable represents any transaction we wish to send to tendermint core
// These methods allow us to sign arbitrary Tx with the KeyStore
type Signable interface {
	// SignBytes is the immutable data, which needs to be signed
	SignBytes() []byte

	// Sign will add a transaction with signature
	Sign(tx *types.Transaction) error

	// Signers will return the public key(s) that signed if the signature
	// is valid, or an error if there is any issue with the signature,
	// including if there are no signatures
	Signers() (common.Address, error)

	// TxBytes returns the transaction data as well as all signatures
	// It should return an error if Sign was never called
	TxBytes() ([]byte, error)
}