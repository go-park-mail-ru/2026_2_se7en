package middleware

import (
	"app/apperrors"
	"app/utils"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const SessionExtensionDays = 30

type ErrorResponse struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type SessionDB interface {
	GetSessionUser(ctx context.Context, sessionID string) (uuid.UUID, error)
	ExtendSession(ctx context.Context, sessionID string, newExpiresAt time.Time) error
}

type AuthMiddleware struct {
	db   SessionDB
	next http.Handler
}

func NewAuthMiddleware(db SessionDB, next http.Handler) http.Handler {
	return &AuthMiddleware{
		db:   db,
		next: next,
	}
}

func (m *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		m.Unauthorized(w, "Session cookie is missing", nil)
		return
	}

	userID, err := m.db.GetSessionUser(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, apperrors.ErrSessionNotFound) {
			m.Unauthorized(w, "Invalid or expired session", nil)
		} else {
			m.ServerError(w, "Internal server error", nil)
		}
		return
	}

	newExpiresAt := time.Now().Add(SessionExtensionDays * 24 * time.Hour)

	errWarn := m.db.ExtendSession(r.Context(), cookie.Value, newExpiresAt)
	if errWarn != nil {
		// TODO тоже запись в логгер, как я думаю
	}

	utils.SetSessionCookie(w, cookie.Value)

	authCtx := NewAuthContext(r.Context(), userID)

	m.next.ServeHTTP(w, r.WithContext(authCtx))
}

func (m *AuthMiddleware) writeError(w http.ResponseWriter, status int, code, message string, details []ErrorDetail) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})

	if err != nil {
		// TODO запись в логгер
	}
}

func (m *AuthMiddleware) Unauthorized(w http.ResponseWriter, msg string, details []ErrorDetail) {
	m.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", msg, details)
}

func (m *AuthMiddleware) ServerError(w http.ResponseWriter, msg string, details []ErrorDetail) {
	m.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", msg, details)
}

func (m *AuthMiddleware) BadRequest(w http.ResponseWriter, msg string, details []ErrorDetail) {
	m.writeError(w, http.StatusBadRequest, "BAD_REQUEST", msg, details)
}
