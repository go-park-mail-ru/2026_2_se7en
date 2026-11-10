package contact

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
)

func ContactDBToModel(db *ContactDB) models.Contact {
	prof := profile.ProfileDBToModel(&db.Profile)
	prof.AvatarURL = db.IconURL
	return models.Contact{
		UserID:        db.UserID,
		ContactUserID: db.ContactUserID,
		Pseudonym:     db.Pseudonym,
		CreatedAt:     db.CreatedAt,
		UpdatedAt:     db.UpdatedAt,
		Profile:       prof,
	}
}
