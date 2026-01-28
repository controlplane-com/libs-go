package errors

import "fmt"

// ErrorKind represents the semantic category of an error (no HTTP concepts)
type ErrorKind string

const (
	ErrKindValidation   ErrorKind = "validation"
	ErrKindNotFound     ErrorKind = "not_found"
	ErrKindConflict     ErrorKind = "conflict"
	ErrKindUnauthorized ErrorKind = "unauthorized"
	ErrKindForbidden    ErrorKind = "forbidden"
	ErrKindInternal     ErrorKind = "internal"
	ErrKindUnavailable  ErrorKind = "unavailable"
)

// DomainError represents a business logic error without HTTP coupling.
// Services and repositories should return DomainErrors instead of ErrorCode
// to maintain proper layer separation.
type DomainError struct {
	Kind    ErrorKind
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// IsDomainError checks if an error is a DomainError and optionally matches a specific kind
func IsDomainError(err error, kind ...ErrorKind) bool {
	de, ok := err.(*DomainError)
	if !ok {
		return false
	}
	if len(kind) == 0 {
		return true
	}
	for _, k := range kind {
		if de.Kind == k {
			return true
		}
	}
	return false
}

// Validation creates a validation error (maps to 400 Bad Request)
func Validation(msg string) *DomainError {
	return &DomainError{Kind: ErrKindValidation, Message: msg}
}

// Validationf creates a validation error with formatted message
func Validationf(format string, args ...interface{}) *DomainError {
	return &DomainError{Kind: ErrKindValidation, Message: fmt.Sprintf(format, args...)}
}

// NotFound creates a not found error (maps to 404 Not Found)
func NotFound(resource, id string) *DomainError {
	return &DomainError{
		Kind:    ErrKindNotFound,
		Message: fmt.Sprintf("%s '%s' not found", resource, id),
	}
}

// NotFoundMsg creates a not found error with custom message
func NotFoundMsg(msg string) *DomainError {
	return &DomainError{Kind: ErrKindNotFound, Message: msg}
}

// Conflict creates a conflict error (maps to 409 Conflict)
func Conflict(msg string) *DomainError {
	return &DomainError{Kind: ErrKindConflict, Message: msg}
}

// Conflictf creates a conflict error with formatted message
func Conflictf(format string, args ...interface{}) *DomainError {
	return &DomainError{Kind: ErrKindConflict, Message: fmt.Sprintf(format, args...)}
}

// Unauthorized creates an unauthorized error (maps to 401 Unauthorized)
func Unauthorized(msg string) *DomainError {
	return &DomainError{Kind: ErrKindUnauthorized, Message: msg}
}

// Forbidden creates a forbidden error (maps to 403 Forbidden)
func Forbidden(msg string) *DomainError {
	return &DomainError{Kind: ErrKindForbidden, Message: msg}
}

// Internal creates an internal error (maps to 500 Internal Server Error)
func Internal(msg string, cause error) *DomainError {
	return &DomainError{Kind: ErrKindInternal, Message: msg, Cause: cause}
}

// Internalf creates an internal error with formatted message
func Internalf(cause error, format string, args ...interface{}) *DomainError {
	return &DomainError{Kind: ErrKindInternal, Message: fmt.Sprintf(format, args...), Cause: cause}
}

// Unavailable creates an unavailable error (maps to 503 Service Unavailable)
func Unavailable(msg string) *DomainError {
	return &DomainError{Kind: ErrKindUnavailable, Message: msg}
}
