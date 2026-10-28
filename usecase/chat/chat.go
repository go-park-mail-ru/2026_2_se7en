package chat

import (
	"context"

	"app/apperrors"
	"app/models"
	"app/repository/chat"

	"github.com/google/uuid"
)

type ChatUsecaseImpl struct {
	chatRepo chat.Repository
}

func NewChatUsecase(chatRepo chat.Repository) *ChatUsecaseImpl {
	return &ChatUsecaseImpl{chatRepo: chatRepo}
}

func (uc *ChatUsecaseImpl) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
	if limit < 1 || limit > 50 {
		return nil, apperrors.ErrInvalidChatLimit
	}
	if offset < 0 {
		return nil, apperrors.ErrInvalidChatOffset
	}
	if chatType != "" && chatType != "dialog" && chatType != "group" && chatType != "channel" {
		return nil, apperrors.ErrInvalidChatType
	}
	items, err := uc.chatRepo.ListUserChats(ctx, userID, limit, offset, chatType)
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []*models.ChatListItem{}
	}

	return items, nil
}
