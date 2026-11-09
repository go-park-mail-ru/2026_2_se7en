package tools

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

const (
	MaxAvatarSize        = 5 * 1024 * 1024
	MaxMultipartOverhead = 1 * 1024 * 1024
	AvatarDir            = "./uploads/avatars"
)

func SaveAvatar(file *models.FileUpload) (string, string, error) {
	if file == nil || file.Reader == nil || file.Size < 0 || file.Size > MaxAvatarSize {
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
		return "", "", err
	}

	filename := uuid.NewString() + ext
	filePath := filepath.Join(AvatarDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", err
	}

	written, err := io.Copy(dst, io.LimitReader(file.Reader, MaxAvatarSize+1))
	if err != nil || written > MaxAvatarSize {
		_ = dst.Close()
		_ = os.Remove(filePath)
		if err == nil {
			err = errors.New("file too large")
		}
		return "", "", err
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(filePath)
		return "", "", err
	}

	url := fmt.Sprintf("/uploads/avatars/%s", filename)

	return url, filename, nil
}

func RemoveAvatar(filename string) error {
	if filename == "" || filepath.Base(filename) != filename {
		return errors.New("invalid avatar filename")
	}
	return os.Remove(filepath.Join(AvatarDir, filename))
}

func ExtractAvatar(w http.ResponseWriter, r *http.Request) (*models.FileUpload, func(), error) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxAvatarSize+MaxMultipartOverhead)
	if err := r.ParseMultipartForm(MaxAvatarSize + 1); err != nil {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
		return nil, nil, err
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, nil, err
	}
	closeFile := func() {
		_ = file.Close()
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}
	if header.Size > MaxAvatarSize {
		closeFile()
		return nil, nil, errors.New("file too large")
	}
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		closeFile()
		return nil, nil, err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		closeFile()
		return nil, nil, err
	}
	contentType := http.DetectContentType(buffer[:n])
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		closeFile()
		return nil, nil, errors.New("invalid avatar type")
	}
	return &models.FileUpload{Reader: file, Filename: header.Filename, Size: header.Size, ContentType: contentType}, closeFile, nil
}
