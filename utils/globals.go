package utils

import (
	"github.com/tendermint/go-wire/data"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type StateChangeObject struct {
	From data.Bytes
	To data.Bytes
	Amount *big.Int
}

var(
	BlockGasFee *big.Int
	StateChangeQueue []StateChangeObject
	ValidatorPubKeys [][]byte

	// record count of failed CheckTx of each from account; used to feed in the nonce check
	CheckFailedCount map[common.Address]uint64 = make(map[common.Address]uint64)
)