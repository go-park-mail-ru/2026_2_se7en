package chat

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
)

func (uc *ChatUsecaseImpl) GetListMessages(ctx context.Context, userID, chatID uuid.UUID, limit int, before *uuid.UUID) ([]models.Message, error) {
	if _, err := uc.repo.GetChat(ctx, userID, chatID); err != nil {
		uc.log.ErrorContext(ctx, "GetListMessages failed", logger.Error(err))

		return nil, err
	}

	if before != nil {
		exists, err := uc.repo.MessageExists(ctx, chatID, *before)
		if err != nil {
			uc.log.ErrorContext(ctx, "GetListMessages failed", logger.Error(err))

			return nil, err
		}

		if !exists {
			return nil, apperrors.ErrMessageNotFound
		}
	}

	return uc.repo.GetListMessages(ctx, userID, chatID, limit, before)
}

func (uc *ChatUsecaseImpl) CreateMessage(ctx context.Context, userID, chatID uuid.UUID, replyToMessageID, stickerID *uuid.UUID, content *string, messageType string) (*models.Message, error) {
	return uc.repo.CreateMessage(ctx, userID, chatID, replyToMessageID, stickerID, content, messageType)
}

func (uc *ChatUsecaseImpl) MarkMessageRead(ctx context.Context, userID, chatID, messageID uuid.UUID) (bool, error) {
	changed, err := uc.repo.MarkMessageRead(ctx, userID, chatID, messageID)
	if err != nil || changed {
		if err != nil {
			uc.log.ErrorContext(ctx, "MarkMessageRead failed", logger.Error(err))
		}

		return changed, err
	}

	if _, err := uc.repo.GetChat(ctx, userID, chatID); err != nil {
		uc.log.ErrorContext(ctx, "MarkMessageRead failed", logger.Error(err))

		return false, err
	}

	authorID, err := uc.repo.GetMessageAuthor(ctx, chatID, messageID)
	if err != nil {
		uc.log.ErrorContext(ctx, "MarkMessageRead failed", logger.Error(err))

		return false, err
	}

	if authorID == userID {
		return false, apperrors.ErrInvalidChatOperation
	}

	return false, nil
}
