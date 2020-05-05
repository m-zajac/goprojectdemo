package app

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInvalidRequestError(t *testing.T) {
	stdErr := errors.New("simple error")
	assert.False(t, IsInvalidRequestError(stdErr))

	irErr := InvalidRequestError("invalid request")
	assert.True(t, IsInvalidRequestError(irErr))

	wrapperErr := fmt.Errorf("wrapping message: %w", irErr)
	assert.True(t, IsInvalidRequestError(wrapperErr))
}
