package user

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	CheckEmailExist(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, email, passwordHash string, phone *string) (*models.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}
