package users

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

func UserProfilesToResponse(profiles []models.UserSearchResult, isOnline func(uuid.UUID) bool) []UserSearchResponse {
	result := make([]UserSearchResponse, 0, len(profiles))
	for _, item := range profiles {
		profile := item.Profile
		result = append(result, UserSearchResponse{
			UserID: profile.UserID, Nickname: profile.Nickname, FirstName: profile.FirstName,
			LastName: profile.LastName, IconID: profile.IconID, IconURL: profile.AvatarURL,
			IsOnline: isOnline != nil && isOnline(profile.UserID), IsContact: item.IsContact,
		})
	}
	return result
}
