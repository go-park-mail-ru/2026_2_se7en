package profile

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type UserRepository interface {
	FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateEmail(ctx context.Context, userID uuid.UUID, email string) error
	UpdatePhone(ctx context.Context, userID uuid.UUID, phone *string) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

type ProfileRepository interface {
	GetWithFileByUserID(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	Update(ctx context.Context, profileID uuid.UUID, nickname, firstName string, lastName *string, bio *string) error
	UpdateIcon(ctx context.Context, profileID uuid.UUID, fileID uuid.UUID) error
}

type FileRepository interface {
	Create(ctx context.Context, url, filename, mimeType string) (*models.File, error)
}
