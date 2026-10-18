package register

import (
	"context"

	"app/models"

	"github.com/google/uuid"
)

type RegisterUsecase interface {
	Register(ctx context.Context, user *models.User, plainPassword string) (*RegisterResult, error)
}

type RegisterResult struct {
	User      *models.User
	SessionID uuid.UUID
}
