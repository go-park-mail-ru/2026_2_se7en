package profile

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/auth/model"
	usecaseprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/auth/usecase/profile"
	"github.com/google/uuid"

	mediamodel "github.com/go-park-mail-ru/2026_2_se7en/internal/media/model"
)

type ProfileUsecase interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, *models.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, params usecaseprofile.UpdateParams) error
	UploadAvatar(ctx context.Context, userID uuid.UUID, file *mediamodel.FileUpload) (*mediamodel.File, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
}
