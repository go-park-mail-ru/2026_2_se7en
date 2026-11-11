package register

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
)

type RegisterUsecase interface {
	Register(ctx context.Context, user *models.User, plainPassword string) (*models.UserWithSession, error)
}
