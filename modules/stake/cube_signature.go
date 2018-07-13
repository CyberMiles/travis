package stake

import (
	"fmt"
	"strings"
	"crypto/sha256"
	"crypto/rsa"
	"crypto"
	"encoding/pem"
	"crypto/x509"
	"github.com/ethereum/go-ethereum/common"
	"github.com/CyberMiles/travis/utils"
	"encoding/json"
	travis "github.com/CyberMiles/travis/types"
)

func VerifyCubeSignature(address common.Address, nonce uint64, cubeBatch string, sig string) error {
	message := fmt.Sprintf("%s|%d", strings.ToLower(address.String()), nonce)
	//fmt.Printf("message: %s\n", message)
	hashed := sha256.Sum256([]byte(message))
	publicKeyStr, err := getCubePublicKeyString(cubeBatch)
	if err != nil {
		fmt.Printf("Error occurred while retrieving public key string: %s\n", err)
		return err
	}

	pk, err := loadPublicKey(publicKeyStr)
	if err != nil {
		fmt.Printf("Error occurred while loading public key: %s\n", err)
		return err
	}

	err = rsa.VerifyPKCS1v15(pk, crypto.SHA256, hashed[:], common.Hex2Bytes(sig))
	if err != nil {
		fmt.Printf("Error occurred while verifying the signature: %s\n", err)
		return err
	}

	return nil
}

func getCubePublicKeyString(cubeBatch string) (string, error) {
	pksBytes := utils.GetParams().CubePubKeys
	var pks []travis.GenesisCubePubKey
	err := json.Unmarshal([]byte(pksBytes), &pks)
	if err != nil {
		return "", err
	}

	for _, pk := range pks {
		if pk.CubeBatch == cubeBatch {
			return pk.PubKey, nil
		}
	}

	return "", nil
}

func loadPublicKey(publicKeyStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyStr))
	if block == nil {
		return nil, fmt.Errorf("public key error")
	}

	pkInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pk := pkInterface.(*rsa.PublicKey)
	return pk, nil
}
