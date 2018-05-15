package types

import (
	"github.com/tendermint/go-amino"
	"github.com/tendermint/go-crypto"
	"fmt"
	"encoding/json"
)

var cdc = amino.NewCodec()

func init() {
	crypto.RegisterAmino(cdc)
}

type jsonPubKey struct {
	Pt string `json:"type"`
	Pv string `json:"value"`
}

func GetPubKey(pubKeyStr string) (pk PubKey, err error) {

	if len(pubKeyStr) == 0 {
		err = fmt.Errorf("must use --pubkey flag")
		return
	}
	jpk := jsonPubKey{
		"AC26791624DE60",
		pubKeyStr,
	}
	b, err := json.Marshal(jpk)

	var cpk crypto.PubKey
	err = cdc.UnmarshalJSON(b, &cpk)

	pk = PubKey{cpk}
	return
}

func PubKeyString(pk PubKey) string {
	b, err := cdc.MarshalJSON(pk.PubKey)
	if err != nil {
		return ""
	}
	var jpk jsonPubKey
	json.Unmarshal(b, &jpk)
	return jpk.Pv
}

type PubKey struct {
	crypto.PubKey
}

func (pk *PubKey) MarshalJSON() ([]byte, error) {
	return cdc.MarshalJSON(pk.PubKey)
}

func (pk *PubKey) UnmarshalJSON(b []byte) error {
	err := cdc.UnmarshalJSON(b, &pk.PubKey)
	return err
}
