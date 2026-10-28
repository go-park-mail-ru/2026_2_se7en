package contact

import (
	"app/models"
	"app/repository/profile"
)

func ContactDBToModel(db *ContactDB) models.Contact {
	return models.Contact{
		UserID:    db.UserID,
		Pseudonym: db.Pseudonym,
		Profile:   profile.ProfileDBToModel(&db.Profile),
	}
}
