package errors

import "net/http"

var UnauthorizedError = NewErrorCode("unauthorized", http.StatusUnauthorized)
