package register

import (
	"strings"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
)

func (r *RegisterRequest) IsValid() []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail

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

	if err := tools.ValidateEmail(r.Email); err != nil {
		details = append(details, *err)
	}

	if err := tools.ValidatePassword(r.Password); err != nil {
		details = append(details, *err)
	}

	if err := tools.ValidatePhoneNumber(r.PhoneNumber); err != nil {
		details = append(details, *err)
	}

	if err := tools.ValidateNickname(r.Nickname); err != nil {
		details = append(details, *err)
	}

	if err := tools.ValidateFirstName(r.FirstName); err != nil {
		details = append(details, *err)
	}

	if err := tools.ValidateLastName(r.LastName); err != nil {
		details = append(details, *err)
	}

	return details
}
