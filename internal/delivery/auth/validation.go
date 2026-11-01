package auth

import (
	"strings"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools"
)

func (r *LoginRequest) IsValid() []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail

	r.Email = strings.ToLower(strings.TrimSpace(r.Email))

	if err := tools.ValidateEmail(r.Email); err != nil {
		details = append(details, *err)
	}

	if err := tools.ValidatePassword(r.Password); err != nil {
		details = append(details, *err)
	}

	return details
}
