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

	NonceCheckedTx map[common.Hash]bool = make(map[common.Hash]bool)
)