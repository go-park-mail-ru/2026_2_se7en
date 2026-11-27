package avatar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/infrastructure/http/observability/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/media/model"
	"github.com/google/uuid"
)

const AvatarDir = "./uploads/avatars"

type AvatarRepositoryImpl struct {
	log *slog.Logger
}

func NewAvatarRepository(baseLog *slog.Logger) *AvatarRepositoryImpl {
	if baseLog == nil {
		baseLog = slog.Default()
	}
	return &AvatarRepositoryImpl{
		log: logger.Scope(baseLog, logger.LayerRepository, logger.ModuleFile),
	}
}

func (r *AvatarRepositoryImpl) Save(ctx context.Context, file *models.FileUpload) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	if file == nil || file.Reader == nil || file.Size < 0 || file.Size > models.MaxAvatarSize {
		return "", "", errors.New("file too large")
	}

	var ext string
	switch file.ContentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		return "", "", errors.New("invalid avatar type")
	}

	if err := os.MkdirAll(AvatarDir, 0755); err != nil {
		r.log.ErrorContext(ctx, "SaveAvatar failed", logger.Error(err))
		return "", "", err
	}

	filename := uuid.NewString() + ext
	filePath := filepath.Join(AvatarDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		r.log.ErrorContext(ctx, "SaveAvatar failed", logger.Error(err))
		return "", "", err
	}

	written, copyErr := io.Copy(dst, io.LimitReader(file.Reader, models.MaxAvatarSize+1))
	closeErr := dst.Close()
	if copyErr != nil {
		_ = os.Remove(filePath)
		r.log.ErrorContext(ctx, "SaveAvatar failed", logger.Error(copyErr))
		return "", "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(filePath)
		r.log.ErrorContext(ctx, "CloseAvatar failed", logger.Error(closeErr))
		return "", "", closeErr
	}
	if written > models.MaxAvatarSize {
		_ = os.Remove(filePath)
		return "", "", errors.New("file too large")
	}

	url := fmt.Sprintf("/uploads/avatars/%s", filename)
	return url, filename, nil
}

func (r *AvatarRepositoryImpl) Remove(ctx context.Context, filename string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if filename == "" || filepath.Base(filename) != filename {
		return errors.New("invalid avatar filename")
	}
	return os.Remove(filepath.Join(AvatarDir, filename))
}
