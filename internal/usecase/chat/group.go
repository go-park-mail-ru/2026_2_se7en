package chat

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
	"github.com/google/uuid"
)

func (uc *ChatUsecaseImpl) UploadGroupAvatar(ctx context.Context, userID, chatID uuid.UUID, upload *models.FileUpload) (*models.ChatDetail, error) {
	url, filename, err := tools.SaveAvatar(upload)
	if err != nil {
		return nil, err
	}
	detail, err := uc.repo.UpdateGroupAvatar(ctx, userID, chatID, url, filename, upload.ContentType)
	if err != nil {
		_ = tools.RemoveAvatar(filename)
		return nil, err
	}
	return detail, nil
}

func (uc *ChatUsecaseImpl) CreateGroup(ctx context.Context, userID uuid.UUID, name string, description *string, members []uuid.UUID) (*models.ChatDetail, error) {
	return uc.repo.CreateGroup(ctx, userID, name, description, members)
}

func (uc *ChatUsecaseImpl) UpdateGroup(ctx context.Context, userID, chatID uuid.UUID, name, description *string, setDescription bool) (*models.ChatDetail, error) {
	return uc.repo.UpdateGroup(ctx, userID, chatID, name, description, setDescription)
}

func (uc *ChatUsecaseImpl) GetListMembers(ctx context.Context, userID, chatID uuid.UUID) ([]models.ChatMember, error) {
	if _, err := uc.repo.GetChat(ctx, userID, chatID); err != nil {
		return nil, err
	}

	return uc.repo.GetListMembers(ctx, userID, chatID)
}

func (uc *ChatUsecaseImpl) AddMember(ctx context.Context, userID, chatID, newUserID uuid.UUID) (*models.ChatMember, error) {
	return uc.repo.AddMember(ctx, userID, chatID, newUserID)
}

func (uc *ChatUsecaseImpl) RemoveMember(ctx context.Context, userID, chatID, targetID uuid.UUID) error {
	return uc.repo.RemoveMember(ctx, userID, chatID, targetID)
}
