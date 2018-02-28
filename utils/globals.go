package utils

import "github.com/tendermint/go-wire/data"

type StateChangeObject struct {
	From data.Bytes
	To data.Bytes
	Amount int64
}

var(
	StateChangeQueue []StateChangeObject
	ValidatorPubKeys [][]byte
)