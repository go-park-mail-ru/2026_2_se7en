package contact

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
)

type ContactUsecaseImpl struct {
	log *slog.Logger

	contactRepo ContactRepository
	userRepo    UserRepository
}

func NewContactUsecase(contactRepo ContactRepository, userRepo UserRepository, baseLog *slog.Logger) *ContactUsecaseImpl {
	return &ContactUsecaseImpl{
		log: logger.Scope(baseLog, logger.LayerUsecase, logger.ModuleContact),

		contactRepo: contactRepo,
		userRepo:    userRepo,
	}
}

func (uc *ContactUsecaseImpl) CreateContact(ctx context.Context, userID uuid.UUID, nickname string, pseudonym *string) error {

	user, err := uc.userRepo.FindByNickname(ctx, nickname)
	if err != nil {
		uc.log.ErrorContext(ctx, "CreateContact failed", logger.Error(err))

		if errors.Is(err, apperrors.ErrUserNotFound) {
			return apperrors.ErrUserNotFound
		}
		return err
	}

	if user.ID == userID {
		return apperrors.ErrCannotAddSelfAsContact
	}

	return uc.contactRepo.CreateContact(ctx, userID, user.ID, pseudonym)
}

func (uc *ContactUsecaseImpl) GetContactsList(ctx context.Context, userID uuid.UUID, limit, offset int, q string) ([]models.Contact, error) {
	return uc.contactRepo.GetContactsList(ctx, userID, limit, offset, q)
}

func (uc *ContactUsecaseImpl) DeleteContact(ctx context.Context, userID, contactUserID uuid.UUID) error {

	err := uc.contactRepo.DeleteContact(ctx, userID, contactUserID)
	if err != nil {
		uc.log.ErrorContext(ctx, "DeleteContact failed", logger.Error(err))

		if errors.Is(err, apperrors.ErrContactNotFound) {
			return apperrors.ErrContactNotFound
		}
		return err
	}
	return nil
}
