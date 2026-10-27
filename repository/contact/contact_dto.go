package contact

import (
	"app/repository/profile"

	"github.com/google/uuid"
)

type ContactDB struct {
	UserID    uuid.UUID
	Pseudonym *string
	Profile   profile.ProfileDB
}
