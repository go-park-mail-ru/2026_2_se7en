package chat

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error)
}
