package auth

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"strings"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/validation"
)

func (r *LoginRequest) IsValid() []httperrors.ErrorDetail {
	var details []httperrors.ErrorDetail

	r.Email = strings.ToLower(strings.TrimSpace(r.Email))

	if err := validation.ValidateEmail(r.Email); err != nil {
		details = append(details, *err)
	}

	if err := validation.ValidatePassword(r.Password); err != nil {
		details = append(details, *err)
	}

	return details
}
