package app

import "github.com/pkg/errors"

// InvalidRequestError is special error type returned when any request params are invalid
type InvalidRequestError string

// Error implements error interface
func (e InvalidRequestError) Error() string {
	return string(e)
}

// IsInvalidRequest tells that this error is 'invalid request'.
// Returns always true.
func (InvalidRequestError) IsInvalidRequest() bool {
	return true
}

// IsInvalidRequestError checks if given error is caused by invalid request
func IsInvalidRequestError(err error) bool {
	type invalidReqErr interface {
		IsInvalidRequest() bool
	}

	err = errors.Cause(err)
	if ire, ok := err.(invalidReqErr); ok {
		return ire.IsInvalidRequest()
	}

	return false
}
