package profile

import (
	"time"

	"github.com/google/uuid"
)

type ProfileDB struct {
	ID        uuid.UUID
	IconID    *uuid.UUID
	UserID    uuid.UUID
	Nickname  string
	FirstName string
	LastName  *string
	Bio       *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
