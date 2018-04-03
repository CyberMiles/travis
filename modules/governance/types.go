package governance

import (
	"math/big"

	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
)

type Proposal struct {
	Id           string
	Proposer     common.Address
	BlockHeight  uint64
	From         common.Address
	To           common.Address
	Amount       *big.Int
	Reason       string
	CreatedAt    string
}

func NewProposal(id string, proposer common.Address, blockHeight uint64, from common.Address, to common.Address, amount *big.Int, reason string) *Proposal {
	now := utils.GetNow()
	return &Proposal {
		Id:          id,
		Proposer:    proposer,
		BlockHeight: blockHeight,
		From:        from,
		To:          to,
		Amount:      amount,
		Reason:      reason,
		CreatedAt:    now,
	}
}
