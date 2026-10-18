package profile

import (
	"app/models"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	CheckNicknameExist(ctx context.Context, nickname string) (bool, error)
	CreateProfile(ctx context.Context, iconID *uuid.UUID, userID uuid.UUID, nickname, firstName string, lastName *string, bio *string) (*models.Profile, error)
}
