package utils

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	CommitSeconds = 10
	BlocksPerHour = 60 * 60 / 10
	BlocksPerDay  = 24 * 60 * 60 / 10
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
	proposals          map[string]int64
	minExpire          int64
	minHeightMappedPid []string
}

func (p *pendingProposal) BatchAdd(proposals map[string]int64) {
	p.proposals = proposals
	p.update()
}

func (p *pendingProposal) Add(pid string, expire int64) {
	p.proposals[pid] = expire
	if p.minExpire > expire {
		p.minHeightMappedPid = []string{pid}
		p.minExpire = expire
	} else if p.minExpire == expire {
		p.minHeightMappedPid = append(p.minHeightMappedPid, pid)
	}
}

func (p *pendingProposal) Del(pid string) {
	expire := p.proposals[pid]
	delete(p.proposals, pid)
	if p.minExpire == expire {
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
	min := int64(math.MaxInt64)

	for pid, ts := range p.proposals {
		if min > ts {
			min = ts
			p.minHeightMappedPid = []string{pid}
		} else if min == ts {
			p.minHeightMappedPid = append(p.minHeightMappedPid, pid)
		}
	}
	p.minExpire = min
}

func (p *pendingProposal) ReachMin(timestamp int64) (pids []string) {
	if shouldBePacked(p.minExpire, timestamp) {
		pids = p.minHeightMappedPid

		for _, pid := range pids {
			delete(p.proposals, pid)
		}

		for pid, ts := range p.proposals {
			if shouldBePacked(ts, timestamp) {
				delete(p.proposals, pid)
				pids = append(pids, pid)
			}
		}

		p.update()
	}
	return
}

func shouldBePacked(timestamp, lastTs int64) bool {
	if timestamp < lastTs || timestamp-lastTs < CommitSeconds {
		return true
	}
	return false
}

func IsEthTx(tx *types.Transaction) bool {
	zero := big.NewInt(0)
	return tx.Data() == nil ||
		tx.GasPrice().Cmp(zero) != 0 ||
		tx.Gas() != 0 ||
		tx.Value().Cmp(zero) != 0 ||
		tx.To() != nil
}

func CalGasFee(gasUsed uint64, gasPrice uint64) *big.Int {
	gasFee := big.NewInt(int64(0))
	gasFee = gasFee.Mul(big.NewInt(int64(gasUsed)), big.NewInt(int64(gasPrice)))
	return gasFee
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
		make(map[string]int64),
		math.MaxInt64,
		nil,
	}
	MintAccount    = common.HexToAddress("0000000000000000000000000000000000000000")
	HoldAccount    = common.HexToAddress("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	GovHoldAccount = common.HexToAddress("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
)
