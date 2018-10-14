package genesis

import (
	"fmt"
)

// ExampleDefaultGenesisBlock testing the allocate accounts count of genesis
func ExampleDefaultGenesisBlock() {
	totalAlloc := len(DefaultGenesisBlock().Alloc)
	fmt.Println(totalAlloc)

	// Output:
	// 22034
}
