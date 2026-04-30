package transport

import (
	"errors"
	"fmt"
	"net/http"
)

type APIError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func New(code int, message string) *APIError {
	if code < 400 || code > 599 {
		code = http.StatusInternalServerError
	}
	if message == "" {
		message = http.StatusText(code)
	}

	return &APIError{
		Code:    code,
		Message: message,
	}
}

func Wrap(code int, err error, message string) *APIError {
	apiErr := New(code, message)
	apiErr.Err = err
	return apiErr
}

func BadRequest(format string, args ...any) *APIError {
	return New(http.StatusBadRequest, fmt.Sprintf(format, args...))
}

func NotFound(format string, args ...any) *APIError {
	return New(http.StatusNotFound, fmt.Sprintf(format, args...))
}

func Internal(format string, args ...any) *APIError {
	return New(http.StatusInternalServerError, fmt.Sprintf(format, args...))
}

func WrapBadRequest(err error, message string) *APIError {
	return Wrap(http.StatusBadRequest, err, message)
}

func WrapInternal(err error, message string) *APIError {
	return Wrap(http.StatusInternalServerError, err, message)
}

func FromError(err error) *APIError {
	if err == nil {
		return Internal("internal server error")
	}

	var typed *APIError
	if errors.As(err, &typed) {
		return New(typed.Code, typed.Message).withCause(typed.Err)
	}

	return WrapInternal(err, "internal server error")
}

func (e *APIError) withCause(err error) *APIError {
	e.Err = err
	return e
}
