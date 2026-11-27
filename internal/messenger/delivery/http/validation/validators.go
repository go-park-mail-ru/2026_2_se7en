package validation

import (
	"strings"
	"unicode/utf8"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/infrastructure/http/httperrors"
	"github.com/google/uuid"
)

const (
	MaxChatNameLen        = 100
	MaxChatDescriptionLen = 500
	MaxMessageContentLen  = 4096
)

func ValidateChatName(name string) *httperrors.ErrorDetail {
	if strings.TrimSpace(name) == "" {
		return &httperrors.ErrorDetail{Field: "name", Reason: "Name is required"}
	}
	if utf8.RuneCountInString(name) > MaxChatNameLen {
		return &httperrors.ErrorDetail{Field: "name", Reason: "Name must be at most 100 characters"}
	}
	return nil
}

func ValidateChatDescription(description *string) *httperrors.ErrorDetail {
	if description != nil && utf8.RuneCountInString(*description) > MaxChatDescriptionLen {
		return &httperrors.ErrorDetail{Field: "description", Reason: "Description must be at most 500 characters"}
	}
	return nil
}

func ValidateMessageFields(messageType string, content *string, stickerID *uuid.UUID) *httperrors.ErrorDetail {
	switch messageType {
	case "text":
		if content == nil || strings.TrimSpace(*content) == "" {
			return &httperrors.ErrorDetail{Field: "content", Reason: "Text message requires content"}
		}
		if utf8.RuneCountInString(*content) > MaxMessageContentLen {
			return &httperrors.ErrorDetail{Field: "content", Reason: "Content must be at most 4096 characters"}
		}
		if stickerID != nil {
			return &httperrors.ErrorDetail{Field: "sticker_id", Reason: "Text message cannot contain a sticker"}
		}
	case "sticker":
		if stickerID == nil {
			return &httperrors.ErrorDetail{Field: "sticker_id", Reason: "Sticker message requires a sticker"}
		}
		if content != nil {
			return &httperrors.ErrorDetail{Field: "content", Reason: "Sticker message cannot contain text"}
		}
	default:
		return &httperrors.ErrorDetail{Field: "type", Reason: "Message type must be text or sticker"}
	}
	return nil
}
