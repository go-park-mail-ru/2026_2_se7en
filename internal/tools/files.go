package tools

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

const (
	MaxAvatarSize = 5 * 1024 * 1024
	AvatarDir     = "./uploads/avatars"
)

func SaveAvatar(file *models.FileUpload) (string, string, error) {
	if file.Size > MaxAvatarSize {
		return "", "", errors.New("file too large")
	}

	if err := os.MkdirAll(AvatarDir, 0755); err != nil {
		return "", "", err
	}

	ext := filepath.Ext(file.Filename)
	filename := uuid.NewString() + ext
	filePath := filepath.Join(AvatarDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", err
	}

	defer dst.Close()

	if _, err := io.Copy(dst, file.Reader); err != nil {
		return "", "", err
	}

	url := fmt.Sprintf("/uploads/avatars/%s", filename)
	
	return url, filename, nil
}
