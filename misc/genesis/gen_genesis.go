//package main
package genesis

import (
	"time"
	"strings"
	"math/rand"
	"fmt"
	"github.com/ethereum/go-ethereum/core"
	"math/big"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/common"
	"encoding/csv"
	"os"
	"io"
)

var (
	balance, _ = big.NewInt(0).SetString("10000000000000000000000000000000000", 10)
)

func main() {
	config := &params.ChainConfig{
		ChainID: big.NewInt(15),
		HomesteadBlock: big.NewInt(0),
		EIP155Block: big.NewInt(0),
		EIP158Block: big.NewInt(0),
	}

	gen := &core.Genesis{
		Config: config,
		Nonce: uint64(0xdeadbeefdeadbeef),
		Timestamp: uint64(0x0),
		ExtraData: nil,
		GasLimit: uint64(0x1e8480000),
		Difficulty: big.NewInt(0x40),
		Mixhash: common.HexToHash("0x0"),
		//Alloc: *(devAllocs()),
		//Alloc: *(simulateAllocs()),
		Alloc: *(mainnetAllocs()),
		ParentHash: common.HexToHash("0x0"),
	}
	//getAllocs()
	if genJSON, err := gen.MarshalJSON();  err != nil {
		panic(err)
	} else {
		fmt.Println(string(genJSON))
	}
}

func simulateAllocs() *core.GenesisAlloc {
	num := 100
	allocs := make(core.GenesisAlloc, num)
	hexes := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	// fmt.Println(time.Now().Unix())
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	var addr []string
	for i := 0; i < num; i++ {
		addr = make([]string, 40)
		for j := 0; j < 40; j++ {
			addr = append(addr, hexes[rand.Intn(len(hexes))])
		}
		allocs[common.HexToAddress(strings.Join(addr, ""))] = core.GenesisAccount{Balance: big.NewInt(0x100000)}
	}
	return &allocs
}

func devAllocs() *core.GenesisAlloc {
	allocs := make(core.GenesisAlloc, 2)
	allocs[common.HexToAddress("0x7eff122b94897ea5b0e2a9abf47b86337fafebdc")] = core.GenesisAccount{Balance: balance}
	allocs[common.HexToAddress("0x77beb894fc9b0ed41231e51f128a347043960a9d")] = core.GenesisAccount{Balance: balance}
	return &allocs
}

func mainnetAllocs() *core.GenesisAlloc {
	/* content of example.csv
	0x7eff122b94897ea5b0e2a9abf47b86337fafebdc,1000000000000000000
	0x77beb894fc9b0ed41231e51f128a347043960a9d,1000000000000000000
	*/
	file := "/tmp/erc20_cmt.csv"
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	csvr := csv.NewReader(f)

	allocs := make(core.GenesisAlloc, 10)
	for {
		row, err := csvr.Read()
		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			return &allocs
		}

		balance, success := big.NewInt(0).SetString(strings.Trim(row[1]," "), 10)
		if !success {
			panic("convert alloc balance error!")
		}
		fmt.Printf("%s: %v\n", row[0], common.HexToAddress(row[0]))
		allocs[common.HexToAddress(row[0])] = core.GenesisAccount{Balance:balance}
	}
	return &allocs
}
