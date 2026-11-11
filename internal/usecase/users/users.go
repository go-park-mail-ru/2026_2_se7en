package users

import (
	"context"
	"log/slog"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
)

type UsersUsecaseImpl struct {
	log       *slog.Logger
	usersRepo UsersRepository
}

func NewUsersUsecase(usersRepo UsersRepository, baseLog *slog.Logger) *UsersUsecaseImpl {
	return &UsersUsecaseImpl{
		log:       logger.Scope(baseLog, logger.LayerUsecase, logger.ModuleUser),
		usersRepo: usersRepo,
	}
}

func (uc *UsersUsecaseImpl) GetUsersList(ctx context.Context, requesterID uuid.UUID, limit, offset int, nickname string) ([]models.UserSearchResult, error) {
	return uc.usersRepo.GetUsersList(ctx, requesterID, limit, offset, nickname)
}
