package ethereum

import (
	"fmt"

	"math/big"

	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/utils"
	schedule "github.com/CyberMiles/travis/vm/ethereum/schedule_tx"
	"github.com/ethereum/go-ethereum/common"
	um "github.com/ethereum/go-ethereum/core/vm/umbrella"
)

type EthUmbrella struct {
	DeliveringTxHash *common.Hash
	lastTxHash       *common.Hash
	hashIndex        int
}

func (eu *EthUmbrella) GetValidators() []common.Address {
	validators := stake.GetCandidates().Validators()
	if validators == nil || validators.Len() == 0 {
		return nil
	}
	vs := []common.Address{}
	for i, _ := range validators {
		vs = append(vs, common.HexToAddress(validators[i].OwnerAddress))
	}
	return vs
}

func (eu *EthUmbrella) EmitScheduleTx(stx um.ScheduleTx) {
	if eu.DeliveringTxHash != nil {
		if eu.lastTxHash != nil && eu.DeliveringTxHash.String() == eu.lastTxHash.String() {
			eu.hashIndex += 1
		} else {
			eu.lastTxHash = eu.DeliveringTxHash
			eu.hashIndex = 1
		}
		schedule.SaveScheduleTx(fmt.Sprintf("%s_%d", eu.DeliveringTxHash.String(), eu.hashIndex), &stx)
	}
}

func (eu *EthUmbrella) GetDueTxs() []um.ScheduleTx {
	return nil
}

func (eu *EthUmbrella) DefaultGasPrice() *big.Int {
	return new(big.Int).SetUint64(utils.GetParams().GasPrice)
}

func (eu *EthUmbrella) FreeGasLimit() *big.Int {
	return new(big.Int).SetUint64(utils.GetParams().LowPriceTxGasLimit)
}
