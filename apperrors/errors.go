package apperrors

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrUserNotFound = errors.New("user not found")
	// ErrSessionNotFound = errors.New("session not found")
)

type ErrorDetail struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type AppError struct {
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

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) ToResponse() *ErrorResponse {
	return &ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	}
}

func WriteError(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	_ = json.NewEncoder(w).Encode(err.ToResponse())
}

func NewBadRequest(msg string, details []ErrorDetail) *AppError {
	return &AppError{
		Code:    "BAD_REQUEST",
		Message: msg,
		Status:  http.StatusBadRequest,
		Details: details,
	}
}

func NewUnauthorized(msg string) *AppError {
	return &AppError{
		Code:    "UNAUTHORIZED",
		Message: msg,
		Status:  http.StatusUnauthorized,
	}
}

func NewForbidden(msg string) *AppError {
	return &AppError{
		Code:    "FORBIDDEN",
		Message: msg,
		Status:  http.StatusForbidden,
	}
}

func NewNotFound(msg string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: msg,
		Status:  http.StatusNotFound,
	}
}

func NewConflict(msg string, details []ErrorDetail) *AppError {
	return NewConflictWithCode("CONFLICT", msg, details)
}

func NewConflictWithCode(code, msg string, details []ErrorDetail) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Status:  http.StatusConflict,
		Details: details,
	}
}

func NewInternalError(msg string, err error) *AppError {
	return &AppError{
		Code:    "INTERNAL_ERROR",
		Message: msg,
		Status:  http.StatusInternalServerError,
		Err:     err,
	}
}

var (
	ErrSessionNotFound = &AppError{
		Code:    "SESSION_NOT_FOUND",
		Message: "Session not found",
		Status:  http.StatusUnauthorized,
	}
	ErrEmailTaken = &AppError{
		Code:    "EMAIL_TAKEN",
		Message: "Email already taken",
		Status:  http.StatusConflict,
	}
	ErrNicknameTaken = &AppError{
		Code:    "NICKNAME_TAKEN",
		Message: "Nickname already taken",
		Status:  http.StatusConflict,
	}
)
