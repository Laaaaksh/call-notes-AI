package apperror

import "net/http"

// Code represents an error code type
type Code string

const (
	CodeBadRequest         Code = "BAD_REQUEST"
	CodeNotFound           Code = "NOT_FOUND"
	CodeConflict           Code = "CONFLICT"
	CodeInternalError      Code = "INTERNAL_ERROR"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
	CodeValidationError    Code = "VALIDATION_ERROR"
	CodeDuplicateRequest   Code = "DUPLICATE_REQUEST"
	CodeRateLimited        Code = "RATE_LIMITED"
	CodeTimeout            Code = "TIMEOUT"
)

// HTTPStatus returns the HTTP status code for an error code
func (c Code) HTTPStatus() int {
	switch c {
	case CodeBadRequest, CodeValidationError:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict, CodeDuplicateRequest:
		return http.StatusConflict
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeTimeout:
		return http.StatusGatewayTimeout
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case CodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// String returns the string representation of the error code
func (c Code) String() string {
	return string(c)
}
