package domain

import "errors"

// Sentinel errors the adapters map to transport-level responses: httpapi
// turns ErrNotFound into 404, ErrForbidden into 403 and ErrValidation into
// 400, so services never need to know about HTTP.
var (
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("forbidden")
	ErrValidation = errors.New("validation failed")
	ErrConflict   = errors.New("conflict")
)

// ValidationError names the field that failed, wrapping ErrValidation so
// callers can match with errors.Is while still reporting specifics.
type ValidationError struct {
	Field  string
	Reason string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Reason
}

func (e ValidationError) Unwrap() error {
	return ErrValidation
}

// Invalid is a shorthand constructor for ValidationError.
func Invalid(field, reason string) error {
	return ValidationError{Field: field, Reason: reason}
}
