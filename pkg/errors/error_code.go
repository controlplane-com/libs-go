package errors

import "errors"

type ErrorCode interface {
	error
	Code() int
}

type errorCode struct {
	error
	ErrorCode    int    `json:"code,omitempty"`
	ErrorMessage string `json:"error"`
}

func (e *errorCode) Error() string {
	return e.error.Error()
}

func (e *errorCode) Code() int {
	return e.ErrorCode
}

func NewErrorCode(message string, code int) ErrorCode {
	return &errorCode{
		error:        errors.New(message),
		ErrorMessage: message,
		ErrorCode:    code,
	}
}
