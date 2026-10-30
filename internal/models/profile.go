package models

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	IconID    *uuid.UUID
	Nickname  string
	FirstName string
	LastName  *string
	Bio       *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
