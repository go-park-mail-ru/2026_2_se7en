package session

import (
	"time"

	"github.com/google/uuid"
)

type SessionDB struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
