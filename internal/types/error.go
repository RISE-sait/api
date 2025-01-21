package types

import "net/http"

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func (e *NotFoundError) StatusCode() int {
	return http.StatusNotFound
}

type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	return e.Message
}

func (e *InternalError) StatusCode() int {
	return http.StatusInternalServerError
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) StatusCode() int {
	return http.StatusBadRequest
}
