package chat

import (
	"context"

	"app/internal/models"

	"github.com/google/uuid"
)

type ChatUsecase interface {
	ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error)
}
