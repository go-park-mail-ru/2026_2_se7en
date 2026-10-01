package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const UserContextKey contextKey = "user_id"

type ErrorResponse struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type SesssionDB interface {
	GetSessionUser(ctx context.Context, sessionID string) (uuid.UUID, error)
}

type AuthMiddleware struct {
	db   SesssionDB
	next http.Handler
}

func NewAuthMiddleware(db SesssionDB, next http.Handler) http.Handler {
	return &AuthMiddleware{
		db:   db,
		next: next,
	}
}

func (m *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		m.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Session cookie is missing", nil)
		return
	}

	userID, err := m.db.GetSessionUser(r.Context(), cookie.Value)
	if err != nil {
		m.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired session", nil)
		return
	}

	ctx := context.WithValue(r.Context(), UserContextKey, userID)

	m.next.ServeHTTP(w, r.WithContext(ctx))
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserContextKey).(uuid.UUID)
	return userID, ok
}

func (m *AuthMiddleware) writeError(w http.ResponseWriter, status int, code, message string, details []ErrorDetail) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}
