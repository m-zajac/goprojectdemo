package app

import (
	"testing"

	"github.com/pkg/errors"
)

func TestIsInvalidRequestError(t *testing.T) {
	stdErr := errors.New("simple error")
	if IsInvalidRequestError(stdErr) {
		t.Error("simple error reported as invalid request")
	}

	irErr := InvalidRequestError("invalid request")
	if !IsInvalidRequestError(irErr) {
		t.Error("invalid request error not reported as invalid request")
	}

	wrapperErr := errors.Wrap(irErr, "wrapping message")
	if !IsInvalidRequestError(wrapperErr) {
		t.Error("wrapped invalid request error not reported as invalid request")
	}
}
