package validation

import (
	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MinPasswordLen        = 8
	MaxPasswordLen        = 16
	MinNicknameLen        = 3
	MaxNicknameLen        = 16
	MaxEmailLen           = 64
	MinFirstNameLen       = 2
	MaxFirstNameLen       = 32
	MinLastNameLen        = 2
	MaxLastNameLen        = 32
	MaxChatNameLen        = 100
	MaxChatDescriptionLen = 500
	MaxMessageContentLen  = 4096
	MaxPseudonymLen       = 64
)

var (
	phoneNumberRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	nicknameRegex    = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func ValidateEmail(email string) *httperrors.ErrorDetail {
	if email == "" {
		return &httperrors.ErrorDetail{Field: "email", Reason: "required"}
	}
	if len(email) > MaxEmailLen {
		return &httperrors.ErrorDetail{Field: "email", Reason: "too_long"}
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return &httperrors.ErrorDetail{Field: "email", Reason: "invalid_format"}
	}
	return nil
}

func ValidatePassword(password string) *httperrors.ErrorDetail {
	if password == "" {
		return &httperrors.ErrorDetail{Field: "password", Reason: "required"}
	}
	if len(password) < MinPasswordLen || len(password) > MaxPasswordLen {
		return &httperrors.ErrorDetail{Field: "password", Reason: "invalid_length"}
	}
	return nil
}

func ValidatePhoneNumber(phone *string) *httperrors.ErrorDetail {
	if phone == nil {
		return nil
	}
	p := strings.TrimSpace(*phone)
	if p == "" {
		return nil
	}
	if !phoneNumberRegex.MatchString(p) {
		return &httperrors.ErrorDetail{Field: "phone_number", Reason: "invalid_format"}
	}
	return nil
}

func ValidateNickname(nickname string) *httperrors.ErrorDetail {
	if nickname == "" {
		return &httperrors.ErrorDetail{Field: "nickname", Reason: "required"}
	}
	if len(nickname) < MinNicknameLen || len(nickname) > MaxNicknameLen {
		return &httperrors.ErrorDetail{Field: "nickname", Reason: "invalid_length"}
	}
	if !nicknameRegex.MatchString(nickname) {
		return &httperrors.ErrorDetail{Field: "nickname", Reason: "invalid_format"}
	}
	return nil
}

func ValidateFirstName(firstName string) *httperrors.ErrorDetail {
	firstName = strings.TrimSpace(firstName)
	if firstName == "" {
		return &httperrors.ErrorDetail{Field: "first_name", Reason: "required"}
	}
	fnLen := utf8.RuneCountInString(firstName)
	if fnLen < MinFirstNameLen || fnLen > MaxFirstNameLen {
		return &httperrors.ErrorDetail{Field: "first_name", Reason: "invalid_length"}
	}
	if strings.ContainsAny(firstName, "0123456789") {
		return &httperrors.ErrorDetail{Field: "first_name", Reason: "contains_digits"}
	}
	return nil
}

func ValidateLastName(lastName *string) *httperrors.ErrorDetail {
	if lastName == nil {
		return nil
	}
	ln := strings.TrimSpace(*lastName)
	if ln == "" {
		return nil
	}
	lnLen := utf8.RuneCountInString(ln)
	if lnLen < MinLastNameLen || lnLen > MaxLastNameLen {
		return &httperrors.ErrorDetail{Field: "last_name", Reason: "invalid_length"}
	}
	return nil
}

func ValidateChatName(name string) *httperrors.ErrorDetail {
	if strings.TrimSpace(name) == "" {
		return &httperrors.ErrorDetail{Field: "name", Reason: "Name is required"}
	}
	if utf8.RuneCountInString(name) > MaxChatNameLen {
		return &httperrors.ErrorDetail{Field: "name", Reason: "Name must be at most 100 characters"}
	}
	return nil
}

func ValidateChatDescription(description *string) *httperrors.ErrorDetail {
	if description != nil && utf8.RuneCountInString(*description) > MaxChatDescriptionLen {
		return &httperrors.ErrorDetail{Field: "description", Reason: "Description must be at most 500 characters"}
	}
	return nil
}

func ValidateMessageFields(messageType string, content *string, stickerID *uuid.UUID) *httperrors.ErrorDetail {
	switch messageType {
	case "text":
		if content == nil || strings.TrimSpace(*content) == "" {
			return &httperrors.ErrorDetail{Field: "content", Reason: "Text message requires content"}
		}
		if utf8.RuneCountInString(*content) > MaxMessageContentLen {
			return &httperrors.ErrorDetail{Field: "content", Reason: "Content must be at most 4096 characters"}
		}
		if stickerID != nil {
			return &httperrors.ErrorDetail{Field: "sticker_id", Reason: "Text message cannot contain a sticker"}
		}
	case "sticker":
		if stickerID == nil {
			return &httperrors.ErrorDetail{Field: "sticker_id", Reason: "Sticker message requires a sticker"}
		}
		if content != nil {
			return &httperrors.ErrorDetail{Field: "content", Reason: "Sticker message cannot contain text"}
		}
	default:
		return &httperrors.ErrorDetail{Field: "type", Reason: "Message type must be text or sticker"}
	}
	return nil
}

func ValidatePseudonym(pseudonym *string) *httperrors.ErrorDetail {
	if pseudonym != nil && len(*pseudonym) > MaxPseudonymLen {
		return &httperrors.ErrorDetail{Field: "pseudonym", Reason: "too_long"}
	}

	return nil
}
