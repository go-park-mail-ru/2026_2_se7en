package profile

import (
	"context"
	"mime/multipart"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	repoprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
	usecaseprofile "github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/profile"
	"github.com/google/uuid"
)

type ProfileUsecase interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, *repoprofile.ProfileWithFileDB, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, params usecaseprofile.UpdateParams) error
	UploadAvatar(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (string, error)
}
