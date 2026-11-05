package chat

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"

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

func (uc *ChatUsecaseImpl) FindForUser(ctx context.Context, userID, chatID uuid.UUID) (*models.ChatDetail, error) {
	return uc.chatRepo.FindForUser(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) GetOrCreateDialog(ctx context.Context, userID, peerID uuid.UUID) (*models.ChatDetail, error) {
	return uc.chatRepo.GetOrCreateDialog(ctx, userID, peerID)
}

func (uc *ChatUsecaseImpl) HideDialog(ctx context.Context, userID, chatID uuid.UUID) error {
	return uc.chatRepo.HideDialog(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error) {
	return uc.chatRepo.CreateGroup(ctx, userID, name, description, members)
}

func (uc *ChatUsecaseImpl) UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error) {
	return uc.chatRepo.UpdateGroup(ctx, userID, chatID, name, description, setDescription)
}

func (uc *ChatUsecaseImpl) DeleteGroup(ctx context.Context, userID, chatID uuid.UUID) error {
	return uc.chatRepo.DeleteGroup(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) ListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error) {
	return uc.chatRepo.ListMembers(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) error {
	return uc.chatRepo.AddMember(ctx, userID, chatID, newUserID)
}

func (uc *ChatUsecaseImpl) RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error {
	return uc.chatRepo.RemoveMember(ctx, userID, chatID, targetID)
}

func (uc *ChatUsecaseImpl) ListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error) {
	return uc.chatRepo.ListMessages(ctx, userID, chatID, limit, before)
}
