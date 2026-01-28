package errors

import "net/http"

// ToHTTPStatus maps an error to an HTTP status code.
// This should ONLY be called at the controller/transport layer.
func ToHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Check for DomainError first
	if de, ok := err.(*DomainError); ok {
		switch de.Kind {
		case ErrKindValidation:
			return http.StatusBadRequest
		case ErrKindNotFound:
			return http.StatusNotFound
		case ErrKindConflict:
			return http.StatusConflict
		case ErrKindUnauthorized:
			return http.StatusUnauthorized
		case ErrKindForbidden:
			return http.StatusForbidden
		case ErrKindUnavailable:
			return http.StatusServiceUnavailable
		case ErrKindInternal:
			return http.StatusInternalServerError
		default:
			return http.StatusInternalServerError
		}
	}

	// Legacy ErrorCode support for backward compatibility
	if ec, ok := err.(ErrorCode); ok {
		return ec.Code()
	}

	// Default to 500 for unknown errors
	return http.StatusInternalServerError
}

// ToErrorCode converts any error to an ErrorCode for HTTP response serialization.
// This should ONLY be called at the controller/transport layer.
func ToErrorCode(err error) ErrorCode {
	if err == nil {
		return nil
	}

	// If already an ErrorCode, return as-is
	if ec, ok := err.(ErrorCode); ok {
		return ec
	}

	return NewErrorCode(err.Error(), ToHTTPStatus(err))
}
