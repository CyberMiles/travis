package types

import "github.com/tendermint/go-crypto"

// GenesisValidator is an initial validator.
type GenesisValidator struct {
	PubKey    crypto.PubKey `json:"pub_key"`
	Power     int64         `json:"power"`
	Name      string        `json:"name"`
	Address   string        `json:"address"`
	Cut       int64         `json:"cut"`
	MaxAmount int64         `json:"max_amount"`
}

