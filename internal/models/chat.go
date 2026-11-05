package models

import (
	"time"

	"github.com/google/uuid"
)

type MessagePreview struct {
	ID        uuid.UUID
	Content   *string
	Type      string
	CreatedAt time.Time
}

type Chat struct {
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

type ChatListItem struct {
	Chat        *Chat
	LastMessage *MessagePreview
}

type ChatDetail struct {
	Chat *Chat
	Role string
}

type ChatMember struct {
	UserID  uuid.UUID
	Role    string
	Profile *Profile
}

type Message struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	ChatID           uuid.UUID
	ReplyToMessageID *uuid.UUID
	StickerID        *uuid.UUID
	Content          *string
	IsEdited         bool
	Type             string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
