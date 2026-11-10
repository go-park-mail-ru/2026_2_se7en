package middleware

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type SessionRepository interface {
	GetSessionUser(ctx context.Context, sessionID uuid.UUID) (*models.SessionUser, error)
	ExtendSession(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error
}
