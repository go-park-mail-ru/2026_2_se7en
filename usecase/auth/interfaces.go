package auth

import (
	"app/models"
	"context"

	"github.com/google/uuid"
)

type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (*models.UserWithSession, error)
	GetMe(ctx context.Context, userID uuid.UUID) (*models.UserWithSession, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
}
