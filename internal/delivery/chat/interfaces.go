package chat

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type ChatUsecase interface {
	IsMember(ctx context.Context, userID, chatID uuid.UUID) (bool, error)
	GetListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error)
	GetChatDetail(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error)
	DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error

	CreateDialog(ctx context.Context, userID, peerID uuid.UUID) (*models.ChatDetail, models.DialogStatus, error)
	HideDialog(ctx context.Context, userID, chatID uuid.UUID) error

	CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error)
	UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error)
	GetListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error)
	AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) (*models.ChatMember, error)
	RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error

	GetListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error)
	CreateMessage(ctx context.Context, userID, chatID uuid.UUID, replyToMessageID, stickerID *uuid.UUID, content *string, messageType string) (*models.Message, error)
	MarkMessageRead(ctx context.Context, userID, chatID, messageID uuid.UUID) (bool, error)
}
