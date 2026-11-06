package models

import (
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	UserID        uuid.UUID
	ContactUserID uuid.UUID
	Pseudonym     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Profile       *Profile
	IsOnline      bool
}
