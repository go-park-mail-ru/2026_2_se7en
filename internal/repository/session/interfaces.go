package session

import (
	"app/internal/models"
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error)
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
	GetSessionUser(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error)
	ExtendSession(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error
}
