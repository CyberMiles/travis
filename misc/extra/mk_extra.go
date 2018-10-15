package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"runtime"
)

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"geth",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

func main() {
	//data := hexutil.MustDecode("0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa")
	//fmt.Println(string(hexutil.Encode("Powered by Cybermiles")))
	fmt.Println(hexutil.Encode(makeExtraData(nil)))
	fmt.Println(hexutil.Encode(makeExtraData([]byte("CyberMiles Foundation Limited"))))
	fmt.Println(string(hexutil.MustDecode("0x43796265724d696c657320466f756e646174696f6e204c696d69746564")))

	fmt.Println(hexutil.Encode(makeExtraData([]byte("CyberMiles for E-commerce"))))
	fmt.Println(string(hexutil.MustDecode("0x43796265724d696c657320666f7220452d636f6d6d65726365")))
}
