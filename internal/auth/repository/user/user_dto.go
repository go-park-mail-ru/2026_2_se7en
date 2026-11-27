package user

import (
	"time"

	"github.com/google/uuid"
)

type UserDB struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	PhoneNumber  *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
