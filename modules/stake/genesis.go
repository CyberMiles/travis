package stake

import (
	"github.com/tendermint/go-crypto"
	"github.com/ethereum/go-ethereum/common"
)

/**** code to parse accounts from genesis docs ***/

// GenesisValidator - genesis validator parameters
type genesisValidator struct {
	Address common.Address 	`json:"address"`
	PubKey  crypto.PubKey 	`json:"pub_key"`
	Power 	int64        	`json:"power"`
	Name    string          `json:"name"`
}
