package auth

import (
	"context"

	"app/models"

	"github.com/google/uuid"
)

type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (*LoginResult, error)
	GetMe(ctx context.Context, userID uuid.UUID) (*MeResult, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
}

type LoginResult struct {
	User      *models.User
	SessionID uuid.UUID
}

type MeResult struct {
	User *models.User
}
