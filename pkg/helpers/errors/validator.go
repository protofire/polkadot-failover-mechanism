package errors

import "fmt"

type ValidatorErrorType int

func (vt ValidatorErrorType) String() string {
	switch vt {
	case ValidatorErrorUnknown:
		return "ValidatorErrorUnknown"
	case ValidatorErrorNotFound:
		return "ValidatorErrorNotFound"
	case ValidatorErrorMultiple:
		return "ValidatorErrorMultiple"
	}
	return ""
}

const (
	ValidatorErrorUnknown ValidatorErrorType = iota
	ValidatorErrorNotFound
	ValidatorErrorMultiple
)

type ValidatorError struct {
	Message string
	Kind    ValidatorErrorType
}

func NewValidatorError(message string, kind ValidatorErrorType) ValidatorError {
	return ValidatorError{
		Message: message,
		Kind:    kind,
	}
}

func (v ValidatorError) Error() string {
	return fmt.Sprintf("%s: Error type: %s", v.Message, v.Kind)
}

func (v ValidatorError) IsNotFound() bool {
	return v.Kind == ValidatorErrorNotFound
}

func (v ValidatorError) MultipleValidators() bool {
	return v.Kind == ValidatorErrorMultiple
}
