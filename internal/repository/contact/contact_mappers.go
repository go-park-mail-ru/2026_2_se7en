package contact

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
)

func ContactDBToModel(db *ContactDB) models.Contact {
	return models.Contact{
		UserID:        db.UserID,
		ContactUserID: db.ContactUserID,
		Pseudonym:     db.Pseudonym,
		CreatedAt:     db.CreatedAt,
		UpdatedAt:     db.UpdatedAt,
		Profile:       profile.ProfileDBToModel(&db.Profile),
		IsOnline:      db.IsOnline,
	}
}
