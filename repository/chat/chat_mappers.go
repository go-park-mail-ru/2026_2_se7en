package chat

import "app/models"

func ChatListItemDBToModel(db *ChatListItemDB) *models.ChatListItem {
	if db == nil {
		return nil
	}
	item := &models.ChatListItem{
		Chat: &models.Chat{
			ID:           db.Chat.ID,
			IconID:       db.Chat.IconID,
			Name:         db.Chat.Name,
			Description:  db.Chat.Description,
			Type:         db.Chat.Type,
			MembersCount: db.Chat.MembersCount,
			CreatedAt:    db.Chat.CreatedAt,
			UpdatedAt:    db.Chat.UpdatedAt,
			DeletedAt:    db.Chat.DeletedAt,
		},
	}
	if db.LastMessage != nil {
		item.LastMessage = &models.MessagePreview{
			ID:        db.LastMessage.ID,
			Content:   db.LastMessage.Content,
			Type:      db.LastMessage.Type,
			CreatedAt: db.LastMessage.CreatedAt,
		}
	}
	return item
}

func ChatDetailDBToModel(db *ChatDetailDB) *models.ChatDetail {
	if db == nil {
		return nil
	}
	return &models.ChatDetail{
		Chat: &models.Chat{
			ID:           db.Chat.ID,
			IconID:       db.Chat.IconID,
			Name:         db.Chat.Name,
			Description:  db.Chat.Description,
			Type:         db.Chat.Type,
			MembersCount: db.Chat.MembersCount,
			CreatedAt:    db.Chat.CreatedAt,
			UpdatedAt:    db.Chat.UpdatedAt,
			DeletedAt:    db.Chat.DeletedAt,
		},
		Role: db.Role,
	}
}

func ChatMemberDBToModel(db *ChatMemberDB) models.ChatMember {
	return models.ChatMember{
		UserID: db.UserID,
		Role:   db.Role,
		Profile: &models.Profile{
			ID:        db.Profile.ID,
			IconID:    db.Profile.IconID,
			UserID:    db.Profile.UserID,
			Nickname:  db.Profile.Nickname,
			FirstName: db.Profile.FirstName,
			LastName:  db.Profile.LastName,
			Bio:       db.Profile.Bio,
			CreatedAt: db.Profile.CreatedAt,
			UpdatedAt: db.Profile.UpdatedAt,
		},
	}
}

func MessageDBToModel(db *MessageDB) models.Message {
	return models.Message{
		ID:               db.ID,
		UserID:           db.UserID,
		ChatID:           db.ChatID,
		ReplyToMessageID: db.ReplyToMessageID,
		StickerID:        db.StickerID,
		Content:          db.Content,
		IsEdited:         db.IsEdited,
		Type:             db.Type,
		CreatedAt:        db.CreatedAt,
		UpdatedAt:        db.UpdatedAt,
	}
}
