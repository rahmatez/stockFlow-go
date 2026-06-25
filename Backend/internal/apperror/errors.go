package apperror

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrConflict       = errors.New("conflict")
	ErrValidation     = errors.New("validation error")
	ErrLimitExceeded  = errors.New("plan limit exceeded")
	ErrInvalidStatus  = errors.New("invalid status transition")
	ErrInsufficientStock = errors.New("insufficient stock")
)
