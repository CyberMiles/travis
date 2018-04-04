package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type StateChangeObject struct {
	From common.Address
	To common.Address
	Amount *big.Int
}

var(
	BlockGasFee *big.Int
	StateChangeQueue []StateChangeObject
	ValidatorPubKeys [][]byte

	NonceCheckedTx map[common.Hash]bool = make(map[common.Hash]bool)
)