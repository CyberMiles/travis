package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type StateChangeObject struct {
	From   common.Address
	To     common.Address
	Amount *big.Int

	Reactor StateChangeReactor
}

type StateChangeReactor interface {
	React(result, msg string)
}

var (
	BlockGasFee      *big.Int
	StateChangeQueue []StateChangeObject
	NonceCheckedTx   map[common.Hash]bool = make(map[common.Hash]bool)
	MintAccount                           = common.HexToAddress("0000000000000000000000000000000000000000")
	HoldAccount                           = common.HexToAddress("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
)
