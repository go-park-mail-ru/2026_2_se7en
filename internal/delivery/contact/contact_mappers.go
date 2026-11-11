package contact

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
)

func ContactToResponse(c models.Contact) ContactResponse {
	return ContactResponse{
		UserID:        c.UserID,
		ContactUserID: c.ContactUserID,
		Pseudonym:     c.Pseudonym,
		Nickname:      c.Profile.Nickname,
		FirstName:     c.Profile.FirstName,
		LastName:      c.Profile.LastName,
		IconID:        c.Profile.IconID,
		IconURL:       c.Profile.AvatarURL,
		IsOnline:      c.IsOnline,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

func ContactsToResponse(contacts []models.Contact) []ContactResponse {
	result := make([]ContactResponse, 0, len(contacts))
	for _, c := range contacts {
		result = append(result, ContactToResponse(c))
	}
	return result
}
