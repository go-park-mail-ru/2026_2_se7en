package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"
const sessionIDKey contextKey = "sessionID"

type AuthMiddleware struct {
	log *slog.Logger

	sessionRepo SessionRepository
	next        http.Handler
}

func NewAuthMiddleware(sessionRepo SessionRepository, next http.Handler, baseLog *slog.Logger) http.Handler {
	return &AuthMiddleware{
		log: logger.Scope(baseLog, logger.LayerMiddleware, logger.ModuleAuth),

		sessionRepo: sessionRepo,
		next:        next,
	}
}

func (m *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		m.log.DebugContext(r.Context(), "session cookie missing")

		httperrors.WriteError(w, httperrors.NewUnauthorized("Session cookie is missing"))
		return
	}

	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		m.log.WarnContext(r.Context(), "invalid session cookie")
		httperrors.WriteError(w, httperrors.NewUnauthorized("Invalid session cookie"))
		return
	}

	sessionUser, err := m.sessionRepo.GetSessionUser(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, apperrors.ErrSessionNotFound) {
			m.log.WarnContext(r.Context(), "session expired or not found")
			httperrors.WriteError(w, httperrors.NewUnauthorized("Invalid or expired session"))
		} else {
			m.log.ErrorContext(r.Context(), "session lookup failed", logger.Error(err))
			httperrors.WriteError(w, httperrors.NewInternalError("Internal server error", err))
		}
		return
	}
	if sessionUser == nil {
		err := apperrors.ErrSessionUserMissing
		m.log.ErrorContext(r.Context(), "session lookup failed", logger.Error(err))
		httperrors.WriteError(w, httperrors.NewInternalError("Internal server error", err))
		return
	}
	userID := sessionUser.UserID

	newExpiresAt := time.Now().Add(SessionTTL)
	if err := m.sessionRepo.ExtendSession(r.Context(), sessionID, newExpiresAt); err != nil {
		if errors.Is(err, apperrors.ErrSessionNotFound) {
			m.log.WarnContext(r.Context(), "session expired or not found")
			httperrors.WriteError(w, httperrors.NewUnauthorized("Invalid or expired session"))
		} else {
			m.log.ErrorContext(r.Context(), "session lookup failed", logger.Error(err))
			httperrors.WriteError(w, httperrors.NewInternalError("Internal server error", err))
		}
		return
	}

	SetSessionCookie(w, r, cookie.Value)

	authCtx := NewAuthContext(NewSessionContext(r.Context(), sessionID), userID)

	authCtx = logger.WithUserID(authCtx, userID.String())

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
