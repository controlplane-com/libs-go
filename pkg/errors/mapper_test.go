package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestToHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: http.StatusOK,
		},
		{
			name:     "validation error",
			err:      Validation("invalid"),
			expected: http.StatusBadRequest,
		},
		{
			name:     "not found error",
			err:      NotFound("user", "123"),
			expected: http.StatusNotFound,
		},
		{
			name:     "conflict error",
			err:      Conflict("duplicate"),
			expected: http.StatusConflict,
		},
		{
			name:     "unauthorized error",
			err:      Unauthorized("invalid token"),
			expected: http.StatusUnauthorized,
		},
		{
			name:     "forbidden error",
			err:      Forbidden("access denied"),
			expected: http.StatusForbidden,
		},
		{
			name:     "internal error",
			err:      Internal("failed", nil),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "unavailable error",
			err:      Unavailable("down"),
			expected: http.StatusServiceUnavailable,
		},
		{
			name:     "legacy ErrorCode",
			err:      NewErrorCode("custom", http.StatusTeapot),
			expected: http.StatusTeapot,
		},
		{
			name:     "unknown error type",
			err:      errors.New("something went wrong"),
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToHTTPStatus(tt.err)
			if result != tt.expected {
				t.Errorf("ToHTTPStatus() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestToErrorCode(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "domain error",
			err:          Validation("invalid input"),
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "invalid input",
		},
		{
			name:         "legacy ErrorCode passthrough",
			err:          NewErrorCode("legacy error", http.StatusConflict),
			expectedCode: http.StatusConflict,
			expectedMsg:  "legacy error",
		},
		{
			name:         "regular error",
			err:          errors.New("unknown error"),
			expectedCode: http.StatusInternalServerError,
			expectedMsg:  "unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToErrorCode(tt.err)
			if result.Code() != tt.expectedCode {
				t.Errorf("ToErrorCode().Code() = %d, expected %d", result.Code(), tt.expectedCode)
			}
			if result.Error() != tt.expectedMsg {
				t.Errorf("ToErrorCode().Error() = %s, expected %s", result.Error(), tt.expectedMsg)
			}
		})
	}
}

func TestToErrorCodeNil(t *testing.T) {
	result := ToErrorCode(nil)
	if result != nil {
		t.Errorf("ToErrorCode(nil) should return nil, got %v", result)
	}
}
