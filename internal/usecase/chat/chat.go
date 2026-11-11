package chat

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
)

type ChatUsecaseImpl struct {
	log *slog.Logger

	repo ChatRepository
}

func NewChatUsecase(repo ChatRepository, baseLog *slog.Logger) *ChatUsecaseImpl {
	return &ChatUsecaseImpl{
		log:  logger.Scope(baseLog, logger.LayerUsecase, logger.ModuleChat),
		repo: repo}
}

func (uc *ChatUsecaseImpl) IsMember(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatMembership, error) {
	return uc.repo.IsMember(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) GetActiveMemberIDs(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error) {
	return uc.repo.GetActiveMemberIDs(ctx, chatID)
}

func (uc *ChatUsecaseImpl) GetListUserChats(ctx context.Context, userID uuid.UUID, limit int, chatType string, cursorTime *time.Time, cursorID uuid.UUID) ([]*models.ChatListItem, error) {
	items, err := uc.repo.GetListUserChats(ctx, userID, limit, chatType, cursorTime, cursorID)
	if err != nil {
		uc.log.ErrorContext(ctx, "GetListUserChats failed", logger.Error(err))

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
