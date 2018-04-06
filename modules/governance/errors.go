// nolint
package governance

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/errors"
)

var (
	errMissingSignature      = fmt.Errorf("Missing signature")
	errInvalidParameter      = fmt.Errorf("Invalid parameter")
	errRepeatedVote          = fmt.Errorf("Repeated vote")
	errInvalidValidator      = fmt.Errorf("Invalid validator")
)

func ErrMissingSignature() error {
	return errors.WithCode(errMissingSignature, errors.CodeTypeUnauthorized)
}

func ErrInvalidParamerter() error {
	return errors.WithCode(errInvalidParameter, errors.CodeTypeBaseInvalidInput)
}

func ErrRepeatedVote() error {
	return errors.WithCode(errRepeatedVote, errors.CodeTypeBaseInvalidInput)
}

func ErrInvalidValidator() error {
	return errors.WithCode(errInvalidValidator, errors.CodeTypeBaseInvalidInput)
}
