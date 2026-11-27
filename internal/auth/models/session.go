package models

import (
	"time"

	"github.com/google/uuid"
)

const SessionTTL = 30 * 24 * time.Hour

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SessionUser struct {
	UserID uuid.UUID
}
