package register

import (
	"context"

	"app/internal/models"
)

type RegisterUsecase interface {
	Register(ctx context.Context, user *models.User, plainPassword string) (*models.UserWithSession, error)
}
