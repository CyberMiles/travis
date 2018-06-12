package utils

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type StateChangeObject struct {
	From   common.Address
	To     common.Address
	Amount *big.Int

	Reactor StateChangeReactor
}

type StateChangeReactor interface {
	React(result, msg string)
}

type pendingProposal struct {
	proposals            map[string]uint64
	minExpireBlockHeight uint64
	minHeightMappedPid   []string
}

func (p *pendingProposal) BatchAdd(proposals map[string]uint64) {
	p.proposals = proposals
	p.update()
}

func (p *pendingProposal) Add(pid string, expireBlockHeight uint64) {
	p.proposals[pid] = expireBlockHeight
	if p.minExpireBlockHeight > expireBlockHeight {
		p.minHeightMappedPid = []string{pid}
		p.minExpireBlockHeight = expireBlockHeight
	} else if p.minExpireBlockHeight == expireBlockHeight {
		p.minHeightMappedPid = append(p.minHeightMappedPid, pid)
	}
}

func (p *pendingProposal) Del(pid string) {
	expireBlockHeight := p.proposals[pid]
	delete(p.proposals, pid)
	if p.minExpireBlockHeight == expireBlockHeight {
		if len(p.minHeightMappedPid) == 1 {
			p.update()
		} else {
			for idx, id := range p.minHeightMappedPid {
				if id == pid {
					p.minHeightMappedPid = append(p.minHeightMappedPid[:idx], p.minHeightMappedPid[idx+1:]...)
					break
				}
			}
		}
	}
}

func (p *pendingProposal) update() {
	min := uint64(math.MaxUint64)

	for pid, bh := range p.proposals {
		if min > bh {
			min = bh
			p.minHeightMappedPid = []string{pid}
		} else if min == bh {
			p.minHeightMappedPid = append(p.minHeightMappedPid, pid)
		}
	}
	p.minExpireBlockHeight = min
}

func (p *pendingProposal) ReachMin(blockHeight uint64) (pids []string) {
	if p.minExpireBlockHeight == blockHeight {
		pids = p.minHeightMappedPid

		for _, pid := range pids {
			delete(p.proposals, pid)
		}
		p.update()
	}
	return
}

func IsEthTx(tx *types.Transaction) bool {
	zero := big.NewInt(0)
	return tx.Data() == nil ||
		tx.GasPrice().Cmp(zero) != 0 ||
		tx.Gas().Cmp(zero) != 0 ||
		tx.Value().Cmp(zero) != 0 ||
		tx.To() != nil
}

var (
	BlockGasFee      *big.Int
	StateChangeQueue []StateChangeObject
	// Recording addresses associated with travis tx (stake/governance) in one block
	// Transfer transaction is not allowed if the sender of which was found in this recording
	// TODO to be removed
	TravisTxAddrs   []*common.Address
	NonceCheckedTx  map[common.Hash]bool = make(map[common.Hash]bool)
	PendingProposal                      = &pendingProposal{
		make(map[string]uint64),
		math.MaxUint64,
		nil,
	}
	MintAccount    = common.HexToAddress("0000000000000000000000000000000000000000")
	HoldAccount    = common.HexToAddress("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	GovHoldAccount = common.HexToAddress("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
)
