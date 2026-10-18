package chat

import (
	"context"

	"app/models"

	"github.com/google/uuid"
)

type ChatRepository interface {
	ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error)
}
