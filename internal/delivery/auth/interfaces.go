package auth

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (*models.UserWithSession, error)
	GetMe(ctx context.Context, userID uuid.UUID) (*models.UserWithSession, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
}
