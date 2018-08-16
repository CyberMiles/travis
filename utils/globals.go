package utils

import (
	"math"
	"math/big"

	"github.com/CyberMiles/travis/sdk"
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
	Amount sdk.Int

	Reactor StateChangeReactor
}

type StateChangeReactor interface {
	React(result, msg string)
}

type pendingProposal struct {
	proposalsTS        map[string]int64
	minExpireTimestamp int64
	minTSMappedPid     []string

	proposalsBH          map[string]int64
	minExpireBlockHeight int64
	minBHMappedPid       []string
}

func (p *pendingProposal) BatchAddTS(proposals map[string]int64) {
	p.proposalsTS = proposals
	p.updateTS()
}

func (p *pendingProposal) BatchAddBH(proposals map[string]int64) {
	p.proposalsBH = proposals
	p.updateBH()
}

func (p *pendingProposal) Add(pid string, expireTimestamp, expireBlockHeight int64) {
	if expireTimestamp > 0 {
		p.proposalsTS[pid] = expireTimestamp
		if p.minExpireTimestamp > expireTimestamp {
			p.minTSMappedPid = []string{pid}
			p.minExpireTimestamp = expireTimestamp
		} else if p.minExpireTimestamp == expireTimestamp {
			p.minTSMappedPid = append(p.minTSMappedPid, pid)
		}
	} else if expireBlockHeight > 0 {
		p.proposalsBH[pid] = expireBlockHeight
		if p.minExpireBlockHeight > expireBlockHeight {
			p.minBHMappedPid = []string{pid}
			p.minExpireBlockHeight = expireBlockHeight
		} else if p.minExpireBlockHeight == expireBlockHeight {
			p.minBHMappedPid = append(p.minBHMappedPid, pid)
		}
	}
}

func (p *pendingProposal) Del(pid string) {
	if expireTimestamp, ok := p.proposalsTS[pid]; ok {
		delete(p.proposalsTS, pid)
		if p.minExpireTimestamp == expireTimestamp {
			if len(p.minTSMappedPid) == 1 {
				p.updateTS()
			} else {
				for idx, id := range p.minTSMappedPid {
					if id == pid {
						p.minTSMappedPid = append(p.minTSMappedPid[:idx], p.minTSMappedPid[idx+1:]...)
						break
					}
				}
			}
		}
	} else if expireBlockHeight, ok := p.proposalsBH[pid]; ok {
		delete(p.proposalsBH, pid)
		if p.minExpireBlockHeight == expireBlockHeight {
			if len(p.minBHMappedPid) == 1 {
				p.updateBH()
			} else {
				for idx, id := range p.minBHMappedPid {
					if id == pid {
						p.minBHMappedPid = append(p.minBHMappedPid[:idx], p.minBHMappedPid[idx+1:]...)
						break
					}
				}
			}
		}
	}
}

func (p *pendingProposal) updateTS() {
	min := int64(math.MaxInt64)

	for pid, ts := range p.proposalsTS {
		if min > ts {
			min = ts
			p.minTSMappedPid = []string{pid}
		} else if min == ts {
			p.minTSMappedPid = append(p.minTSMappedPid, pid)
		}
	}
	p.minExpireTimestamp = min
}

func (p *pendingProposal) updateBH() {
	min := int64(math.MaxInt64)

	for pid, bh := range p.proposalsBH {
		if min > bh {
			min = bh
			p.minBHMappedPid = []string{pid}
		} else if min == bh {
			p.minBHMappedPid = append(p.minBHMappedPid, pid)
		}
	}
	p.minExpireBlockHeight = min
}

func (p *pendingProposal) ReachMin(timestamp, blockHeight int64) (pids []string) {
	if shouldBePacked(p.minExpireTimestamp, timestamp) {
		pids = p.minTSMappedPid

		for _, pid := range pids {
			delete(p.proposalsTS, pid)
		}

		for pid, ts := range p.proposalsTS {
			if shouldBePacked(ts, timestamp) {
				delete(p.proposalsTS, pid)
				pids = append(pids, pid)
			}
		}

		p.updateTS()
	}

	if p.minExpireBlockHeight <= blockHeight {
		pids = append(pids, p.minBHMappedPid...)

		for _, pid := range p.minBHMappedPid {
			delete(p.proposalsBH, pid)
		}

		p.updateBH()
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

func CalGasFee(gasUsed uint64, gasPrice uint64) sdk.Int {
	return sdk.NewInt(int64(gasUsed)).Mul(sdk.NewInt(int64(gasPrice)))
}

var (
	BlockGasFee    = big.NewInt(0)
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
		make(map[string]int64),
		math.MaxInt64,
		nil,
	}
	MintAccount    = common.HexToAddress("0000000000000000000000000000000000000000")
	HoldAccount    = common.HexToAddress("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	GovHoldAccount = common.HexToAddress("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
)
