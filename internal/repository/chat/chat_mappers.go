package chat

import "app/internal/models"

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
