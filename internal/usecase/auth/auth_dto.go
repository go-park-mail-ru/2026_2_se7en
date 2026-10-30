package auth

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

type SessionDB struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
