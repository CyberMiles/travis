package modules

import (
	"math/big"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
)

func Transfer(from, to common.Address, amount *big.Int) error {
	utils.StateChangeQueue = append(utils.StateChangeQueue, utils.StateChangeObject{
		From: from, To: to, Amount: amount})
	return nil
}

func CheckAccount(from common.Address, amount *big.Int) error {
	// todo check to see if balance of sender's account is enough to transfer
	return nil
}