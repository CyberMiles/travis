package sdk

import (
	"encoding/json"
	"math/big"
)

type Rat struct {
	*big.Rat `json:"rat"`
}

var (
	ZeroRat = NewRat(0, 1)
	OneRat  = NewRat(1, 1)
)

func NewRat(a, b int64) Rat {
	return Rat{big.NewRat(a, b)}
}

func NewRatFromString(s string) (Rat, bool) {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return Rat{}, false
	}

	return Rat{r}, true
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
func (r Rat) MarshalJSON() ([]byte, error) {
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
func (r *Rat) UnmarshalJSON(b []byte) (err error) {
	var text string
	if err = json.Unmarshal(b, &text); err != nil {
		return
	}
	tempRat := new(big.Rat)
	if err = tempRat.UnmarshalText([]byte(text)); err != nil {
		return
	}

	r.Rat = tempRat
	return
}

func (r Rat) Equal(r2 Rat) bool { return (r.Rat).Cmp(r2.Rat) == 0 }

func (r Rat) GT(r2 Rat) bool { return (r.Rat).Cmp(r2.Rat) == 1 }

func (r Rat) GTE(r2 Rat) bool { return !r.LT(r2) }

func (r Rat) LT(r2 Rat) bool { return (r.Rat).Cmp(r2.Rat) == -1 }

func (r Rat) LTE(r2 Rat) bool { return !r.GT(r2) }

func (r Rat) IsNil() bool { return r.Rat == nil }
