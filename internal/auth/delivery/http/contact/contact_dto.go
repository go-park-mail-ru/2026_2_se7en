package contact

import (
	"time"

	"github.com/google/uuid"
)

type ContactResponse struct {
	UserID        uuid.UUID  `json:"user_id" format:"uuid" binding:"required"`
	ContactUserID uuid.UUID  `json:"contact_user_id" format:"uuid" binding:"required"`
	Pseudonym     *string    `json:"pseudonym,omitempty"`
	Nickname      string     `json:"nickname" binding:"required"`
	FirstName     string     `json:"first_name" binding:"required"`
	LastName      *string    `json:"last_name,omitempty"`
	IconID        *uuid.UUID `json:"icon_id,omitempty" format:"uuid"`
	IconURL       *string    `json:"icon_url,omitempty"`
	IsOnline      bool       `json:"is_online" binding:"required"`
	CreatedAt     time.Time  `json:"created_at" format:"date-time" binding:"required"`
	UpdatedAt     time.Time  `json:"updated_at" format:"date-time" binding:"required"`
}

type CreateContactRequest struct {
	Nickname  string  `json:"nickname" binding:"required"`
	Pseudonym *string `json:"pseudonym,omitempty"`
}

type DeleteContactRequest struct {
	ContactUserID uuid.UUID `json:"contact_user_id" format:"uuid" binding:"required"`
}
