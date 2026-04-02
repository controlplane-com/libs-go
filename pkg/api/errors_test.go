package api

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name:     "with code",
			err:      &APIError{Status: 404, Message: "not found", Code: "NOT_FOUND"},
			expected: "api error (status=404, code=NOT_FOUND): not found",
		},
		{
			name:     "without code",
			err:      &APIError{Status: 500, Message: "internal error"},
			expected: "api error (status=500): internal error",
		},
		{
			name:     "empty message",
			err:      &APIError{Status: 400, Code: "BAD_REQUEST"},
			expected: "api error (status=400, code=BAD_REQUEST): ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	tests := []struct {
		status   int
		expected error
	}{
		{400, ErrBadRequest},
		{401, ErrUnauthorized},
		{403, ErrForbidden},
		{404, ErrNotFound},
		{409, ErrConflict},
		{500, ErrInternalError},
		{502, ErrInternalError},
		{503, ErrInternalError},
		{504, ErrInternalError},
		{418, nil}, // Unknown status
		{201, nil}, // Success status
		{301, nil}, // Redirect status
	}

	for _, tt := range tests {
		t.Run("status_"+string(rune('0'+tt.status/100))+string(rune('0'+tt.status%100/10))+string(rune('0'+tt.status%10)), func(t *testing.T) {
			err := &APIError{Status: tt.status}
			if got := err.Unwrap(); got != tt.expected {
				t.Errorf("Unwrap() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"api error 404", &APIError{Status: 404, Message: "not found"}, true},
		{"api error 500", &APIError{Status: 500}, false},
		{"sentinel error", ErrNotFound, true},
		{"wrapped error", errors.New("some error"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"api error 401", &APIError{Status: 401}, true},
		{"api error 403", &APIError{Status: 403}, false},
		{"sentinel error", ErrUnauthorized, true},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthorized(tt.err); got != tt.expected {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsForbidden(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"api error 403", &APIError{Status: 403}, true},
		{"api error 401", &APIError{Status: 401}, false},
		{"sentinel error", ErrForbidden, true},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForbidden(tt.err); got != tt.expected {
				t.Errorf("IsForbidden() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsConflict(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"api error 409", &APIError{Status: 409}, true},
		{"api error 404", &APIError{Status: 404}, false},
		{"sentinel error", ErrConflict, true},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConflict(tt.err); got != tt.expected {
				t.Errorf("IsConflict() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsBadRequest(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"api error 400", &APIError{Status: 400}, true},
		{"api error 404", &APIError{Status: 404}, false},
		{"sentinel error", ErrBadRequest, true},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBadRequest(tt.err); got != tt.expected {
				t.Errorf("IsBadRequest() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorsIs(t *testing.T) {
	// Test that errors.Is works correctly with APIError
	tests := []struct {
		name     string
		err      *APIError
		target   error
		expected bool
	}{
		{"404 is ErrNotFound", &APIError{Status: 404}, ErrNotFound, true},
		{"401 is ErrUnauthorized", &APIError{Status: 401}, ErrUnauthorized, true},
		{"403 is ErrForbidden", &APIError{Status: 403}, ErrForbidden, true},
		{"409 is ErrConflict", &APIError{Status: 409}, ErrConflict, true},
		{"400 is ErrBadRequest", &APIError{Status: 400}, ErrBadRequest, true},
		{"500 is ErrInternalError", &APIError{Status: 500}, ErrInternalError, true},
		{"404 is not ErrUnauthorized", &APIError{Status: 404}, ErrUnauthorized, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.expected {
				t.Errorf("errors.Is() = %v, want %v", got, tt.expected)
			}
		})
	}
}
