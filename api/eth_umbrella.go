package api

import (
	"github.com/CyberMiles/travis/modules/stake"
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
