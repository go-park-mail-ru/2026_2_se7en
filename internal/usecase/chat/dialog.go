package chat

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

func (uc *ChatUsecaseImpl) CreateDialog(ctx context.Context, userID, peerID uuid.UUID) (*models.ChatDetail, models.DialogStatus, error) {
	chatID, status, err := uc.repo.CreateChat(ctx, userID, peerID)
	if err != nil {
		return nil, models.DialogExisting, err
	}

	detail, err := uc.repo.GetChat(ctx, userID, chatID)
	if err != nil {
		return nil, models.DialogExisting, err
	}

	return detail, status, nil
}

func (uc *ChatUsecaseImpl) HideDialog(ctx context.Context, userID, chatID uuid.UUID) error {
	return uc.repo.HideDialog(ctx, userID, chatID)
}
