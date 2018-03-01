package utils

import (
	"github.com/tendermint/go-wire/data"
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

)