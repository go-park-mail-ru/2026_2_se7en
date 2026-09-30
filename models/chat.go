package models

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID           uuid.UUID
	IconID       *uuid.UUID
	Name         string
	Description  *string
	Type         string
	MembersCount int
	LastMessage  *MessagePreview
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
