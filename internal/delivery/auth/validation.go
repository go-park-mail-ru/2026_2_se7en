package auth

import (
	"net/mail"
	"strings"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
)

func (r *LoginRequest) IsValid() []apperrors.ErrorDetail {
	var details []apperrors.ErrorDetail
	email := strings.ToLower(strings.TrimSpace(r.Email))
	r.Email = email

	if r.Email == "" {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "required"})
	} else if _, err := mail.ParseAddress(r.Email); err != nil {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "invalid_format"})
	}

	if r.Password == "" {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "required"})
	} else if len(r.Password) < 8 || len(r.Password) > 16 {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "invalid_length"})
	}

	return details
}
