package middleware

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionRepository interface {
	GetSessionUser(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error)
	ExtendSession(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error
}
