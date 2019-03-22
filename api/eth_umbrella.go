package api

import (
	"math/big"
	"github.com/CyberMiles/travis/modules/stake"
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	um "github.com/ethereum/go-ethereum/core/vm/umbrella"
)

type EthUmbrella struct {
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
