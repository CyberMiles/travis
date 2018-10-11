// nolint
package stake

import (
	"fmt"

	"github.com/CyberMiles/travis/sdk/errors"
)

var (
	errCandidateEmpty     = fmt.Errorf("Cannot bond to an empty candidate")
	errBadAmount          = fmt.Errorf("Amount must be > 0")
	errNoBondingAcct      = fmt.Errorf("No bond account for this (address, validator) pair")
	errCommissionNegative = fmt.Errorf("Commission must be positive")
	errCommissionHuge     = fmt.Errorf("Commission cannot be more than 100%")
	errBadCompRate        = fmt.Errorf("Compensation rate must between 0 and 1, and less than the default value")

	errBadValidatorAddr                   = fmt.Errorf("Validator does not exist for that address")
	errCandidateExistsAddr                = fmt.Errorf("Candidate already exist, cannot re-declare candidacy")
	errMissingSignature                   = fmt.Errorf("Missing signature")
	errBondNotNominated                   = fmt.Errorf("Cannot bond to non-nominated account")
	errNoDelegatorForAddress              = fmt.Errorf("Delegator does not contain validator bond")
	errInsufficientFunds                  = fmt.Errorf("Insufficient bond shares")
	errBadRemoveValidator                 = fmt.Errorf("Error removing validator")
	errCandidateVerificationDisallowed    = fmt.Errorf("Verification disallowed")
	errCandidateVerifiedAlready           = fmt.Errorf("Candidate has been verified already")
	errReachMaxAmount                     = fmt.Errorf("Validator has reached its declared max amount CMTs to be staked")
	errDelegationNotExists                = fmt.Errorf("No corresponding delegation exists")
	errInvalidWithdrawalAmount            = fmt.Errorf("Invalid withdrawal amount")
	errCandidateWithdrawalDisallowed      = fmt.Errorf("Candidate can't withdraw the self-staking funds")
	errInvalidCubeSignature               = fmt.Errorf("Invalid cube signature")
	errCandidateHasPendingUnstakeRequests = fmt.Errorf("The candidate has some pending withdrawal requests")
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
