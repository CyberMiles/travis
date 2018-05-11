package utils

import (
	"encoding/hex"
	"fmt"
	"github.com/tendermint/go-crypto"
)

func RemoveFromSlice(slice []interface{}, i int) []interface{} {
	copy(slice[i:], slice[i+1:])
	return slice[:len(slice)-1]
}

func GetPubKey(pubKeyStr string) (pk crypto.PubKey, err error) {

	if len(pubKeyStr) == 0 {
		err = fmt.Errorf("must use --pubkey flag")
		return
	}
	if len(pubKeyStr) != 64 { //if len(pkBytes) != 32 {
		err = fmt.Errorf("pubkey must be Ed25519 hex encoded string which is 64 characters long")
		return
	}
	var pkBytes []byte
	pkBytes, err = hex.DecodeString(pubKeyStr)
	if err != nil {
		return
	}
	var pkEd crypto.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	pk = pkEd
	return
}

func PubKeyString(pk crypto.PubKey) string {
	switch pki := pk.(type) {
	case crypto.PubKeyEd25519:
		return fmt.Sprintf("%X", pki[:])
	case crypto.PubKeySecp256k1:
		return fmt.Sprintf("%X", pki[:])
	default:
		return ""
	}
}
