package profile

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	usecaseprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/profile"
	"github.com/google/uuid"
)

type ProfileUsecase interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, *models.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, params usecaseprofile.UpdateParams) error
	UploadAvatar(ctx context.Context, userID uuid.UUID, file *models.FileUpload) (string, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
}
