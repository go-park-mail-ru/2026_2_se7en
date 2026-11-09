package tools

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
)

type Validatable interface {
	IsValid() []apperrors.ErrorDetail
}

type Handler[T any] func(w http.ResponseWriter, r *http.Request, req *T) *apperrors.AppError

type JSONDecoder[T any] struct {
	log *slog.Logger

	handler Handler[T]
}

func (d *JSONDecoder[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req T

	if r.Body == nil {
		apperrors.WriteError(w, apperrors.NewBadRequest("Request body is empty", nil))
		return
	}
	defer func() { _ = r.Body.Close() }()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		d.log.WarnContext(r.Context(), "invalid request JSON", logger.Error(err))

		apperrors.WriteError(w, apperrors.NewBadRequest("Invalid JSON format", nil))
		return
	}

	if v, ok := any(&req).(Validatable); ok {
		if details := v.IsValid(); len(details) > 0 {
			d.log.WarnContext(r.Context(), "request validation failed")
			apperrors.WriteError(w, apperrors.NewBadRequest("Validation failed", details))
			return
		}
	}

	if appErr := d.handler(w, r, &req); appErr != nil {
		level := slog.LevelError
		if appErr.Status < http.StatusInternalServerError {
			level = slog.LevelWarn
		}
		d.log.LogAttrs(r.Context(), level, "request failed", logger.Error(appErr))

		apperrors.WriteError(w, appErr)
	}
}

func NewJSONDecoder[T any](handler Handler[T], baseLog *slog.Logger) http.Handler {
	return &JSONDecoder[T]{
		log: logger.Scope(baseLog, logger.LayerTools, logger.ModuleJSON),

		handler: handler,
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
