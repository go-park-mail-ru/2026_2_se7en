package chat

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"

	"github.com/google/uuid"
)

type ChatUsecaseImpl struct {
	chatRepo ChatRepository
}

func NewChatUsecase(chatRepo ChatRepository) *ChatUsecaseImpl {
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
