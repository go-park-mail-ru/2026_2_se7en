package profile

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"strings"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/validation"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
)

func ToProfileResponse(user *models.User, prof *models.Profile) ProfileResponse {
	resp := ProfileResponse{
		ID:          user.ID,
		Email:       user.Email,
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

func ToPublicProfileResponse(user *models.User, prof *models.Profile) PublicProfileResponse {
	return PublicProfileResponse{
		ID:        user.ID,
		Nickname:  prof.Nickname,
		FirstName: prof.FirstName,
		LastName:  prof.LastName,
		Bio:       prof.Bio,
		IconID:    prof.IconID,
		IconURL:   prof.AvatarURL,
	}
}

func (r *UpdateProfileRequest) IsValid() []httperrors.ErrorDetail {
	var details []httperrors.ErrorDetail
	if r.Email != nil {
		e := strings.ToLower(strings.TrimSpace(*r.Email))
		r.Email = &e
		if err := validation.ValidateEmail(*r.Email); err != nil {
			details = append(details, *err)
		}
	}

	if r.Nickname != nil {
		if err := validation.ValidateNickname(*r.Nickname); err != nil {
			details = append(details, *err)
		}
	}

	if r.FirstName != nil {
		if err := validation.ValidateFirstName(*r.FirstName); err != nil {
			details = append(details, *err)
		}
	}

	if r.LastName != nil {
		if err := validation.ValidateLastName(r.LastName); err != nil {
			details = append(details, *err)
		}
	}

	if r.PhoneNumber != nil {
		p := strings.TrimSpace(*r.PhoneNumber)
		r.PhoneNumber = &p
		if err := validation.ValidatePhoneNumber(r.PhoneNumber); err != nil {
			details = append(details, *err)
		}
	}

	return details
}

func (r *UpdatePasswordRequest) IsValid() []httperrors.ErrorDetail {
	var details []httperrors.ErrorDetail

	if r.OldPassword == "" {
		details = append(details, httperrors.ErrorDetail{Field: "old_password", Reason: "required"})
	}

	if err := validation.ValidatePassword(r.NewPassword); err != nil {
		details = append(details, *err)
	}

	return details
}
