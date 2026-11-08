package contact

//go:generate mockgen -source=interfaces.go -destination=mock_usecase_test.go -package=contact

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

type ContactUsecase interface {
	GetContactsList(ctx context.Context, userID uuid.UUID, limit, offset int, q string) ([]models.Contact, error)
	CreateContact(ctx context.Context, userID uuid.UUID, nickname string, pseudonym *string) error
	DeleteContact(ctx context.Context, userID, contactUserID uuid.UUID) error
}
