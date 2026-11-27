package csrf

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/google/uuid"
)

func TestHashTokenValidation(t *testing.T) {
	tokenizer, err := NewHMACHashToken("test-secret-with-at-least-32-bytes")
	if err != nil {
		t.Fatal(err)
	}
	session := &models.Session{ID: uuid.New(), UserID: uuid.New()}
	expires := time.Now().Add(time.Hour).Unix()
	token := tokenizer.Create(session, expires)

	if err := tokenizer.Check(session, token); err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	if err := tokenizer.Check(session, tokenizer.Create(session, time.Now().Add(-time.Hour).Unix())); err == nil {
		t.Fatal("expired token accepted")
	}
	if err := tokenizer.Check(session, "broken"); err == nil {
		t.Fatal("malformed token accepted")
	}
	if err := tokenizer.Check(&models.Session{ID: session.ID, UserID: uuid.New()}, token); err == nil {
		t.Fatal("token for another user accepted")
	}
	if _, err := NewHMACHashToken("short"); err == nil {
		t.Fatal("short secret accepted")
	}
}
