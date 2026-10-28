package chat

import (
	"context"

	"app/models"

	"github.com/google/uuid"
)

type Repository interface {
	ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error)
	FindForUser(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error)
	GetOrCreateDialog(ctx context.Context, userID, peerID uuid.UUID) (*models.ChatDetail, error)
	HideDialog(ctx context.Context, userID, chatID uuid.UUID) error
	CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error)
	UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error)
	DeleteGroup(ctx context.Context, userID, chatID uuid.UUID) error
	ListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error)
	AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) error
	RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error
	ListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error)
}
