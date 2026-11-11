package users

//go:generate mockgen -source=interfaces.go -destination=mock_usecase_test.go -package=users

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type UsersUsecase interface {
	GetUsersList(ctx context.Context, requesterID uuid.UUID, limit, offset int, nickname string) ([]models.UserSearchResult, error)
}
