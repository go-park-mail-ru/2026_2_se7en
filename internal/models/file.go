package models

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID        uuid.UUID
	URL       string
	Filename  string
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FileUpload struct {
	Reader      io.Reader
	Filename    string
	Size        int64
	ContentType string
}
