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
			Name:         item.Chat.Name,
			Description:  item.Chat.Description,
			Type:         item.Chat.Type,
			MembersCount: item.Chat.MembersCount,
			CreatedAt:    item.Chat.CreatedAt,
			UpdatedAt:    item.Chat.UpdatedAt,
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
	return &ChatDetailResponse{Chat: ChatToResponse(detail.Chat), Role: detail.Role}
}

func ChatMemberToResponse(member models.ChatMember) ChatMemberResponse {
	response := ChatMemberResponse{UserID: member.UserID, Role: member.Role}
	if member.Profile != nil {
		p := member.Profile
		response.Profile = &MemberProfileResponse{
			ID:        p.ID,
			IconID:    p.IconID,
			UserID:    p.UserID,
			Nickname:  p.Nickname,
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Bio:       p.Bio,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}
	}
	return response
}

func ChatMembersToResponse(members []models.ChatMember) ListMembersResponse {
	items := make([]ChatMemberResponse, 0, len(members))
	for _, member := range members {
		items = append(items, ChatMemberToResponse(member))
	}
	return ListMembersResponse{Items: items}
}

func MessageToResponse(message models.Message) MessageResponse {
	return MessageResponse{
		ID:               message.ID,
		UserID:           message.UserID,
		ChatID:           message.ChatID,
		ReplyToMessageID: message.ReplyToMessageID,
		StickerID:        message.StickerID,
		Content:          message.Content,
		IsEdited:         message.IsEdited,
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
