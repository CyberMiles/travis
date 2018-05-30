// nolint
package governance

import (
	"fmt"

	"github.com/CyberMiles/travis/sdk/errors"
)

var (
	errMissingSignature      = fmt.Errorf("Missing signature")
	errInvalidParameter      = fmt.Errorf("Invalid parameter")
	errRepeatedVote          = fmt.Errorf("Repeated vote")
	errInvalidValidator      = fmt.Errorf("Invalid validator")
	errInsufficientBalance   = fmt.Errorf("Insufficient balance")
	errApprovedProposal         = fmt.Errorf("The proposal has been approved")
	errRejectedProposal         = fmt.Errorf("The proposal has been rejected")
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

func ErrInsufficientBalance() error {
	return errors.WithCode(errInsufficientBalance, errors.CodeTypeBaseInvalidInput)
}

func ErrApprovedProposal() error {
	return errors.WithCode(errApprovedProposal, errors.CodeTypeBaseInvalidInput)
}

func ErrRejectedProposal() error {
	return errors.WithCode(errRejectedProposal, errors.CodeTypeBaseInvalidInput)
}
