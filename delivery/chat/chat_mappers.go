package chat

import "app/models"

func ChatItemsToResponse(items []*models.ChatListItem) ListChatsResponse {
	if items == nil {
		return ListChatsResponse{Items: []ChatItemDTO{}}
	}

	dtoItems := make([]ChatItemDTO, 0, len(items))
	for _, item := range items {
		if item == nil || item.Chat == nil {
			continue
		}

		dto := ChatItemDTO{
			ID:           item.Chat.ID,
			IconID:       item.Chat.IconID,
			Name:         item.Chat.Name,
			Description:  item.Chat.Description,
			Type:         item.Chat.Type,
			MembersCount: item.Chat.MembersCount,
			CreatedAt:    item.Chat.CreatedAt,
			UpdatedAt:    item.Chat.UpdatedAt,
		}

		if item.LastMessage != nil {
			dto.LastMessage = &MessagePreviewDTO{
				ID:        item.LastMessage.ID,
				Content:   item.LastMessage.Content,
				Type:      item.LastMessage.Type,
				CreatedAt: item.LastMessage.CreatedAt,
			}
		}

		dtoItems = append(dtoItems, dto)
	}

	return ListChatsResponse{Items: dtoItems}
}
