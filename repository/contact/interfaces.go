package contact

import (
	"context"

	"app/models"

	"github.com/google/uuid"
)

type Repository interface {
	ListContacts(ctx context.Context, userID uuid.UUID, limit, offset int, query string) ([]models.Contact, error)
	AddContact(ctx context.Context, userID, contactID uuid.UUID) error
	DeleteContact(ctx context.Context, userID, contactID uuid.UUID) (bool, error)
}
