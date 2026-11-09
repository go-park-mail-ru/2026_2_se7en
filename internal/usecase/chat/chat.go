package chat

import (
	"context"
	"time"

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

func (uc *ChatUsecaseImpl) GetActiveMemberIDs(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error) {
	return uc.repo.GetActiveMemberIDs(ctx, chatID)
}

func (uc *ChatUsecaseImpl) GetListUserChats(ctx context.Context, userID uuid.UUID, limit int, chatType string, cursorTime *time.Time, cursorID uuid.UUID) ([]*models.ChatListItem, error) {
	items, err := uc.repo.GetListUserChats(ctx, userID, limit, chatType, cursorTime, cursorID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []*models.ChatListItem{}, nil
	}
	return items, nil
}

func (uc *ChatUsecaseImpl) GetChatDetail(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error) {
	return uc.repo.GetChat(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error {
	return uc.repo.DeleteChat(ctx, userID, chatID)
}
