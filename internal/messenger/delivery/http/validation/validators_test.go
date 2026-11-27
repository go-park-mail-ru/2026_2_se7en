package validation

import (
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/infrastructure/http/httperrors"
	"github.com/google/uuid"
)

func TestChatValidators(t *testing.T) {
	longName := strings.Repeat("a", MaxChatNameLen+1)
	longDescription := strings.Repeat("a", MaxChatDescriptionLen+1)
	longMessage := strings.Repeat("a", MaxMessageContentLen+1)
	text := "hello"
	sticker := uuid.New()
	cases := []struct {
		name      string
		detail    *httperrors.ErrorDetail
		wantValid bool
	}{
		{"chat name valid", ValidateChatName("general"), true},
		{"chat name blank", ValidateChatName("  "), false},
		{"chat name long", ValidateChatName(longName), false},
		{"description nil", ValidateChatDescription(nil), true},
		{"description valid", ValidateChatDescription(&text), true},
		{"description long", ValidateChatDescription(&longDescription), false},
		{"text valid", ValidateMessageFields("text", &text, nil), true},
		{"text missing", ValidateMessageFields("text", nil, nil), false},
		{"text blank", ValidateMessageFields("text", stringPointer(" "), nil), false},
		{"text too long", ValidateMessageFields("text", &longMessage, nil), false},
		{"text with sticker", ValidateMessageFields("text", &text, &sticker), false},
		{"sticker valid", ValidateMessageFields("sticker", nil, &sticker), true},
		{"sticker missing", ValidateMessageFields("sticker", nil, nil), false},
		{"sticker with content", ValidateMessageFields("sticker", &text, &sticker), false},
		{"unknown message", ValidateMessageFields("image", &text, nil), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if gotValid := tc.detail == nil; gotValid != tc.wantValid {
				t.Fatalf("got valid=%t, want %t (detail: %#v)", gotValid, tc.wantValid, tc.detail)
			}
		})
	}
}

func stringPointer(value string) *string { return &value }
