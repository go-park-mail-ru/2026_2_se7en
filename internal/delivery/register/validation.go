package register

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"strings"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/validation"
)

func (r *RegisterRequest) IsValid() []httperrors.ErrorDetail {
	var details []httperrors.ErrorDetail

	if r.PhoneNumber != nil {
		p := strings.TrimSpace(*r.PhoneNumber)
		if p == "" {
			r.PhoneNumber = nil
		} else {
			r.PhoneNumber = &p
		}
	}

	r.FirstName = strings.TrimSpace(r.FirstName)
	if r.LastName != nil {
		l := strings.TrimSpace(*r.LastName)
		if l == "" {
			r.LastName = nil
		} else {
			r.LastName = &l
		}
	}

	if err := validation.ValidateEmail(r.Email); err != nil {
		details = append(details, *err)
	}

	if err := validation.ValidatePassword(r.Password); err != nil {
		details = append(details, *err)
	}

	if err := validation.ValidatePhoneNumber(r.PhoneNumber); err != nil {
		details = append(details, *err)
	}

	if err := validation.ValidateNickname(r.Nickname); err != nil {
		details = append(details, *err)
	}

	if err := validation.ValidateFirstName(r.FirstName); err != nil {
		details = append(details, *err)
	}

	if err := validation.ValidateLastName(r.LastName); err != nil {
		details = append(details, *err)
	}

	return details
}
