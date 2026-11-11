package contact

import (
	"strings"
	"testing"
)

func TestCreateContactRequestValidation(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if details := (&CreateContactRequest{Nickname: "friend"}).IsValid(); len(details) != 0 {
			t.Fatalf("details = %+v", details)
		}
	})
	t.Run("invalid nickname and pseudonym", func(t *testing.T) {
		pseudonym := strings.Repeat("x", 101)
		details := (&CreateContactRequest{Nickname: "!", Pseudonym: &pseudonym}).IsValid()
		if len(details) != 2 {
			t.Fatalf("details = %+v, want both field errors", details)
		}
	})
}
