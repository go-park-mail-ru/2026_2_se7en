package chat

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"

	"github.com/google/uuid"
)

type ChatRepository interface {
	IsMember(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatMembership, error)
	GetActiveMemberIDs(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error)
	GetListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string, cursorTime *time.Time, cursorID uuid.UUID) ([]*models.ChatListItem, error)
	GetChat(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error)
	DeleteChat(ctx context.Context, userID, chatID uuid.UUID) error

	CreateChat(ctx context.Context, userID, peerID uuid.UUID) (uuid.UUID, models.DialogStatus, error)
	HideDialog(ctx context.Context, userID, chatID uuid.UUID) error

	CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error)
	UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error)
	UpdateGroupAvatar(ctx context.Context, userID, chatID uuid.UUID, url, filename, contentType string) (*models.ChatDetail, error)
	GetListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error)
	AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) (*models.ChatMember, error)
	UpdateMemberRole(ctx context.Context, userID, chatID, targetID uuid.UUID, role string) (*models.ChatMember, error)
	RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error

	MessageExists(ctx context.Context, chatID, messageID uuid.UUID) (bool, error)
	GetListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error)
	CreateMessage(ctx context.Context, userID, chatID uuid.UUID, replyToMessageID, stickerID *uuid.UUID, content *string, messageType string) (*models.Message, error)
	MarkMessageRead(ctx context.Context, userID, chatID, messageID uuid.UUID) (bool, error)
	GetMessageAuthor(ctx context.Context, chatID, messageID uuid.UUID) (uuid.UUID, error)
}
