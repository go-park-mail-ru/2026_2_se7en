package chat

import (
	"time"

	"github.com/google/uuid"
)

type ChatDB struct {
	ID           uuid.UUID
	IconID       *uuid.UUID
	Name         string
	Description  *string
	Type         string
	MembersCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type MessagePreviewDB struct {
	ID        uuid.UUID
	Content   *string
	Type      string
	CreatedAt time.Time
}

type ChatListItemDB struct {
	Chat        ChatDB
	LastMessage *MessagePreviewDB
}
