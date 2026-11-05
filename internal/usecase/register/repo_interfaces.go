package register

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
)

type UserRepository interface {
	CheckEmailExist(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, email, passwordHash string, phone *string) (*models.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

type ProfileRepository interface {
	CheckNicknameExist(ctx context.Context, nickname string) (bool, error)
	CreateProfile(ctx context.Context, iconID *uuid.UUID, userID uuid.UUID, nickname, firstName string, lastName *string, bio *string) (*models.Profile, error)
}

type SessionRepository interface {
	CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error)
}
