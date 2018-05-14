package utils

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-crypto"
)

var Cdc = amino.NewCodec()

func init() {
	crypto.RegisterAmino(Cdc)
}
