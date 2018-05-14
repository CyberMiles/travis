package utils

import (
	"fmt"
	"github.com/tendermint/go-crypto"
	"encoding/json"
)

func RemoveFromSlice(slice []interface{}, i int) []interface{} {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}

type JsonPubKey struct {
	Pt string `json:"type"`
	Pv string `json:"value"`
}

func GetPubKey(pubKeyStr string) (pk crypto.PubKey, err error) {

	if len(pubKeyStr) == 0 {
		err = fmt.Errorf("must use --pubkey flag")
		return
	}
	jpk := JsonPubKey{
		"AC26791624DE60",
		pubKeyStr,
	}
	b, err := json.Marshal(jpk)
	err = Cdc.UnmarshalJSON(b, &pk)
	return
}

func PubKeyString(pk crypto.PubKey) string {
	b, _ := Cdc.MarshalJSON(pk)
	var jpk JsonPubKey
	json.Unmarshal(b, &jpk)
	return jpk.Pv
}
