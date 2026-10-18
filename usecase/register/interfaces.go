package register

import (
	"context"

	"app/models"
)

type RegisterUsecase interface {
	Register(ctx context.Context, user *models.User, plainPassword string) (*models.UserWithSession, error)
}
