package chat

import (
	"github.com/google/uuid"
	"time"
)

type MessagePreviewResponse struct {
	ID        uuid.UUID `json:"id" format:"uuid"`
	Content   *string   `json:"content,omitempty"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
}

type ChatItemResponse struct {
	ID           uuid.UUID               `json:"id" format:"uuid"`
	IconID       *uuid.UUID              `json:"icon_id,omitempty" format:"uuid"`
	IconURL      *string                 `json:"icon_url,omitempty"`
	IsOnline     *bool                   `json:"is_online,omitempty"`
	Name         string                  `json:"name"`
	Description  *string                 `json:"description,omitempty"`
	Type         string                  `json:"type"`
	MembersCount int                     `json:"members_count"`
	CreatedAt    time.Time               `json:"created_at" format:"date-time"`
	UpdatedAt    time.Time               `json:"updated_at" format:"date-time"`
	LastMessage  *MessagePreviewResponse `json:"last_message,omitempty"`
	UnreadCount  int                     `json:"unread_count"`
}

type ListChatsResponse struct {
	Items         []ChatItemResponse `json:"items"`
	NextAfterTime *time.Time         `json:"next_after_time,omitempty" format:"date-time"`
	NextAfterID   *uuid.UUID         `json:"next_after_id,omitempty" format:"uuid"`
}

type ChatResponse struct {
	ID           uuid.UUID  `json:"id" format:"uuid"`
	IconID       *uuid.UUID `json:"icon_id,omitempty" format:"uuid"`
	IconURL      *string    `json:"icon_url,omitempty"`
	IsOnline     *bool      `json:"is_online,omitempty"`
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
