package models

import (
	"time"

	"github.com/google/uuid"
)

type MessagePreview struct {
	ID        *uuid.UUID
	Content   *string
	Type      string
	CreatedAt *time.Time
}
