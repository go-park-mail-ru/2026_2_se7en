package http

import (
	"errors"
	"io"
	"net/http"

	models "github.com/go-park-mail-ru/2026_2_se7en/internal/media/model"
)

const MaxMultipartOverhead = 1 * 1024 * 1024

func ExtractAvatar(w http.ResponseWriter, r *http.Request) (*models.FileUpload, func(), error) {
	r.Body = http.MaxBytesReader(w, r.Body, models.MaxAvatarSize+MaxMultipartOverhead)
	if err := r.ParseMultipartForm(models.MaxAvatarSize + 1); err != nil {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
		return nil, nil, err
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		_ = file.Close()
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}

	if header.Size > models.MaxAvatarSize {
		cleanup()
		return nil, nil, errors.New("file too large")
	}

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		cleanup()
		return nil, nil, err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		cleanup()
		return nil, nil, err
	}

	contentType := http.DetectContentType(buffer[:n])
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		cleanup()
		return nil, nil, errors.New("invalid avatar type")
	}

	return &models.FileUpload{
		Reader:      file,
		Filename:    header.Filename,
		Size:        header.Size,
		ContentType: contentType,
	}, cleanup, nil
}
