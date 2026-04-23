package validate

import (
	"encoding/json"
	"errors"
	"net/http"
)

type UnauthorizedError struct {
}

func (err *UnauthorizedError) Error() string {
	return "unauthorized"
}

func IsUnauthorizedError(err error) bool {
	var re *UnauthorizedError
	return errors.As(err, &re)
}

func NewUnauthorizedError() error {
	return &UnauthorizedError{}
}

type InvalidRouteKeyError struct {
	key string
}

func (err *InvalidRouteKeyError) Error() string {
	return "invalid route key: " + err.key
}

func NewRouteKeyError(key string) error {
	return &InvalidRouteKeyError{key}
}

func IsInvalidRouteKeyError(err error) bool {
	var re *InvalidRouteKeyError
	return errors.As(err, &re)
}

// ErrorResponse is the form used for API responses from failures in the API.
type ErrorResponse struct {
	Error  string `json:"error"`
	Fields string `json:"fields,omitempty"`
}

// RequestError is used to pass an error during the request through the
// application with web specific context.
type RequestError struct {
	Err    error
	Status int
	Fields error
}

// NewRequestError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewRequestError(err error, status int) error {
	return &RequestError{err, status, nil}
}

func (err *RequestError) Error() string {
	if err.Err == nil {
		if err.Status != 0 {
			return http.StatusText(err.Status)
		}
		return "request error"
	}
	return err.Err.Error()
}

// IsRequestError checks if an error of type RequestError exists.
func IsRequestError(err error) bool {
	var re *RequestError
	return errors.As(err, &re)
}

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// FieldErrors represents a collection of field errors.
type FieldErrors []FieldError

func (fe FieldErrors) Append(field, reason string) FieldErrors {
	return append(fe, FieldError{
		Field: field,
		Error: reason,
	})
}

func (fe FieldErrors) Nil() bool {
	return len(fe) == 0
}

// Error implements the error interface.
func (fe FieldErrors) Error() string {
	d, err := json.Marshal(fe)
	if err != nil {
		return err.Error()
	}
	return string(d)
}

func NewFieldErrors(errs ...FieldError) FieldErrors {
	return errs
}

func NewFieldError(field, reason string) FieldError {
	return FieldError{Field: field, Error: reason}
}

func IsFieldError(err error) bool {
	v := FieldErrors{}
	return errors.As(err, &v)
}

// Cause iterates through all the wrapped errors until the root
// error value is reached.
func Cause(err error) error {
	root := err
	for {
		if err = errors.Unwrap(root); err == nil {
			return root
		}
		root = err
	}
}
