package gcp

import (
	"errors"
	"fmt"
	"testing"

	helperErrors "github.com/protofire/polkadot-failover-mechanism/pkg/helpers/errors"

	"github.com/stretchr/testify/require"
)

type testError struct {
}

func (t testError) Error() string {
	return "error"
}

func TestValidatorError(t *testing.T) {

	testErr1 := fmt.Errorf("test error: %w", helperErrors.ValidatorError{})
	require.True(t, errors.Is(testErr1, helperErrors.ValidatorError{}))
	require.True(t, errors.As(testErr1, &helperErrors.ValidatorError{}))
	testErr2 := helperErrors.ValidatorError{}
	require.True(t, errors.Is(testErr2, helperErrors.ValidatorError{}))
	require.True(t, errors.As(testErr2, &helperErrors.ValidatorError{}))
	require.False(t, errors.Is(testError{}, helperErrors.ValidatorError{}))

}

func TestValidatorErrorCheck(t *testing.T) {
	testErr := helperErrors.NewValidatorError("cannot find validators", helperErrors.ValidatorErrorNotFound)
	require.True(t, errors.As(testErr, &helperErrors.ValidatorError{}))
}
