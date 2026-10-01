package models

import (
	"time"

	"github.com/google/uuid"
)

type MessagePreview struct {
	ID        uuid.UUID `json:"id"`
	Content   *string   `json:"content,omitempty"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}
