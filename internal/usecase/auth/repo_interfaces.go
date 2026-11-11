package auth

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

type ProfileRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
}

type SessionRepository interface {
	CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error)
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
}
