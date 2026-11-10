package chat

import (
	"github.com/google/uuid"
	"time"
)

type CreateMessageRequest struct {
	ReplyToMessageID *uuid.UUID `json:"reply_to_message_id,omitempty" format:"uuid"`
	StickerID        *uuid.UUID `json:"sticker_id,omitempty" format:"uuid"`
	Content          *string    `json:"content,omitempty"`
	Type             string     `json:"type" `
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
