package chat

import "github.com/go-park-mail-ru/2026_2_se7en/internal/models"

func MessageToResponse(message models.Message) MessageResponse {
	return MessageResponse{
		ID:               message.ID,
		UserID:           message.UserID,
		ChatID:           message.ChatID,
		ReplyToMessageID: message.ReplyToMessageID,
		StickerID:        message.StickerID,
		Content:          message.Content,
		IsEdited:         message.IsEdited,
		IsRead:           message.IsRead,
		Type:             message.Type,
		CreatedAt:        message.CreatedAt,
		UpdatedAt:        message.UpdatedAt,
	}
}

func MessagesToResponse(messages []models.Message) ListMessagesResponse {
	items := make([]MessageResponse, 0, len(messages))
	for _, message := range messages {
		items = append(items, MessageToResponse(message))
	}

	return ListMessagesResponse{Items: items}
}
