package middleware

import (
	"app/apperrors"
	"app/utils"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const SessionExtensionDays = 30

type SessionDB interface {
	GetSessionUser(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error)
	ExtendSession(ctx context.Context, sessionID uuid.UUID, newExpiresAt time.Time) error
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
	cookie, err := r.Cookie(utils.SessionCookieName)
	if err != nil || cookie.Value == "" {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Session cookie is missing"))
		return
	}
	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid session cookie"))
		return
	}

	userID, err := m.db.GetSessionUser(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, apperrors.ErrSessionNotFound) {
			apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid or expired session"))
		} else {
			apperrors.WriteError(w, apperrors.NewInternalError("Internal server error", err))
		}
		return
	}

	newExpiresAt := time.Now().Add(SessionExtensionDays * 24 * time.Hour)
	if err := m.db.ExtendSession(r.Context(), sessionID, newExpiresAt); err != nil {
		if errors.Is(err, apperrors.ErrSessionNotFound) {
			apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid or expired session"))
		} else {
			apperrors.WriteError(w, apperrors.NewInternalError("Internal server error", err))
		}
		return
	}

	utils.SetSessionCookie(w, cookie.Value)

	authCtx := NewAuthContext(r.Context(), userID)

	m.next.ServeHTTP(w, r.WithContext(authCtx))
}
