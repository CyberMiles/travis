package governance

import (
	"github.com/CyberMiles/travis/utils"
	"github.com/ethereum/go-ethereum/common"
	"encoding/json"
	"golang.org/x/crypto/ripemd160"
)

type Proposal struct {
	Id           string
	Type         string
	Proposer     *common.Address
	BlockHeight  uint64
	ExpireBlockHeight uint64
	CreatedAt    string
	Result       string
	ResultMsg    string
	ResultBlockHeight    uint64
	ResultAt     string
	Detail       map[string]interface{}
}

func (p *Proposal) Hash() []byte {
	pp, err := json.Marshal(struct {
		Id           string
		Type         string
		Proposer     *common.Address
		BlockHeight  uint64
		ExpireBlockHeight uint64
		Result       string
		ResultMsg    string
		ResultBlockHeight    uint64
		Detail       map[string]interface{}
	}{
		p.Id,
		p.Type,
		p.Proposer,
		p.BlockHeight,
		p.ExpireBlockHeight,
		p.Result,
		p.ResultMsg,
		p.ResultBlockHeight,
		p.Detail,
	})
	if err != nil {
		panic(err)
	}
	hasher := ripemd160.New()
	hasher.Write(pp)
	return hasher.Sum(nil)
}

func NewProposal(id string, ptype string, proposer *common.Address, blockHeight uint64, from *common.Address, to *common.Address, amount string, reason string, expireBlockHeight uint64) *Proposal {
	now := utils.GetNow()
	return &Proposal {
		id,
		ptype,
		proposer,
		blockHeight,
		expireBlockHeight,
		now,
		"",
		"",
		0,
		"",
		map[string]interface{}{
			"from": from,
			"to": to,
			"amount": amount,
			"reason": reason,
		},
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

