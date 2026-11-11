package validation

import (
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/delivery/httperrors"
	"github.com/google/uuid"
)

func TestValidators(t *testing.T) {
	longEmail := strings.Repeat("a", MaxEmailLen+1)
	longPassword := strings.Repeat("a", MaxPasswordLen+1)
	longNickname := strings.Repeat("a", MaxNicknameLen+1)
	longFirst := strings.Repeat("a", MaxFirstNameLen+1)
	longLast := strings.Repeat("a", MaxLastNameLen+1)
	longChatName := strings.Repeat("a", MaxChatNameLen+1)
	longDescription := strings.Repeat("a", MaxChatDescriptionLen+1)
	longMessage := strings.Repeat("a", MaxMessageContentLen+1)
	longPseudonym := strings.Repeat("a", MaxPseudonymLen+1)
	phone := "+79991234567"
	badPhone := "abc"
	last := "Иванов"
	shortLast := "Я"
	text := "hello"
	sticker := uuid.New()

	cases := []struct {
		name      string
		got       *httperrors.ErrorDetail
		wantValid bool
	}{
		{"email valid", ValidateEmail("a@example.org"), true}, {"email required", ValidateEmail(""), false}, {"email too long", ValidateEmail(longEmail), false}, {"email invalid", ValidateEmail("not-an-email"), false},
		{"password valid", ValidatePassword("12345678"), true}, {"password required", ValidatePassword(""), false}, {"password length", ValidatePassword(longPassword), false},
		{"phone nil", ValidatePhoneNumber(nil), true}, {"phone valid", ValidatePhoneNumber(&phone), true}, {"phone invalid", ValidatePhoneNumber(&badPhone), false},
		{"nickname valid", ValidateNickname("user_123"), true}, {"nickname required", ValidateNickname(""), false}, {"nickname length", ValidateNickname("ab"), false}, {"nickname format", ValidateNickname("bad-name"), false}, {"nickname too long", ValidateNickname(longNickname), false},
		{"first name valid", ValidateFirstName(" Мария "), true}, {"first name required", ValidateFirstName(" "), false}, {"first name length", ValidateFirstName("Я"), false}, {"first name digits", ValidateFirstName("Иван2"), false}, {"first name too long", ValidateFirstName(longFirst), false},
		{"last name nil", ValidateLastName(nil), true}, {"last name blank", ValidateLastName(ptr("  ")), true}, {"last name valid", ValidateLastName(&last), true}, {"last name short", ValidateLastName(&shortLast), false}, {"last name long", ValidateLastName(&longLast), false},
		{"chat name valid", ValidateChatName("general"), true}, {"chat name blank", ValidateChatName("  "), false}, {"chat name long", ValidateChatName(longChatName), false},
		{"description nil", ValidateChatDescription(nil), true}, {"description valid", ValidateChatDescription(&text), true}, {"description long", ValidateChatDescription(&longDescription), false},
		{"text valid", ValidateMessageFields("text", &text, nil), true}, {"text missing", ValidateMessageFields("text", nil, nil), false}, {"text blank", ValidateMessageFields("text", ptr(" "), nil), false}, {"text too long", ValidateMessageFields("text", &longMessage, nil), false}, {"text with sticker", ValidateMessageFields("text", &text, &sticker), false},
		{"sticker valid", ValidateMessageFields("sticker", nil, &sticker), true}, {"sticker missing", ValidateMessageFields("sticker", nil, nil), false}, {"sticker with content", ValidateMessageFields("sticker", &text, &sticker), false}, {"unknown message", ValidateMessageFields("image", &text, nil), false},
		{"pseudonym valid", ValidatePseudonym(&text), true}, {"pseudonym too long", ValidatePseudonym(&longPseudonym), false}, {"pseudonym nil", ValidatePseudonym(nil), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if gotValid := tc.got == nil; gotValid != tc.wantValid {
				t.Fatalf("got valid=%t, want %t (detail: %#v)", gotValid, tc.wantValid, tc.got)
			}
		})
	}
}

func ptr(s string) *string { return &s }
