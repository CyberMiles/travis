package stake

import (
	crypto "github.com/tendermint/go-crypto"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

/**** code to parse accounts from genesis docs ***/

// GenesisValidator - genesis validator parameters
type genesisValidator struct {
	Address common.Address 	`json:"address"`
	PubKey  crypto.PubKey 	`json:"pub_key"`
	Power 	*big.Int        `json:"power"`
	Name    string          `json:"name"`
}
