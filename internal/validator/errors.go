package validator

import "errors"

var (
	ErrMissingToken     = errors.New("missing token")
	ErrValidationFailed = errors.New("validation failed")
	ErrTimeout          = errors.New("validation timeout")
	ErrInternal         = errors.New("internal error")
)
