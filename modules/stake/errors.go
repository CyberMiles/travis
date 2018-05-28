// nolint
package stake

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/errors"
)

var (
	errCandidateEmpty     = fmt.Errorf("Cannot bond to an empty candidate")
	errBadAmount          = fmt.Errorf("Amount must be > 0")
	errNoBondingAcct      = fmt.Errorf("No bond account for this (address, validator) pair")
	errCommissionNegative = fmt.Errorf("Commission must be positive")
	errCommissionHuge     = fmt.Errorf("Commission cannot be more than 100%")

	errBadValidatorAddr                = fmt.Errorf("Validator does not exist for that address")
	errCandidateExistsAddr             = fmt.Errorf("Candidate already exist, cannot re-declare candidacy")
	errMissingSignature                = fmt.Errorf("Missing signature")
	errBondNotNominated                = fmt.Errorf("Cannot bond to non-nominated account")
	errNoCandidateForAddress           = fmt.Errorf("Validator does not exist for that address")
	errNoDelegatorForAddress           = fmt.Errorf("Delegator does not contain validator bond")
	errInsufficientFunds               = fmt.Errorf("Insufficient bond shares")
	errBadRemoveValidator              = fmt.Errorf("Error removing validator")
	errCandidateVerificationDisallowed = fmt.Errorf("verification disallowed")
	errCandidateVerifiedAlready        = fmt.Errorf("candidate has been verified already")
	errReachMaxAmount                  = fmt.Errorf("validator has reached its declared max amount CMTs to be staked")
	errDelegationNotExists             = fmt.Errorf("no corresponding delegation exists")
	errInvalidWithdrawalAmount         = fmt.Errorf("invalid withdrawal amount")
	errCandidateWithdrawalDisallowed   = fmt.Errorf("candidate can't withdraw the self-staking funds")

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
func ErrBondNotNominated() error {
	return errors.WithCode(errBondNotNominated, errors.CodeTypeBaseInvalidOutput)
}
func ErrNoCandidateForAddress() error {
	return errors.WithCode(errNoCandidateForAddress, errors.CodeTypeBaseUnknownAddress)
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

func ErrVerifiedAlready() error {
	return errors.WithCode(errCandidateVerifiedAlready, errors.CodeTypeBaseInvalidOutput)
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
