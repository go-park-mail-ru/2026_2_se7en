package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
)

type contextKey string

const userIDKey contextKey = "userID"
const sessionIDKey contextKey = "sessionID"

type AuthMiddleware struct {
	sessionRepo SessionRepository
	next        http.Handler
}

func NewAuthMiddleware(sessionRepo SessionRepository, next http.Handler) http.Handler {
	return &AuthMiddleware{
		sessionRepo: sessionRepo,
		next:        next,
	}
}

func (m *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Session cookie is missing"))
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid session cookie"))
		return
	}

	userID, err := m.sessionRepo.GetSessionUser(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, apperrors.ErrSessionNotFound) {
			apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid or expired session"))
		} else {
			apperrors.WriteError(w, apperrors.NewInternalError("Internal server error", err))
		}
		return
	}

	newExpiresAt := time.Now().Add(SessionTTL)
	if err := m.sessionRepo.ExtendSession(r.Context(), sessionID, newExpiresAt); err != nil {
		if errors.Is(err, apperrors.ErrSessionNotFound) {
			apperrors.WriteError(w, apperrors.NewUnauthorized("Invalid or expired session"))
		} else {
			apperrors.WriteError(w, apperrors.NewInternalError("Internal server error", err))
		}
		return
	}

	SetSessionCookie(w, r, cookie.Value)

	authCtx := NewAuthContext(NewSessionContext(r.Context(), sessionID), userID)

	m.next.ServeHTTP(w, r.WithContext(authCtx))
}

func NewAuthContext(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

func NewSessionContext(ctx context.Context, sessionID uuid.UUID) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

func SessionIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(uuid.UUID)
	return sessionID, ok
}
