package chat

import "github.com/go-park-mail-ru/2026_2_se7en/internal/models"

func ChatItemsToResponse(items []*models.ChatListItem) ListChatsResponse {
	if items == nil {
		return ListChatsResponse{Items: []ChatItemResponse{}}
	}

	ResponseItems := make([]ChatItemResponse, 0, len(items))
	for _, item := range items {
		if item == nil || item.Chat == nil {
			continue
		}

		Response := ChatItemResponse{
			ID:           item.Chat.ID,
			IconID:       item.Chat.IconID,
			IconURL:      item.IconURL,
			IsOnline:     item.IsOnline,
			Name:         item.Chat.Name,
			Description:  item.Chat.Description,
			Type:         item.Chat.Type,
			MembersCount: item.Chat.MembersCount,
			CreatedAt:    item.Chat.CreatedAt,
			UpdatedAt:    item.Chat.UpdatedAt,
			UnreadCount:  item.UnreadCount,
		}

		if item.LastMessage != nil {
			Response.LastMessage = &MessagePreviewResponse{
				ID:        item.LastMessage.ID,
				Content:   item.LastMessage.Content,
				Type:      item.LastMessage.Type,
				CreatedAt: item.LastMessage.CreatedAt,
			}
		}

		ResponseItems = append(ResponseItems, Response)
	}

	return ListChatsResponse{Items: ResponseItems}
}

func ChatToResponse(chat *models.Chat) *ChatResponse {
	if chat == nil {
		return nil
	}

	return &ChatResponse{
		ID:           chat.ID,
		IconID:       chat.IconID,
		Name:         chat.Name,
		Description:  chat.Description,
		Type:         chat.Type,
		MembersCount: chat.MembersCount,
		CreatedAt:    chat.CreatedAt,
		UpdatedAt:    chat.UpdatedAt,
		DeletedAt:    chat.DeletedAt,
	}
}

func ChatDetailToResponse(detail *models.ChatDetail) *ChatDetailResponse {
	if detail == nil {
		return nil
	}

	chat := ChatToResponse(detail.Chat)
	if chat != nil {
		chat.IconURL = detail.IconURL
		chat.IsOnline = detail.IsOnline
	}
	return &ChatDetailResponse{Chat: chat, Role: detail.Role}
}
