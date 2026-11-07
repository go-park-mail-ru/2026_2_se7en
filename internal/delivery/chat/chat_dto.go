package chat

import (
	"time"

	"github.com/google/uuid"
)

type GetOrCreateDialogRequest struct {
	PeerID uuid.UUID `json:"peer_id" format:"uuid"`
}

type CreateGroupRequest struct {
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Members     []uuid.UUID `json:"members"`
}

type UpdateGroupRequest struct {
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	SetDescription bool    `json:"set_description"`
}

type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" format:"uuid"`
}

type CreateMessageRequest struct {
	ReplyToMessageID *uuid.UUID `json:"reply_to_message_id,omitempty" format:"uuid"`
	StickerID        *uuid.UUID `json:"sticker_id,omitempty" format:"uuid"`
	Content          *string    `json:"content,omitempty"`
	Type             string     `json:"type" `
}

type MessagePreviewResponse struct {
	ID        uuid.UUID `json:"id" format:"uuid"`
	Content   *string   `json:"content,omitempty"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
}

type ChatItemResponse struct {
	ID           uuid.UUID               `json:"id" format:"uuid"`
	IconID       *uuid.UUID              `json:"icon_id,omitempty" format:"uuid"`
	Name         string                  `json:"name"`
	Description  *string                 `json:"description,omitempty"`
	Type         string                  `json:"type"`
	MembersCount int                     `json:"members_count"`
	CreatedAt    time.Time               `json:"created_at" format:"date-time"`
	UpdatedAt    time.Time               `json:"updated_at" format:"date-time"`
	LastMessage  *MessagePreviewResponse `json:"last_message,omitempty"`
}

type ListChatsResponse struct {
	Items []ChatItemResponse `json:"items"`
}

type ChatResponse struct {
	ID           uuid.UUID  `json:"id" format:"uuid"`
	IconID       *uuid.UUID `json:"icon_id,omitempty" format:"uuid"`
	Name         string     `json:"name"`
	Description  *string    `json:"description,omitempty"`
	Type         string     `json:"type"`
	MembersCount int        `json:"members_count"`
	CreatedAt    time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt    time.Time  `json:"updated_at" format:"date-time"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" format:"date-time"`
}

type ChatDetailResponse struct {
	Chat *ChatResponse `json:"chat"`
	Role string        `json:"role"`
}

type MemberProfileResponse struct {
	ID        uuid.UUID  `json:"id" format:"uuid"`
	IconID    *uuid.UUID `json:"icon_id,omitempty" format:"uuid"`
	UserID    uuid.UUID  `json:"user_id" format:"uuid"`
	Nickname  string     `json:"nickname"`
	FirstName string     `json:"first_name"`
	LastName  *string    `json:"last_name,omitempty"`
	Bio       *string    `json:"bio,omitempty"`
	CreatedAt time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt time.Time  `json:"updated_at" format:"date-time"`
}

type ChatMemberResponse struct {
	UserID  uuid.UUID              `json:"user_id" format:"uuid"`
	Role    string                 `json:"role"`
	Profile *MemberProfileResponse `json:"profile"`
}

type ListMembersResponse struct {
	Items []ChatMemberResponse `json:"items"`
}

type MessageResponse struct {
	ID               uuid.UUID  `json:"id" format:"uuid"`
	UserID           uuid.UUID  `json:"user_id" format:"uuid"`
	ChatID           uuid.UUID  `json:"chat_id" format:"uuid"`
	ReplyToMessageID *uuid.UUID `json:"reply_to_message_id,omitempty" format:"uuid"`
	StickerID        *uuid.UUID `json:"sticker_id,omitempty" format:"uuid"`
	Content          *string    `json:"content,omitempty"`
	IsEdited         bool       `json:"is_edited"`
	IsRead           bool       `json:"is_read"`
	Type             string     `json:"type"`
	CreatedAt        time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt        time.Time  `json:"updated_at" format:"date-time"`
}

type ListMessagesResponse struct {
	Items []MessageResponse `json:"items"`
}
