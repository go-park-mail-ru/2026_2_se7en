package contact

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"

	"github.com/google/uuid"
)

type ContactDB struct {
	UserID    uuid.UUID
	Pseudonym *string
	Profile   profile.ProfileDB
}
