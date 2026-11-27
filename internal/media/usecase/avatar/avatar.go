package avatar

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/infrastructure/http/observability/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/media/model"
)

type AvatarUsecase interface {
	UploadAvatar(ctx context.Context, file *models.FileUpload) (url string, filename string, err error)
	RemoveAvatar(ctx context.Context, filename string) error
}

type AvatarUsecaseImpl struct {
	log *slog.Logger

	repo AvatarRepository
}

func NewAvatarUsecase(repo AvatarRepository, baseLog *slog.Logger) *AvatarUsecaseImpl {
	return &AvatarUsecaseImpl{
		log:  logger.Scope(baseLog, logger.LayerUsecase, logger.ModuleFile),
		repo: repo,
	}
}

func (uc *AvatarUsecaseImpl) UploadAvatar(ctx context.Context, file *models.FileUpload) (string, string, error) {
	if file == nil || file.Reader == nil {
		return "", "", errors.New("invalid avatar")
	}
	if file.Size < 0 || file.Size > models.MaxAvatarSize {
		return "", "", errors.New("file too large")
	}
	switch file.ContentType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return "", "", errors.New("invalid avatar type")
	}

	url, filename, err := uc.repo.Save(ctx, file)
	if err != nil {
		uc.log.ErrorContext(ctx, "UploadAvatar failed", logger.Error(err))
		return "", "", err
	}
	return url, filename, nil
}

func (uc *AvatarUsecaseImpl) RemoveAvatar(ctx context.Context, filename string) error {
	if err := uc.repo.Remove(ctx, filename); err != nil {
		uc.log.ErrorContext(ctx, "RemoveAvatar failed", logger.Error(err))
		return err
	}
	return nil
}
