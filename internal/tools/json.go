package tools

import (
	"encoding/json"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
)

type Validatable interface {
	IsValid() []apperrors.ErrorDetail
}

type Handler[T any] func(w http.ResponseWriter, r *http.Request, req *T) *apperrors.AppError

type JSONDecoder[T any] struct {
	handler Handler[T]
}

func (d *JSONDecoder[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req T

	if r.Body == nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Request body is empty", nil))
		return
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid JSON format", nil))
		return
	}

	if v, ok := any(&req).(Validatable); ok {
		if details := v.IsValid(); len(details) > 0 {
			apperrors.WriteError(w, apperrors.NewBadRequest("Validation failed", details))
			return
		}
	}

	if appErr := d.handler(w, r, &req); appErr != nil {
		apperrors.WriteError(w, appErr)
	}
}

func NewJSONDecoder[T any](handler Handler[T]) http.Handler {
	return &JSONDecoder[T]{
		handler: handler,
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
