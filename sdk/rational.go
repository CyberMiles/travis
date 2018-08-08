package sdk

import "math/big"

type Rat struct {
	*big.Rat `json:"rat"`
}

func OneRat() Rat {
	return Rat{big.NewRat(1, 1)}
}

func NewRat(a, b int64) Rat {
	return Rat{big.NewRat(a, b)}
}

func (r Rat) Add(r2 Rat) Rat {
	return Rat{new(big.Rat).Add(r.Rat, r2.Rat)}
}

func (r Rat) Sub(r2 Rat) Rat {
	return Rat{new(big.Rat).Sub(r.Rat, r2.Rat)}
}

func (r Rat) Mul(r2 Rat) Rat {
	return Rat{new(big.Rat).Mul(r.Rat, r2.Rat)}
}

func (r Rat) Quo(r2 Rat) Rat {
	return Rat{new(big.Rat).Quo(r.Rat, r2.Rat)}
}
