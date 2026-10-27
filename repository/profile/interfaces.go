package profile

import (
	"context"

	"app/models"

	"github.com/google/uuid"
)

type Repository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	CheckNicknameExist(ctx context.Context, nickname string) (bool, error)
	CreateProfile(ctx context.Context, iconID *uuid.UUID, userID uuid.UUID, nickname, firstName string, lastName *string, bio *string) (*models.Profile, error)
	FindProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error)
	FindPasswordHash(ctx context.Context, userID uuid.UUID) (string, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, oldHash, newHash string) (bool, error)
	UpdatePhone(ctx context.Context, userID uuid.UUID, phone *string) error
}
