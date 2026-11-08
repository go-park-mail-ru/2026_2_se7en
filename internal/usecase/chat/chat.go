package chat

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type ChatUsecaseImpl struct {
	repo ChatRepository
}

func NewChatUsecase(repo ChatRepository) *ChatUsecaseImpl {
	return &ChatUsecaseImpl{repo: repo}
}

func (uc *ChatUsecaseImpl) IsMember(ctx context.Context, userID, chatID uuid.UUID) (bool, error) {
	return uc.repo.IsMember(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) GetListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
	items, err := uc.repo.GetListUserChats(ctx, userID, limit, offset, chatType)
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []*models.ChatListItem{}
	}

	return items, nil
}

func (uc *ChatUsecaseImpl) GetChatDetail(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error) {
	return uc.repo.GetChat(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error {
	return uc.repo.DeleteChat(ctx, userID, chatID)
}
