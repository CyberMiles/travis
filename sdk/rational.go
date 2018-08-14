package sdk

import (
	"encoding/json"
	"math/big"
)

type Rat struct {
	*big.Rat `json:"rat"`
}

func ZeroRat() Rat {
	return Rat{big.NewRat(0, 1)}
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

func (r Rat) Cmp(r2 Rat) int {
	return r.Rat.Cmp(r2.Rat)
}

//Wraps r.MarshalText().
func (r Rat) MarshalJson() ([]byte, error) {
	if r.Rat == nil {
		r.Rat = new(big.Rat)
	}

	text, err := r.Rat.MarshalText()
	if err != nil {
		return nil, err
	}

	return json.Marshal(string(text))
}

// Requires a valid JSON string - strings quotes and calls UnmarshalText
func (r *Rat) UnmarshalJson(text []byte) (err error) {
	tempRat := big.NewRat(0, 1)
	err = tempRat.UnmarshalText([]byte(text))
	if err != nil {
		return err
	}

	r.Rat = tempRat
	return nil
}

func (r Rat) Equal(r2 Rat) bool { return (r.Rat).Cmp(r2.Rat) == 0 }

func (r Rat) GT(r2 Rat) bool { return (r.Rat).Cmp(r2.Rat) == 1 }

func (r Rat) GTE(r2 Rat) bool { return !r.LT(r2) }

func (r Rat) LT(r2 Rat) bool { return (r.Rat).Cmp(r2.Rat) == -1 }

func (r Rat) LTE(r2 Rat) bool { return !r.GT(r2) }
