// nolint
package stake

import (
	"fmt"

	"github.com/CyberMiles/travis/sdk/errors"
)

var (
	errBadAmount                          = fmt.Errorf("Amount must be > 0")
	errBadCompRate                        = fmt.Errorf("Compensation rate must between 0 and 1, and less than the default value")
	errBadValidatorAddr                   = fmt.Errorf("Candidate does not exist for that address")
	errCandidateExistsAddr                = fmt.Errorf("Candidate already exists, cannot re-declare candidate")
	errMissingSignature                   = fmt.Errorf("Missing signature")
	errInsufficientFunds                  = fmt.Errorf("Insufficient funds")
	errCandidateVerificationDisallowed    = fmt.Errorf("Verification disallowed")
	errCandidateVerifiedAlready           = fmt.Errorf("Candidate has been verified already")
	errReachMaxAmount                     = fmt.Errorf("Validator has reached its declared max amount CMTs to be staked")
	errDelegationNotExists                = fmt.Errorf("No corresponding delegation exists")
	errInvalidWithdrawalAmount            = fmt.Errorf("Invalid withdrawal amount")
	errCandidateWithdrawalDisallowed      = fmt.Errorf("Candidate can't withdraw the self-staking funds")
	errInvalidCubeSignature               = fmt.Errorf("Invalid cube signature")
	errCandidateHasPendingUnstakeRequests = fmt.Errorf("Candidate has some pending withdrawal requests")
	errAddressAlreadyDeclared             = fmt.Errorf("Address has been declared")
	errPubKeyAlreadyDeclared              = fmt.Errorf("PubKey has been declared")
	errCandidateAlreadyActivated          = fmt.Errorf("Candidate has been activated")
	errCandidateAlreadyDeactivated        = fmt.Errorf("Candidate has been deactivated")
	errBadRequest                         = fmt.Errorf("Bad request")

	invalidInput = errors.CodeTypeBaseInvalidInput
)

func ErrBadValidatorAddr() error {
	return errors.WithCode(errBadValidatorAddr, errors.CodeTypeBaseUnknownAddress)
}
func ErrCandidateExistsAddr() error {
	return errors.WithCode(errCandidateExistsAddr, errors.CodeTypeBaseInvalidInput)
}
func ErrMissingSignature() error {
	return errors.WithCode(errMissingSignature, errors.CodeTypeUnauthorized)
}

func ErrInsufficientFunds() error {
	return errors.WithCode(errInsufficientFunds, errors.CodeTypeBaseInvalidInput)
}

func ErrBadAmount() error {
	return errors.WithCode(errBadAmount, errors.CodeTypeBaseInvalidOutput)
}

func ErrVerificationDisallowed() error {
	return errors.WithCode(errCandidateVerificationDisallowed, errors.CodeTypeBaseInvalidOutput)
}

func ErrReachMaxAmount() error {
	return errors.WithCode(errReachMaxAmount, errors.CodeTypeBaseInvalidOutput)
}

func ErrDelegationNotExists() error {
	return errors.WithCode(errDelegationNotExists, errors.CodeTypeBaseInvalidOutput)
}

func ErrInvalidWithdrawalAmount() error {
	return errors.WithCode(errInvalidWithdrawalAmount, errors.CodeTypeBaseInvalidOutput)
}

func ErrCandidateWithdrawalDisallowed() error {
	return errors.WithCode(errCandidateWithdrawalDisallowed, errors.CodeTypeBaseInvalidOutput)
}

func ErrInvalidCubeSignature() error {
	return errors.WithCode(errInvalidCubeSignature, errors.CodeTypeBaseInvalidOutput)
}

func ErrBadCompRate() error {
	return errors.WithCode(errBadCompRate, errors.CodeTypeBaseInvalidOutput)
}

func ErrCandidateHasPendingUnstakeRequests() error {
	return errors.WithCode(errCandidateHasPendingUnstakeRequests, errors.CodeTypeBaseInvalidOutput)
}

func ErrBadRequest() error {
	return errors.WithCode(errBadRequest, errors.CodeTypeBaseInvalidOutput)
}

func ErrAddressAlreadyDeclared() error {
	return errors.WithCode(errAddressAlreadyDeclared, errors.CodeTypeBaseInvalidOutput)
}

func ErrPubKeyAleadyDeclared() error {
	return errors.WithCode(errPubKeyAlreadyDeclared, errors.CodeTypeBaseInvalidOutput)
}

func ErrCandidateAlreadyActivated() error {
	return errors.WithCode(errCandidateAlreadyActivated, errors.CodeTypeBaseInvalidOutput)
}

func ErrCandidateAlreadyDeactivated() error {
	return errors.WithCode(errCandidateAlreadyDeactivated, errors.CodeTypeBaseInvalidOutput)
}
