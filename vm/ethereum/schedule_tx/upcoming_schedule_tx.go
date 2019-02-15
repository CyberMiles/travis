package schedule_tx

import (
	"math"
	"github.com/CyberMiles/travis/utils"
)

type UpcomingScheduleTx struct {
	tsHash map[int64][]string
	minTs int64
}

func (ust *UpcomingScheduleTx) ReloadFromDB() {
	minTs, tsHash := GetUpcomingTsHash()
	ust.minTs = minTs
	ust.tsHash = tsHash
}

func (ust *UpcomingScheduleTx) Add(ts int64, hash string) {
	if ust.minTs == 0 || ts < ust.minTs - utils.CommitSeconds {
		ust.tsHash = make(map[int64][]string)
		ust.tsHash[ts] = []string{hash}
		ust.minTs = ts
	} else if ts < ust.minTs {
		ust.tsHash[ts] = []string{hash}
		ust.minTs = ts
		ust.update(false)
	} else if ts == ust.minTs {
		ha := ust.tsHash[ts]
		ha = append(ha, hash)
	} else if ts < ust.minTs + utils.CommitSeconds {
		repeat := false
		for _ts, _ := range ust.tsHash {
			if ts == _ts {
				ha := ust.tsHash[ts]
				ha = append(ha, hash)
				repeat = true
				break
			}
		}
		if !repeat {
			ust.tsHash[ts] = []string{hash}
		}
	}
}

func (ust *UpcomingScheduleTx) update(fresh bool) {
	min := int64(math.MaxInt64)
	if fresh {
		for _ts, _ := range ust.tsHash {
			if min > _ts {
				min = _ts
			}
		}
		ust.minTs = min
	} else {
		min = ust.minTs
	}

	for _ts, _ := range ust.tsHash {
		if _ts >= min + utils.CommitSeconds {
			delete(ust.tsHash, _ts)
		}
	}
}

func (ust *UpcomingScheduleTx) Due(lastTs int64) (tsHash map[int64][]string) {
	if becomeDue(ust.minTs, lastTs) {
		for _ts, _hash := range ust.tsHash {
			if becomeDue(_ts, lastTs) {
				tsHash[_ts] = _hash
			}
		}
	}
	return
}

func becomeDue(ts, lastTs int64) bool {
	if ts < lastTs || ts < lastTs - utils.CommitSeconds {
		return true
	}
	return false
}
