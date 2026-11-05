package models

import (
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
