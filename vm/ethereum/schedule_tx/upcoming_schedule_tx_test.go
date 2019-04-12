package schedule_tx

import (
	"testing"
)

var baseTs int64 = 1555058000
var hash = "0x9f7cdbafefb48b21e2174709ae3188061ec2aa7c7368eeee9de867596ac2d500"
var hash1 = "0xb68c0da126d0f202a2b3edf2482ecb905a20375c0197b0fe01702c20ed20b9d3"
var hash2 = "0x5c9195d1fa5418e518cdd65d34d60ec32d84190fb26903c4d83cec681fc1d6ad"
var hash3 = "0xf2349454a91ce479bfcc5585849000a1ba0aa8ac843a4398b7e0c5db38cac476"
var hash4 = "0xe64c0cf7d5fccfa367a308762f8736a4467aa22f5c5be533ff8732ea11ea3659"
var hash5 = "0x3e71849f09a266f0612b2e7a300a32ccc2f01b2e20ab8929b96a27d55bf6259d"
var hash6 = "0x1fb63d99f775e08f847734a75692d5c88e4bd3356780c2f01b74334ed914d685"
var hash7 = "0x9f7cdbafefb48b21e2174709ae3188061ec2aa7c7368eeee9de867596ac2d500"
var hash8 = "0xff41c3019e5d10435c371384451ba144469ecc50c3f89726a79233dd896dbb80"
var hash9 = "0xaba183f4455ca929edd0f98297f4a9db036387015adf93e0e77835d00a06e142"

func TestAdd(t *testing.T) {
	ust := UpcomingScheduleTx{}

	ust.Add(baseTs, hash)
	if ust.minTs != baseTs {
		t.Error("minTs should be", baseTs)
	}
	ha := ust.tsHash[baseTs]
	if len(ha) != 1 || ha[0] != hash {
		t.Error("tsHash not as expected")
	}

	ts := baseTs + 3
	ust.Add(ts, hash1)
	if ust.minTs != baseTs {
		t.Error("minTs should be", baseTs)
	}
	ha = ust.tsHash[ts]
	if len(ha) != 1 || ha[0] != hash1 {
		t.Error("tsHash not as expected")
	}

	ts = baseTs + 11
	ust.Add(ts, hash2)
	ha = ust.tsHash[ts]
	if ha != nil {
		t.Error("expect tsHash to be nil")
	}

	ts = baseTs + 3
	ust.Add(ts, hash2)
	ha = ust.tsHash[ts]
	if len(ha) != 2 {
		t.Error("length of tsHash is expected to be 2, but got", len(ha))
	}

	ust.Add(baseTs, hash3)
	ha = ust.tsHash[baseTs]
	if len(ha) != 2 {
		t.Error("length of tsHash is expected to be 2, but got", len(ha))
	}

	ts = baseTs - 2
	ust.Add(ts, hash4)
	if ust.minTs != ts {
		t.Error("minTs should be", ts)
	}
	ha = ust.tsHash[ts]
	if len(ha) != 1 {
		t.Error("tsHash not as expected")
	}
	ha = ust.tsHash[baseTs]
	if len(ha) != 2 {
		t.Error("length of tsHash is expected to be 2, but got", len(ha))
	}
	ts = baseTs + 3
	ha = ust.tsHash[ts]
	if len(ha) != 2 {
		t.Error("length of tsHash is expected to be 2, but got", len(ha))
	}
	if len(ust.tsHash) != 3 {
		t.Error("length of tsHash is expected to be 3, but got", len(ha))
	}

	ts = baseTs - 7
	ust.Add(ts, hash5)
	if ust.minTs != ts {
		t.Error("minTs should be", ts)
	}
	ha = ust.tsHash[ts]
	if len(ha) != 1 {
		t.Error("tsHash not as expected")
	}
	ts = baseTs + 3
	ha = ust.tsHash[ts]
	if ha != nil {
		t.Error("length of tsHash is expected to be 2, but got", len(ha))
	}

	ts = baseTs - 17
	ust.Add(ts, hash6)
	if len(ust.tsHash) != 1 {
		t.Error("length of tsHash is expected to be 1, but got", len(ha))
	}
}

func TestDue(t *testing.T) {
	ust := UpcomingScheduleTx{}

	ust.Add(baseTs, hash)
	ust.Add(baseTs+3, hash1)
	ust.Add(baseTs+3, hash2)
	ust.Add(baseTs+7, hash3)
	ust.Add(baseTs+8, hash4)
	tsHash := ust.Due(baseTs - 4)
	if len(tsHash) != 2 {
		t.Error("length of tsHash is expected to be 1, but got", len(tsHash))
	}
	ha := tsHash[baseTs+3]
	if len(ha) != 2 {
		t.Error("length of tsHash is expected to be 1, but got", len(ha))
	}
	if len(ust.tsHash) != 2 {
		t.Error("length of tsHash is expected to be 1, but got", len(ust.tsHash))
	}
}
