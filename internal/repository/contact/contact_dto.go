package contact

import (
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
	"github.com/google/uuid"
)

type ContactDB struct {
	UserID        uuid.UUID
	ContactUserID uuid.UUID
	Pseudonym     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Profile       profile.ProfileDB
	IconURL       *string
}
