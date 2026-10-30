package chat

import (
	"app/internal/models"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error)
}
