package register

import (
	"net/mail"
	"strings"
	"unicode/utf8"

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

	if r.Email == "" {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "required"})
	} else if len(r.Email) > 255 {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "too_long"})
	} else if _, err := mail.ParseAddress(r.Email); err != nil {
		details = append(details, apperrors.ErrorDetail{Field: "email", Reason: "invalid_format"})
	}

	if r.Password == "" {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "required"})
	} else if len(r.Password) < 8 || len(r.Password) > 16 {
		details = append(details, apperrors.ErrorDetail{Field: "password", Reason: "invalid_length"})
	}

	if r.PhoneNumber != nil && !tools.IsPhoneNumberValid(*r.PhoneNumber) {
		details = append(details, apperrors.ErrorDetail{Field: "phone_number", Reason: "invalid_format"})
	}

	if r.Nickname == "" {
		details = append(details, apperrors.ErrorDetail{Field: "nickname", Reason: "required"})
	} else if len(r.Nickname) < 3 || len(r.Nickname) > 16 || !tools.IsNicknameValid(r.Nickname) {
		details = append(details, apperrors.ErrorDetail{Field: "nickname", Reason: "invalid_format"})
	}

	fnLen := utf8.RuneCountInString(r.FirstName)
	if fnLen == 0 {
		details = append(details, apperrors.ErrorDetail{Field: "first_name", Reason: "required"})
	} else if strings.ContainsAny(r.FirstName, "0123456789") || fnLen < 2 || fnLen > 32 {
		details = append(details, apperrors.ErrorDetail{Field: "first_name", Reason: "invalid_format"})
	}

	if r.LastName != nil {
		lnLen := utf8.RuneCountInString(*r.LastName)
		if lnLen > 0 && (lnLen < 2 || lnLen > 32) {
			details = append(details, apperrors.ErrorDetail{Field: "last_name", Reason: "invalid_length"})
		}
	}

	return details
}
