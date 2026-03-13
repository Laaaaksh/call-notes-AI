package apperror

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeBadRequest         Code = "BAD_REQUEST"
	CodeNotFound           Code = "NOT_FOUND"
	CodeConflict           Code = "CONFLICT"
	CodeInternalError       Code = "INTERNAL_ERROR"
	CodeUnauthorized        Code = "UNAUTHORIZED"
	CodeServiceUnavailable  Code = "SERVICE_UNAVAILABLE"
)

func (c Code) String() string { return string(c) }

func (c Code) HTTPStatus() int {
	switch c {
	case CodeBadRequest:
		return 400
	case CodeUnauthorized:
		return 401
	case CodeNotFound:
		return 404
	case CodeConflict:
		return 409
	case CodeServiceUnavailable:
		return 503
	default:
		return 500
	}
}

type IError interface {
	error
	Code() Code
	PublicMessage() string
	HTTPStatus() int
}

type Error struct {
	code          Code
	cause         error
	publicMessage string
}

var _ IError = (*Error)(nil)

func New(code Code, cause error) *Error {
	return &Error{code: code, cause: cause, publicMessage: code.String()}
}

func NewWithMessage(code Code, cause error, publicMessage string) *Error {
	return &Error{code: code, cause: cause, publicMessage: publicMessage}
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.code, e.publicMessage, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.code, e.publicMessage)
}

func (e *Error) Code() Code             { return e.code }
func (e *Error) PublicMessage() string  { return e.publicMessage }
func (e *Error) HTTPStatus() int        { return e.code.HTTPStatus() }
func (e *Error) Unwrap() error          { return e.cause }

func Is(err error, target error) bool   { return errors.Is(err, target) }
func As(err error, target interface{}) bool { return errors.As(err, target) }
