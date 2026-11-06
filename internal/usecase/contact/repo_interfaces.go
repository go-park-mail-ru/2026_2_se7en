package contact

//go:generate mockgen -source=repo_interfaces.go -destination=mock_repo_test.go -package=contact

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type ContactRepository interface {
	GetContactsList(ctx context.Context, userID uuid.UUID) ([]models.Contact, error)
	CreateContact(ctx context.Context, userID, contactUserID uuid.UUID, pseudonym *string) error
	DeleteContact(ctx context.Context, userID, contactUserID uuid.UUID) error
	CheckContactExists(ctx context.Context, userID, contactUserID uuid.UUID) (bool, error)
}

type UserRepository interface {
	FindByNickname(ctx context.Context, nickname string) (*models.User, error)
}
