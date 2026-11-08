package profile

import (
	"strings"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
)

func ToProfileResponse(user *models.User, prof *models.Profile) ProfileResponse {
	resp := ProfileResponse{
		ID:    user.ID,
		Email: user.Email,
		PhoneNumber: user.PhoneNumber,
	}
	
	if prof != nil {
		resp.Nickname = prof.Nickname
		resp.FirstName = prof.FirstName
		resp.LastName = prof.LastName
		resp.Bio = prof.Bio
		resp.IconID = prof.IconID
		if prof.AvatarURL != nil {
			resp.IconURL = prof.AvatarURL
		}
	}

	return resp
}

func (r *UpdateProfileRequest) IsValid() []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail
	if r.Email != nil {
		e := strings.ToLower(strings.TrimSpace(*r.Email))
		r.Email = &e
		if err := tools.ValidateEmail(*r.Email); err != nil {
			details = append(details, *err)
		}
	}

	if r.Nickname != nil {
		if err := tools.ValidateNickname(*r.Nickname); err != nil {
			details = append(details, *err)
		}
	}

	if r.FirstName != nil {
		if err := tools.ValidateFirstName(*r.FirstName); err != nil {
			details = append(details, *err)
		}
	}

	if r.LastName != nil {
		if err := tools.ValidateLastName(r.LastName); err != nil {
			details = append(details, *err)
		}
	}

	if r.PhoneNumber != nil {
		p := strings.TrimSpace(*r.PhoneNumber)
		r.PhoneNumber = &p
		if err := tools.ValidatePhoneNumber(r.PhoneNumber); err != nil {
			details = append(details, *err)
		}
	}

	return details
}
