package contact

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
)

func ContactDBToModel(db *ContactDB) models.Contact {
	return models.Contact{
		UserID:    db.UserID,
		Pseudonym: db.Pseudonym,
		Profile:   profile.ProfileDBToModel(&db.Profile),
	}
}
