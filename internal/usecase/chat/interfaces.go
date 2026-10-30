package chat

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"

	"github.com/google/uuid"
)

type ChatUsecase interface {
	ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error)
}
