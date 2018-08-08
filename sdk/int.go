package sdk

import "math/big"

type Int struct {
	*big.Int `json:"int"`
}

func NewInt(x int64) Int {
	return Int{big.NewInt(x)}
}

func NewIntFromBigInt(i *big.Int) Int {
	return Int{i}
}

func NewIntFromString(s string) (res Int, ok bool) {
	i, ok := new(big.Int).SetString(s, 0)
	if !ok {
		return
	}

	return Int{i}, true
}

func (i Int) Add(i2 Int) Int {
	return Int{new(big.Int).Add(i.Int, i2.Int)}
}

func (i Int) Sub(i2 Int) Int {
	return Int{new(big.Int).Sub(i.Int, i2.Int)}
}

func (i Int) Mul(i2 Int) Int {
	return Int{new(big.Int).Mul(i.Int, i2.Int)}
}

func (i Int) Div(i2 Int) Int {
	return Int{new(big.Int).Div(i.Int, i2.Int)}
}
