package models

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID           uuid.UUID       `json:"id"`
	IconID       *uuid.UUID      `json:"icon_id,omitempty"`
	Name         string          `json:"name"`
	Description  *string         `json:"description,omitempty"`
	Type         string          `json:"type"`
	MembersCount int             `json:"members_count"`
	LastMessage  *MessagePreview `json:"last_message,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"-"`
	DeletedAt    *time.Time      `json:"-"`
}
