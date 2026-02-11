package api

import (
	"errors"
	"fmt"
)

// Common errors returned by the API client.
var (
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("resource conflict")
	ErrBadRequest    = errors.New("bad request")
	ErrInternalError = errors.New("internal server error")
)

// APIError represents an error response from the Control Plane API.
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
	ID      string `json:"id,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("api error (status=%d, code=%s): %s", e.Status, e.Code, e.Message)
	}
	return fmt.Sprintf("api error (status=%d): %s", e.Status, e.Message)
}

// Unwrap returns the underlying sentinel error based on status code.
func (e *APIError) Unwrap() error {
	switch e.Status {
	case 400:
		return ErrBadRequest
	case 401:
		return ErrUnauthorized
	case 403:
		return ErrForbidden
	case 404:
		return ErrNotFound
	case 409:
		return ErrConflict
	case 500, 502, 503, 504:
		return ErrInternalError
	default:
		return nil
	}
}

// IsNotFound returns true if the error is a 404 Not Found error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden returns true if the error is a 403 Forbidden error.
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsConflict returns true if the error is a 409 Conflict error.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsBadRequest returns true if the error is a 400 Bad Request error.
func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest)
}
