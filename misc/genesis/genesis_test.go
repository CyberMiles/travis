package genesis

import (
	"fmt"
	"math/big"
)

// ExampleDefaultGenesisBlock testing the allocate accounts count of genesis
func ExampleDefaultGenesisBlock() {
	genBlock := DefaultGenesisBlock()
	allocs := genBlock.Alloc
	totalAlloc := len(allocs)

	totalBalance := big.NewInt(0)

	for _, alloc := range allocs {
		totalBalance.Add(totalBalance, alloc.Balance)
	}



	fmt.Println(totalAlloc)
	fmt.Println(totalBalance.String())
	fmt.Println(string(genBlock.ExtraData))
	// Output:
	// 22035
	// 1000000000000000000000000000
	// CyberMiles for E-commerce
}
