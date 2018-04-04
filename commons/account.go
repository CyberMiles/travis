package commons

import (
	"math/big"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"path/filepath"
	"time"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"math"
)

const (
	// Ethereum default keystore directory
	datadirDefaultKeyStore = "keystore"
)

var (
	emHome = os.ExpandEnv("$HOME/.travis")
)

func MakeAccountManager() (*accounts.Manager, string, error) {
	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP
	keydir := filepath.Join(emHome, datadirDefaultKeyStore)

	ephemeral := keydir
	if err := os.MkdirAll(keydir, 0700); err != nil {
		return nil, "", err
	}
	// Assemble the account manager and supported backends
	backends := []accounts.Backend{
		keystore.NewKeyStore(keydir, scryptN, scryptP),
	}

	return accounts.NewManager(backends...), ephemeral, nil
}

// fetchKeystore retrives the encrypted keystore from the account manager.
func fetchKeystore(am *accounts.Manager) *keystore.KeyStore {
	return am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}

func UnlockAccount(am *accounts.Manager, addr common.Address, password string, duration *uint64) (bool, error) {
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return false, fmt.Errorf("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	err := fetchKeystore(am).TimedUnlock(accounts.Account{Address: addr}, password, d)
	return err == nil, err
}

func Transfer(from, to common.Address, amount *big.Int) error {
	utils.StateChangeQueue = append(utils.StateChangeQueue, utils.StateChangeObject{
		From: from, To: to, Amount: amount})
	return nil
}

func CheckAccount(from common.Address, amount *big.Int) error {
	// todo check to see if balance of sender's account is enough to transfer
	return nil
}