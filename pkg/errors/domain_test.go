package errors

import (
	"errors"
	"testing"
)

func TestValidation(t *testing.T) {
	err := Validation("invalid input")
	if err.Kind != ErrKindValidation {
		t.Errorf("expected kind %s, got %s", ErrKindValidation, err.Kind)
	}
	if err.Error() != "invalid input" {
		t.Errorf("expected message 'invalid input', got '%s'", err.Error())
	}
}

func TestValidationf(t *testing.T) {
	err := Validationf("field %s is required", "name")
	if err.Kind != ErrKindValidation {
		t.Errorf("expected kind %s, got %s", ErrKindValidation, err.Kind)
	}
	if err.Error() != "field name is required" {
		t.Errorf("expected message 'field name is required', got '%s'", err.Error())
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("user", "123")
	if err.Kind != ErrKindNotFound {
		t.Errorf("expected kind %s, got %s", ErrKindNotFound, err.Kind)
	}
	if err.Error() != "user '123' not found" {
		t.Errorf("expected message \"user '123' not found\", got '%s'", err.Error())
	}
}

func TestNotFoundMsg(t *testing.T) {
	err := NotFoundMsg("resource does not exist")
	if err.Kind != ErrKindNotFound {
		t.Errorf("expected kind %s, got %s", ErrKindNotFound, err.Kind)
	}
	if err.Error() != "resource does not exist" {
		t.Errorf("expected message 'resource does not exist', got '%s'", err.Error())
	}
}

func TestConflict(t *testing.T) {
	err := Conflict("resource already exists")
	if err.Kind != ErrKindConflict {
		t.Errorf("expected kind %s, got %s", ErrKindConflict, err.Kind)
	}
	if err.Error() != "resource already exists" {
		t.Errorf("expected message 'resource already exists', got '%s'", err.Error())
	}
}

func TestConflictf(t *testing.T) {
	err := Conflictf("user %s already exists", "john")
	if err.Kind != ErrKindConflict {
		t.Errorf("expected kind %s, got %s", ErrKindConflict, err.Kind)
	}
	if err.Error() != "user john already exists" {
		t.Errorf("expected message 'user john already exists', got '%s'", err.Error())
	}
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("invalid credentials")
	if err.Kind != ErrKindUnauthorized {
		t.Errorf("expected kind %s, got %s", ErrKindUnauthorized, err.Kind)
	}
	if err.Error() != "invalid credentials" {
		t.Errorf("expected message 'invalid credentials', got '%s'", err.Error())
	}
}

func TestForbidden(t *testing.T) {
	err := Forbidden("access denied")
	if err.Kind != ErrKindForbidden {
		t.Errorf("expected kind %s, got %s", ErrKindForbidden, err.Kind)
	}
	if err.Error() != "access denied" {
		t.Errorf("expected message 'access denied', got '%s'", err.Error())
	}
}

func TestInternal(t *testing.T) {
	cause := errors.New("database connection failed")
	err := Internal("failed to fetch user", cause)
	if err.Kind != ErrKindInternal {
		t.Errorf("expected kind %s, got %s", ErrKindInternal, err.Kind)
	}
	expected := "failed to fetch user: database connection failed"
	if err.Error() != expected {
		t.Errorf("expected message '%s', got '%s'", expected, err.Error())
	}
	if err.Unwrap() != cause {
		t.Errorf("expected unwrapped error to be cause")
	}
}

func TestInternalf(t *testing.T) {
	cause := errors.New("timeout")
	err := Internalf(cause, "operation %s failed", "save")
	if err.Kind != ErrKindInternal {
		t.Errorf("expected kind %s, got %s", ErrKindInternal, err.Kind)
	}
	expected := "operation save failed: timeout"
	if err.Error() != expected {
		t.Errorf("expected message '%s', got '%s'", expected, err.Error())
	}
}

func TestInternalWithoutCause(t *testing.T) {
	err := Internal("something went wrong", nil)
	if err.Error() != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got '%s'", err.Error())
	}
	if err.Unwrap() != nil {
		t.Errorf("expected nil unwrap")
	}
}

func TestUnavailable(t *testing.T) {
	err := Unavailable("service temporarily unavailable")
	if err.Kind != ErrKindUnavailable {
		t.Errorf("expected kind %s, got %s", ErrKindUnavailable, err.Kind)
	}
	if err.Error() != "service temporarily unavailable" {
		t.Errorf("expected message 'service temporarily unavailable', got '%s'", err.Error())
	}
}

func TestIsDomainError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		kinds    []ErrorKind
		expected bool
	}{
		{
			name:     "domain error without kind filter",
			err:      Validation("test"),
			kinds:    nil,
			expected: true,
		},
		{
			name:     "domain error with matching kind",
			err:      Validation("test"),
			kinds:    []ErrorKind{ErrKindValidation},
			expected: true,
		},
		{
			name:     "domain error with non-matching kind",
			err:      Validation("test"),
			kinds:    []ErrorKind{ErrKindNotFound},
			expected: false,
		},
		{
			name:     "domain error with multiple kinds - one matches",
			err:      NotFound("user", "1"),
			kinds:    []ErrorKind{ErrKindValidation, ErrKindNotFound},
			expected: true,
		},
		{
			name:     "non-domain error",
			err:      errors.New("regular error"),
			kinds:    nil,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			kinds:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDomainError(tt.err, tt.kinds...)
			if result != tt.expected {
				t.Errorf("IsDomainError() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
