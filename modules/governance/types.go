package governance

import (
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"encoding/json"
	"golang.org/x/crypto/ripemd160"
)

type Proposal struct {
	Id           string
	Proposer     *common.Address
	BlockHeight  uint64
	From         *common.Address
	To           *common.Address
	Amount       string
	Reason       string
	ExpireBlockHeight uint64
	CreatedAt    string
	Result       string
	ResultMsg    string
	ResultBlockHeight    uint64
	ResultAt     string
}

func (p *Proposal) Hash() []byte {
	pp, err := json.Marshal(struct {
		Id           string
		Proposer     *common.Address
		BlockHeight  uint64
		From         *common.Address
		To           *common.Address
		Amount       string
		Reason       string
		ExpireBlockHeight uint64
		Result       string
		ResultMsg    string
		ResultBlockHeight    uint64
	}{
		p.Id,
		p.Proposer,
		p.BlockHeight,
		p.From,
		p.To,
		p.Amount,
		p.Reason,
		p.ExpireBlockHeight,
		p.Result,
		p.ResultMsg,
		p.ResultBlockHeight,
	})
	if err != nil {
		panic(err)
	}
	hasher := ripemd160.New()
	hasher.Write(pp)
	return hasher.Sum(nil)
}

func NewProposal(id string, proposer *common.Address, blockHeight uint64, from *common.Address, to *common.Address, amount string, reason string, expireBlockHeight uint64) *Proposal {
	now := utils.GetNow()
	return &Proposal {
		id,
		proposer,
		blockHeight,
		from,
		to,
		amount,
		reason,
		expireBlockHeight,
		now,
		"",
		"",
		0,
		"",
	}
}

type Vote struct {
	ProposalId     string
	Voter          common.Address
	BlockHeight    uint64
	Answer         string
	CreatedAt      string
}

func (v *Vote) Hash() []byte {
	vote, err := json.Marshal(struct {
		ProposalId     string
		Voter          common.Address
		BlockHeight    uint64
		Answer         string
	}{
		v.ProposalId,
		v.Voter,
		v.BlockHeight,
		v.Answer,
	})
	if err != nil {
		panic(err)
	}
	hasher := ripemd160.New()
	hasher.Write(vote)
	return hasher.Sum(nil)
}

func NewVote(proposalId string, voter common.Address, blockHeight uint64, answer string) *Vote {
	now := utils.GetNow()
	return &Vote {
		proposalId,
		voter,
		blockHeight,
		answer,
		now,
	}
}

