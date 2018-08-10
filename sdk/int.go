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

func add(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Add(i, i2) }

func sub(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Sub(i, i2) }

func mul(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Mul(i, i2) }

func div(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Div(i, i2) }

func (i Int) Add(i2 Int) Int {
	return Int{add(i.Int, i2.Int)}
}

func (i Int) Sub(i2 Int) Int {
	return Int{sub(i.Int, i2.Int)}
}

func (i Int) Mul(i2 Int) Int {
	return Int{mul(i.Int, i2.Int)}
}

func (i Int) Div(i2 Int) Int {
	return Int{div(i.Int, i2.Int)}
}

func (i Int) MulRat(r Rat) Int {
	return Int{div(mul(i.Int, r.Num()), r.Denom())}
}
