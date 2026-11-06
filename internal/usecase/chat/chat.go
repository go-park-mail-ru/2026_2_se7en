package chat

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"

	"github.com/google/uuid"
)

type ChatUsecaseImpl struct {
	chatRepo ChatRepository
}

func NewChatUsecase(chatRepo ChatRepository) *ChatUsecaseImpl {
	return &ChatUsecaseImpl{chatRepo: chatRepo}
}

func (uc *ChatUsecaseImpl) ListUserChats(ctx context.Context, userID uuid.UUID, limit, offset int, chatType string) ([]*models.ChatListItem, error) {
	if limit < 1 || limit > 50 {
		return nil, apperrors.ErrInvalidChatLimit
	}
	if offset < 0 {
		return nil, apperrors.ErrInvalidChatOffset
	}
	if chatType != "" && chatType != "dialog" && chatType != "group" && chatType != "channel" {
		return nil, apperrors.ErrInvalidChatType
	}
	items, err := uc.chatRepo.ListUserChats(ctx, userID, limit, offset, chatType)
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []*models.ChatListItem{}
	}

	return items, nil
}

func (uc *ChatUsecaseImpl) GetChatDetail(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error) {
	return uc.chatRepo.GetChatDetail(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) GetOrCreateDialog(ctx context.Context, userID, peerID uuid.UUID) (*models.ChatDetail, error) {
	if userID == peerID {
		return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "peer_id", Reason: "Cannot create a dialog with yourself"}})
	}
	return uc.chatRepo.GetOrCreateDialog(ctx, userID, peerID)
}

func (uc *ChatUsecaseImpl) HideDialog(ctx context.Context, userID, chatID uuid.UUID) error {
	return uc.chatRepo.HideDialog(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error) {
	if detail := tools.ValidateChatName(name); detail != nil {
		return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{*detail})
	}
	if detail := tools.ValidateChatDescription(description); detail != nil {
		return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{*detail})
	}
	seen := make(map[uuid.UUID]struct{}, len(members))
	for _, memberID := range members {
		if memberID == userID {
			return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "members", Reason: "Creator must not be included in members"}})
		}
		if _, exists := seen[memberID]; exists {
			return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "members", Reason: "Members must be unique"}})
		}
		seen[memberID] = struct{}{}
	}
	return uc.chatRepo.CreateGroup(ctx, userID, name, description, members)
}

func (uc *ChatUsecaseImpl) UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error) {
	if name == nil && !setDescription {
		return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "name", Reason: "At least one field must be updated"}})
	}
	if description != nil && !setDescription {
		return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "description", Reason: "Description update flag is required"}})
	}
	if name != nil {
		if detail := tools.ValidateChatName(*name); detail != nil {
			return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{*detail})
		}
	}
	if setDescription {
		if detail := tools.ValidateChatDescription(description); detail != nil {
			return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{*detail})
		}
	}
	return uc.chatRepo.UpdateGroup(ctx, userID, chatID, name, description, setDescription)
}

func (uc *ChatUsecaseImpl) DeleteGroup(ctx context.Context, userID, chatID uuid.UUID) error {
	return uc.chatRepo.DeleteGroup(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) ListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error) {
	return uc.chatRepo.ListMembers(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) (*models.ChatMember, error) {
	if newUserID == userID {
		return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "new_user_id", Reason: "Cannot add yourself"}})
	}
	return uc.chatRepo.AddMember(ctx, userID, chatID, newUserID)
}

func (uc *ChatUsecaseImpl) RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error {
	return uc.chatRepo.RemoveMember(ctx, userID, chatID, targetID)
}

func (uc *ChatUsecaseImpl) ListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error) {
	if limit < 1 || limit > 50 {
		return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{{Field: "limit", Reason: "Limit must be between 1 and 50"}})
	}
	return uc.chatRepo.ListMessages(ctx, userID, chatID, limit, before)
}

func (uc *ChatUsecaseImpl) CreateMessage(ctx context.Context, userID, chatID uuid.UUID, replyToMessageID, stickerID *uuid.UUID, content *string, messageType string) (*models.Message, error) {
	if detail := tools.ValidateMessageFields(messageType, content, stickerID); detail != nil {
		return nil, apperrors.NewBadRequest("Validation failed", []apperrors.ErrorDetail{*detail})
	}
	return uc.chatRepo.CreateMessage(ctx, userID, chatID, replyToMessageID, stickerID, content, messageType)
}
