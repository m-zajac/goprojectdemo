package app

import "errors"

// InvalidRequestError is special error type returned when any request params are invalid.
type InvalidRequestError string

// Error implements error interface.
func (e InvalidRequestError) Error() string {
	return string(e)
}

// IsInvalidRequest tells that this error is 'invalid request'.
// Returns always true.
func (InvalidRequestError) IsInvalidRequest() bool {
	return true
}

// IsInvalidRequestError checks if given error is caused by invalid request.
func IsInvalidRequestError(err error) bool {
	type invalidReqErr interface {
		IsInvalidRequest() bool
	}

	var ie invalidReqErr
	if errors.As(err, &ie) {
		return ie.IsInvalidRequest()
	}

	return false
}

// TooManyRequestsError is special error type returned when there's too many request to handle at a time.
type TooManyRequestsError string

// Error implements error interface.
func (e TooManyRequestsError) Error() string {
	return string(e)
}

// IsTooManyRequests tells that this error is 'too many requests'.
// Returns always true.
func (TooManyRequestsError) IsTooManyRequests() bool {
	return true
}

// IsTooManyRequestsError checks if given error is caused by too many requests.
func IsTooManyRequestsError(err error) bool {
	type tooManyReqErr interface {
		IsTooManyRequests() bool
	}

	var ie tooManyReqErr
	if errors.As(err, &ie) {
		return ie.IsTooManyRequests()
	}

	return false
}

// ScheduledForLaterError is special error type returned request could not be immediately processed and is scheduled for later.
type ScheduledForLaterError string

// Error implements error interface.
func (e ScheduledForLaterError) Error() string {
	return string(e)
}

// IsScheduledForLater tells that this error means 'scheduled for later processing'.
// Returns always true.
func (ScheduledForLaterError) IsScheduledForLater() bool {
	return true
}

// IsScheduledForLaterError checks if given error means 'scheduled for later processing'.
func IsScheduledForLaterError(err error) bool {
	type scheduledForLaterReqErr interface {
		IsScheduledForLater() bool
	}

	var ie scheduledForLaterReqErr
	if errors.As(err, &ie) {
		return ie.IsScheduledForLater()
	}

	return false
}
