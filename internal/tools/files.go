package tools

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
)

const (
	MaxAvatarSize = 5 * 1024 * 1024
	AvatarDir     = "./uploads/avatars"
)

func SaveAvatar(ctx context.Context, file *models.FileUpload, log *slog.Logger) (string, string, error) {
	if file.Size > MaxAvatarSize {
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
		return "", "", errors.New("invalid file type")
	}

	if err := os.MkdirAll(AvatarDir, 0755); err != nil {
		log.ErrorContext(ctx, "SaveAvatar failed", logger.Error(err))

		return "", "", err
	}

	filename := uuid.NewString() + ext
	filePath := filepath.Join(AvatarDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		log.ErrorContext(ctx, "SaveAvatar failed", logger.Error(err))

		return "", "", err
	}

	defer func() {
		if err := dst.Close(); err != nil {
			log.ErrorContext(ctx, "CloseAvatar failed", logger.Error(err))
		}
	}()

	if _, err := io.Copy(dst, file.Reader); err != nil {
		log.ErrorContext(ctx, "SaveAvatar failed", logger.Error(err))

		return "", "", err
	}

	url := fmt.Sprintf("/uploads/avatars/%s", filename)

	return url, filename, nil
}
