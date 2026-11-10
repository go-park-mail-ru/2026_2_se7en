package httperrors

import (
	"encoding/json"
	"net/http"
)

type ErrorDetail struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type HTTPError struct {
	Code    string
	Message string
	Status  int
	Details []ErrorDetail
	Err     error
}

type ErrorResponse struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func (e *HTTPError) ToResponse() *ErrorResponse {
	return &ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	}
}

func WriteError(w http.ResponseWriter, err *HTTPError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	_ = json.NewEncoder(w).Encode(err.ToResponse())
}

func NewBadRequest(msg string, details []ErrorDetail) *HTTPError {
	return &HTTPError{
		Code:    "BAD_REQUEST",
		Message: msg,
		Status:  http.StatusBadRequest,
		Details: details,
	}
}

func NewUnauthorized(msg string) *HTTPError {
	return &HTTPError{
		Code:    "UNAUTHORIZED",
		Message: msg,
		Status:  http.StatusUnauthorized,
	}
}

func NewForbidden(msg string) *HTTPError {
	return &HTTPError{
		Code:    "FORBIDDEN",
		Message: msg,
		Status:  http.StatusForbidden,
	}
}

func NewNotFound(msg string) *HTTPError {
	return &HTTPError{
		Code:    "NOT_FOUND",
		Message: msg,
		Status:  http.StatusNotFound,
	}
}

func NewConflict(msg string, details []ErrorDetail) *HTTPError {
	return NewConflictWithCode("CONFLICT", msg, details)
}

func NewConflictWithCode(code, msg string, details []ErrorDetail) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: msg,
		Status:  http.StatusConflict,
		Details: details,
	}
}

func NewInternalError(msg string, err error) *HTTPError {
	return &HTTPError{
		Code:    "INTERNAL_ERROR",
		Message: msg,
		Status:  http.StatusInternalServerError,
		Err:     err,
	}
}

func (e *HTTPError) ErrorCode() string {
	return e.Code
}

func (e *HTTPError) HTTPStatus() int {
	return e.Status
}
