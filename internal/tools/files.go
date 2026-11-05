package tools

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const (
	MaxAvatarSize = 5 * 1024 * 1024
	AvatarDir     = "./uploads/avatars"
)

func SaveAvatar(file multipart.File, header *multipart.FileHeader) (string, string, error) {
	if header.Size > MaxAvatarSize {
		return "", "", errors.New("file too large")
	}

	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "", "", err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", err
	}

	contentType := http.DetectContentType(buffer)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return "", "", errors.New("invalid file type")
	}

	if err := os.MkdirAll(AvatarDir, 0755); err != nil {
		return "", "", err
	}

	ext := filepath.Ext(header.Filename)
	filename := uuid.NewString() + ext
	filePath := filepath.Join(AvatarDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", err
	}

	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", "", err
	}

	url := fmt.Sprintf("/uploads/avatars/%s", filename)
	
	return url, filename, nil
}
