package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PhoneNumber  *string
	PasswordHash string
	Profile      *Profile
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type UserWithSession struct {
	User      *User
	SessionID uuid.UUID
}
