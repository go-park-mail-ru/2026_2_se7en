package chat

import (
	"time"

	"github.com/google/uuid"
)

type MessagePreviewDTO struct {
	ID        uuid.UUID `json:"id"`
	Content   *string   `json:"content,omitempty"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatItemDTO struct {
	ID           uuid.UUID          `json:"id"`
	IconID       *uuid.UUID         `json:"icon_id,omitempty"`
	Name         string             `json:"name"`
	Description  *string            `json:"description,omitempty"`
	Type         string             `json:"type"`
	MembersCount int                `json:"members_count"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	LastMessage  *MessagePreviewDTO `json:"last_message,omitempty"`
}

type ListChatsResponse struct {
	Items []ChatItemDTO `json:"items"`
}
