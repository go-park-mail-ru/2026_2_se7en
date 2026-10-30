package chat

import (
	"time"

	"github.com/google/uuid"
)

type MessagePreviewDTO struct {
	ID        uuid.UUID `json:"id" format:"uuid"`
	Content   *string   `json:"content,omitempty"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
}

type ChatItemDTO struct {
	ID           uuid.UUID          `json:"id" format:"uuid"`
	IconID       *uuid.UUID         `json:"icon_id,omitempty" format:"uuid"`
	Name         string             `json:"name"`
	Description  *string            `json:"description,omitempty"`
	Type         string             `json:"type"`
	MembersCount int                `json:"members_count"`
	CreatedAt    time.Time          `json:"created_at" format:"date-time"`
	UpdatedAt    time.Time          `json:"updated_at" format:"date-time"`
	LastMessage  *MessagePreviewDTO `json:"last_message,omitempty"`
}

type ListChatsResponse struct {
	Items []ChatItemDTO `json:"items"`
}
