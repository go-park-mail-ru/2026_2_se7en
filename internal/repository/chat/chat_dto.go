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

type ChatDetailDB struct {
	Chat ChatDB
	Role string
}

type ChatMemberDB struct {
	UserID  uuid.UUID
	Role    string
	Profile MemberProfileDB
}

type MemberProfileDB struct {
	ID        uuid.UUID
	IconID    *uuid.UUID
	UserID    uuid.UUID
	Nickname  string
	FirstName string
	LastName  *string
	Bio       *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MessageDB struct {
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
