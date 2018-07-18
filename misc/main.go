package main

import (
	"fmt"
	"github.com/CyberMiles/travis/misc/genesis"
)

func main() {
	defaltConfig := genesis.DevGenesisBlock();
	gen, err := defaltConfig.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(gen))
}
