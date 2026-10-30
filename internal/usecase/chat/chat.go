package chat

import (
	"context"

	"app/internal/models"
	"app/internal/repository/chat"

	"github.com/google/uuid"
)

type ChatUsecaseImpl struct {
	chatRepo chat.Repository
}

func NewChatUsecase(chatRepo chat.Repository) *ChatUsecaseImpl {
	return &ChatUsecaseImpl{chatRepo: chatRepo}
}

func (uc *ChatUsecaseImpl) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
	items, err := uc.chatRepo.ListUserChats(ctx, userID, limit, offset, chatType)
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []*models.ChatListItem{}
	}

	return items, nil
}
