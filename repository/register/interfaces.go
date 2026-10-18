package register

import (
	"context"
	"time"

	"app/models"

	"github.com/google/uuid"
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
