package users

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	profileRepo "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"
)

func UserSearchDBToModel(db *UserSearchDB) models.UserSearchResult {
	profile := profileRepo.ProfileDBToModel(&db.Profile)
	profile.AvatarURL = db.AvatarURL
	return models.UserSearchResult{Profile: *profile, IsContact: db.IsContact}
}
