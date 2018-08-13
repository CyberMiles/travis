package stake

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	travis "github.com/CyberMiles/travis/types"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
)

var bigZero = big.NewInt(0)
var bigOne = big.NewInt(1)

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

	m := new(big.Int)
	bs, err := hex.DecodeString(sig)
	if err != nil {
		fmt.Printf("decode signature error: %v\n", err)
		return err
	}

	m.SetBytes(bs)
	c := encrypt(new(big.Int), pk, m)

	if !bytes.Equal(c.Bytes(), hashed[:]) {
		return ErrInvalidCubeSignature()
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

func modInverse(a, n *big.Int) (ia *big.Int, ok bool) {
	g := new(big.Int)
	x := new(big.Int)
	g.GCD(x, nil, a, n)
	if g.Cmp(bigOne) != 0 {
		// In this case, a and n aren't coprime and we cannot calculate
		// the inverse. This happens because the values of n are nearly
		// prime (being the product of two primes) rather than truly
		// prime.
		return
	}

	if x.Cmp(bigOne) < 0 {
		// 0 is not the multiplicative inverse of any element so, if x
		// < 1, then x is negative.
		x.Add(x, n)
	}

	return x, true
}

func digest(message string) []byte {
	sha_256 := sha256.New()
	sha_256.Write(bytes.NewBufferString(message).Bytes())
	return sha_256.Sum(nil)
}

func encrypt(c *big.Int, pub *rsa.PublicKey, m *big.Int) *big.Int {
	e := big.NewInt(int64(pub.E))
	c.Exp(m, e, pub.N)
	return c
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
