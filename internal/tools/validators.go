package tools

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
)

const (
	MinPasswordLen  = 8
	MaxPasswordLen  = 16
	MinNicknameLen  = 3
	MaxNicknameLen  = 16
	MaxEmailLen     = 64
	MinFirstNameLen = 2
	MaxFirstNameLen = 32
	MinLastNameLen  = 2
	MaxLastNameLen  = 32
	MaxPseudonymLen = 64
)

var (
	phoneNumberRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	nicknameRegex    = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func ValidateEmail(email string) *apperrors.ErrorDetail {
	if email == "" {
		return &apperrors.ErrorDetail{Field: "email", Reason: "required"}
	}
	if len(email) > MaxEmailLen {
		return &apperrors.ErrorDetail{Field: "email", Reason: "too_long"}
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return &apperrors.ErrorDetail{Field: "email", Reason: "invalid_format"}
	}
	return nil
}

func ValidatePassword(password string) *apperrors.ErrorDetail {
	if password == "" {
		return &apperrors.ErrorDetail{Field: "password", Reason: "required"}
	}
	if len(password) < MinPasswordLen || len(password) > MaxPasswordLen {
		return &apperrors.ErrorDetail{Field: "password", Reason: "invalid_length"}
	}
	return nil
}

func ValidatePhoneNumber(phone *string) *apperrors.ErrorDetail {
	if phone == nil {
		return nil
	}
	p := strings.TrimSpace(*phone)
	if p == "" {
		return nil
	}
	if !phoneNumberRegex.MatchString(p) {
		return &apperrors.ErrorDetail{Field: "phone_number", Reason: "invalid_format"}
	}
	return nil
}

func ValidateNickname(nickname string) *apperrors.ErrorDetail {
	if nickname == "" {
		return &apperrors.ErrorDetail{Field: "nickname", Reason: "required"}
	}
	if len(nickname) < MinNicknameLen || len(nickname) > MaxNicknameLen {
		return &apperrors.ErrorDetail{Field: "nickname", Reason: "invalid_length"}
	}
	if !nicknameRegex.MatchString(nickname) {
		return &apperrors.ErrorDetail{Field: "nickname", Reason: "invalid_format"}
	}
	return nil
}

func ValidateFirstName(firstName string) *apperrors.ErrorDetail {
	firstName = strings.TrimSpace(firstName)
	if firstName == "" {
		return &apperrors.ErrorDetail{Field: "first_name", Reason: "required"}
	}
	fnLen := utf8.RuneCountInString(firstName)
	if fnLen < MinFirstNameLen || fnLen > MaxFirstNameLen {
		return &apperrors.ErrorDetail{Field: "first_name", Reason: "invalid_length"}
	}
	if strings.ContainsAny(firstName, "0123456789") {
		return &apperrors.ErrorDetail{Field: "first_name", Reason: "contains_digits"}
	}
	return nil
}

func ValidateLastName(lastName *string) *apperrors.ErrorDetail {
	if lastName == nil {
		return nil
	}
	ln := strings.TrimSpace(*lastName)
	if ln == "" {
		return nil
	}
	lnLen := utf8.RuneCountInString(ln)
	if lnLen < MinLastNameLen || lnLen > MaxLastNameLen {
		return &apperrors.ErrorDetail{Field: "last_name", Reason: "invalid_length"}
	}
	return nil
}

func ValidatePseudonym(pseudonym *string) *apperrors.ErrorDetail {
	if pseudonym != nil && len(*pseudonym) > MaxPseudonymLen {
		return &apperrors.ErrorDetail{Field: "pseudonym", Reason: "too_long"}
	}

	return nil
}
