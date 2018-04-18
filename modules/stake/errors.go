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

	errBadValidatorAddr      = fmt.Errorf("Validator does not exist for that address")
	errCandidateExistsAddr   = fmt.Errorf("Candidate already exist, cannot re-declare candidacy")
	errMissingSignature      = fmt.Errorf("Missing signature")
	errBondNotNominated      = fmt.Errorf("Cannot bond to non-nominated account")
	errNoCandidateForAddress = fmt.Errorf("Validator does not exist for that address")
	errNoDelegatorForAddress = fmt.Errorf("Delegator does not contain validator bond")
	errInsufficientFunds     = fmt.Errorf("Insufficient bond shares")
	errBadRemoveValidator    = fmt.Errorf("Error removing validator")
	errFullSlot              = fmt.Errorf("Slot is full")
	errBadSlot               = fmt.Errorf("Slot does not exist")
	errCancelledSlot         = fmt.Errorf("Slot was cancelled already")
	errBadSlotDelegate       = fmt.Errorf("Slot delegate does not exist")

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

func ErrFullSlot() error {
	return errors.WithCode(errFullSlot, errors.CodeTypeBaseInvalidInput)
}

func ErrBadSlot() error {
	return errors.WithCode(errBadSlot, errors.CodeTypeBaseInvalidInput)
}

func ErrCancelledSlot() error {
	return errors.WithCode(errCancelledSlot, errors.CodeTypeBaseInvalidInput)
}

func ErrBadSlotDelegate() error {
	return errors.WithCode(errBadSlotDelegate, errors.CodeTypeBaseInvalidInput)
}

func ErrBadAmount() error {
	return errors.WithCode(errBadAmount, errors.CodeTypeBaseInvalidOutput)
}
