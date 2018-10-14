package common

import (
	"fmt"
)

// A panic resulting from a sanity check means there is a programmer error
// and some guarantee is not satisfied.
// XXX DEPRECATED
func PanicSanity(v interface{}) {
	panic(fmt.Sprintf("Panicked on a Sanity Check: %v", v))
}
