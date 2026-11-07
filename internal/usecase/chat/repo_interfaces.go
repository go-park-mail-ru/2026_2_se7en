package chat

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"

	"github.com/google/uuid"
)

type ChatRepository interface {
	IsMember(ctx context.Context, userID, chatID uuid.UUID) (bool, error)
	ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error)
	GetChatDetail(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error)
	GetOrCreateDialog(ctx context.Context, userID, peerID uuid.UUID) (*models.ChatDetail, models.DialogStatus, error)
	HideDialog(ctx context.Context, userID, chatID uuid.UUID) error
	CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error)
	UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error)
	DeleteGroup(ctx context.Context, userID, chatID uuid.UUID) error
	ListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error)
	AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) (*models.ChatMember, error)
	RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error
	ListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error)
	CreateMessage(ctx context.Context, userID, chatID uuid.UUID, replyToMessageID, stickerID *uuid.UUID, content *string, messageType string) (*models.Message, error)
	MarkMessageRead(ctx context.Context, userID, chatID, messageID uuid.UUID) (bool, error)
}
