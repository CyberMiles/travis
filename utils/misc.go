package utils

import (
	"bytes"
	"fmt"
	"github.com/CyberMiles/travis/sdk"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strconv"
	"strings"
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
		return sdk.ZeroInt
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

func LeftPad(str string, count int) string {
	padding := strings.Repeat("0", count)
	return fmt.Sprintf("%s%s", padding, str)
}

func IsEmptyAddress(address common.Address) bool {
	emptyAddress := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	return bytes.Equal(emptyAddress, address.Bytes())
}

func ConvertDaysToHeight(days int64) int64 {
	return days * 24 * 60 * 60 / CommitSeconds
}

func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
