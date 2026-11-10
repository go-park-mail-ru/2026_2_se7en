package chat

import (
	"strings"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
)

func (db *ChatDB) applyPresentation() {
	if db.Type != "dialog" {
		db.IconURL = db.ChatIconURL
		return
	}

	db.IconURL = db.PeerIconURL
	nameParts := make([]string, 0, 2)
	if db.PeerFirstName != nil {
		nameParts = append(nameParts, *db.PeerFirstName)
	}
	if db.PeerLastName != nil {
		nameParts = append(nameParts, *db.PeerLastName)
	}
	if fullName := strings.Trim(strings.Join(nameParts, " "), " "); fullName != "" {
		db.Name = fullName
	} else if db.PeerNickname != nil {
		db.Name = *db.PeerNickname
	}
}

func ChatListItemDBToModel(db *ChatListItemDB) *models.ChatListItem {
	if db == nil {
		return nil
	}
	db.Chat.applyPresentation()
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
		PeerID:      db.Chat.PeerID,
		IconURL:     db.Chat.IconURL,
		UnreadCount: db.UnreadCount,
		ActivityAt:  db.ActivityAt,
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
	db.Chat.applyPresentation()
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
		PeerID:  db.Chat.PeerID,
		IconURL: db.Chat.IconURL,
		Role:    db.Role,
	}
}

func ChatMemberDBToModel(db *ChatMemberDB) models.ChatMember {
	return models.ChatMember{
		UserID: db.UserID,
		Role:   db.Role,
		Profile: &models.Profile{
			ID:        db.Profile.ID,
			IconID:    db.Profile.IconID,
			AvatarURL: db.Profile.IconURL,
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
		IsRead:           db.IsRead,
		Type:             db.Type,
		CreatedAt:        db.CreatedAt,
		UpdatedAt:        db.UpdatedAt,
	}
}
