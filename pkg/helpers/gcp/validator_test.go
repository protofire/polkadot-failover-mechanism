package gcp

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testError struct {
}

func (t testError) Error() string {
	return "error"
}

func TestValidatorError(t *testing.T) {

	testErr1 := fmt.Errorf("test error: %w", ValidatorError{})
	require.True(t, errors.Is(testErr1, ValidatorError{}))
	require.True(t, errors.As(testErr1, &ValidatorError{}))
	testErr2 := ValidatorError{}
	require.True(t, errors.Is(testErr2, ValidatorError{}))
	require.True(t, errors.As(testErr2, &ValidatorError{}))
	require.False(t, errors.Is(testError{}, ValidatorError{}))

}

func TestValidatorErrorCheck(t *testing.T) {
	testErr := NewValidatorError("cannot find validators", ValidatorErrorNotFound)
	require.True(t, errors.As(testErr, &ValidatorError{}))
}
