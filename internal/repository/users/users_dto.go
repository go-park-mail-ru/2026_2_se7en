package users

import profileRepo "github.com/go-park-mail-ru/2026_2_se7en/internal/repository/profile"

type UserSearchDB struct {
	Profile   profileRepo.ProfileDB
	AvatarURL *string
	IsContact bool
}
