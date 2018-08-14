package utils

import (
	"fmt"
	"github.com/CyberMiles/travis/sdk"
	"math/big"
	"strconv"
)

func ParseFloat(str string) float64 {
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}

	return value
}

func ParseInt(str string) sdk.Int {
	value, ok := sdk.NewIntFromString(str)
	if !ok {
		return sdk.ZeroInt()
	}

	return value
}

func ToWei(value int64) (result *big.Int) {
	result = new(big.Int)
	result.Mul(big.NewInt(value), big.NewInt(1e18))
	return
}

func RoundFloat(f float64, n int) float64 {
	format := "%." + strconv.Itoa(n) + "f"
	res, _ := strconv.ParseFloat(fmt.Sprintf(format, f), 64)
	return res
}
