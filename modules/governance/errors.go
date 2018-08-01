// nolint
package governance

import (
	"fmt"

	"github.com/CyberMiles/travis/sdk/errors"
)

var (
	errMissingSignature      = fmt.Errorf("Missing signature")
	errInvalidParameter      = fmt.Errorf("Invalid parameter")
	errInsufficientParameters= fmt.Errorf("Insufficient parameters")
	errInvalidExpire         = fmt.Errorf("Invalid expire time")
	errRepeatedVote          = fmt.Errorf("Repeated vote")
	errInvalidValidator      = fmt.Errorf("Invalid validator")
	errInsufficientBalance   = fmt.Errorf("Insufficient balance")
	errApprovedProposal      = fmt.Errorf("The proposal has been approved")
	errRejectedProposal      = fmt.Errorf("The proposal has been rejected")
	errInvalidFileurlJson    = fmt.Errorf("The fileurl is not a valid json")
	errInvalidMd5Json        = fmt.Errorf("The md5 is not a valid json")
	errNoFileurl             = fmt.Errorf("Can not find fileurl for current os")
	errNoMd5                 = fmt.Errorf("Can not find md5 for current os")
)

func ErrMissingSignature() error {
	return errors.WithCode(errMissingSignature, errors.CodeTypeUnauthorized)
}

func ErrInvalidParameter() error {
	return errors.WithCode(errInvalidParameter, errors.CodeTypeBaseInvalidInput)
}

func ErrInsufficientParameters() error {
	return errors.WithCode(errInsufficientParameters, errors.CodeTypeBaseInvalidInput)
}

func ErrInvalidExpire() error {
	return errors.WithCode(errInvalidExpire, errors.CodeTypeBaseInvalidInput)
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

func ErrInvalidFileurlJson() error {
	return errors.WithCode(errInvalidFileurlJson, errors.CodeTypeBaseInvalidInput)
}

func ErrInvalidMd5Json() error {
	return errors.WithCode(errInvalidMd5Json, errors.CodeTypeBaseInvalidInput)
}

func ErrNoFileurl() error {
	return errors.WithCode(errNoFileurl, errors.CodeTypeBaseInvalidInput)
}

func ErrNoMd5() error {
	return errors.WithCode(errNoMd5, errors.CodeTypeBaseInvalidInput)
}
