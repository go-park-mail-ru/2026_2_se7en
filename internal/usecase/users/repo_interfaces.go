package users

//go:generate mockgen -source=repo_interfaces.go -destination=mock_repo_test.go -package=users

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type UsersRepository interface {
	GetUsersList(ctx context.Context, requesterID uuid.UUID, limit, offset int, nickname string) ([]models.UserSearchResult, error)
}
