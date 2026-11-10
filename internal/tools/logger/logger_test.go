package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func record(t *testing.T, buffer *bytes.Buffer) map[string]any {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buffer.Bytes()), &value); err != nil {
		t.Fatalf("invalid JSON log: %v: %s", err, buffer.String())
	}
	return value
}

func TestContextAndServiceFields(t *testing.T) {
	var output bytes.Buffer
	base, err := newJSONTestLogger(&output, LevelDebug)
	if err != nil {
		t.Fatal(err)
	}
	scoped := Scope(base, LayerUsecase, ModuleChat)
	ctx := WithUserID(WithRequestID(context.Background(), "request-one"), "user-one")
	scoped.InfoContext(ctx, "group created", KeyChatID, "chat-one")
	got := record(t, &output)
	for key, want := range map[string]string{KeyLayer: LayerUsecase, KeyModule: ModuleChat, KeyRequestID: "request-one", KeyUserID: "user-one", KeyChatID: "chat-one"} {
		if got[key] != want {
			t.Errorf("%s=%v, want %s", key, got[key], want)
		}
	}
	output.Reset()
	scoped.InfoContext(context.Background(), "another operation")
	got = record(t, &output)
	if _, ok := got[KeyRequestID]; ok {
		t.Error("request metadata leaked into another context")
	}
}

func TestSensitiveAttributesAndDatabaseDetailsAreRedacted(t *testing.T) {
	var output bytes.Buffer
	log, _ := newJSONTestLogger(&output, LevelDebug)
	dbErr := &pgconn.PgError{Code: "23505", ConstraintName: "uq_profile_nickname", Message: "secret-message", Detail: "secret-row-value", InternalQuery: "secret-query"}
	scoped := log.With("password", "secret-password", "email", "alice@example.com", "phone_number", "+79991234567")
	scoped.Info("operation", "cookie", "secret-cookie", "session_id", "secret-session", "access_token", "secret-token", "content", "secret-message-body", "error", dbErr,
		slog.Group("nested", slog.String("password_hash", "secret-hash"), slog.String("authorization", "secret-auth")))
	got := record(t, &output)
	for _, secret := range []string{"secret-password", "alice@example.com", "+79991234567", "secret-cookie", "secret-session", "secret-token", "secret-message-body", "secret-hash", "secret-auth", "secret-row-value", "secret-message", "secret-query"} {
		if strings.Contains(output.String(), secret) {
			t.Errorf("secret appeared in log: %s", secret)
		}
	}
	if got["email"] != MaskEmail("alice@example.com") {
		t.Error("email was not masked")
	}
	db := got[KeyError].(map[string]any)
	if db[KeyCode] != "23505" || db[KeyConstraint] != "uq_profile_nickname" {
		t.Fatalf("database diagnostics missing: %v", db)
	}
}

func TestErrorDoesNotExposeUnknownErrorText(t *testing.T) {
	var output bytes.Buffer
	log, _ := newJSONTestLogger(&output, LevelDebug)
	log.ErrorContext(context.Background(), "connection failed", Error(errors.New("postgres://name:password@host/db")))
	if strings.Contains(output.String(), "password") || strings.Contains(output.String(), "postgres://") {
		t.Fatal("connection credentials exposed")
	}
}

func TestLevelsAndTextOutput(t *testing.T) {
	for _, level := range []string{LevelDebug, LevelInfo, LevelWarn, LevelError} {
		t.Run(level, func(t *testing.T) {
			var output bytes.Buffer
			log, err := newJSONTestLogger(&output, level)
			if err != nil {
				t.Fatal(err)
			}
			log.Debug("debug entry")
			if (level == LevelDebug) != (output.Len() > 0) {
				t.Fatalf("wrong debug filtering for %s", level)
			}
			output.Reset()
			log.Error("error entry")
			if record(t, &output)["level"] != "ERROR" {
				t.Fatal("error was filtered")
			}
		})
	}
	var output bytes.Buffer
	log, err := New(&output, LevelInfo)
	if err != nil {
		t.Fatal(err)
	}
	log.Info("text entry")
	if !strings.Contains(output.String(), "level=INFO") || strings.HasPrefix(output.String(), "{") {
		t.Fatal("text format not used")
	}
	if _, err := New(&output, "unsupported"); err == nil {
		t.Error("invalid level accepted")
	}
}

func TestMasksSupportUnicodeAndMissingValues(t *testing.T) {
	if got := MaskEmail("юлия@example.com"); got != "ю***@***.com" {
		t.Fatalf("mask=%q", got)
	}
	if MaskEmail("invalid") != "***" || MaskPhone(nil) != "***" {
		t.Fatal("missing value was not masked")
	}
}

func newJSONTestLogger(out io.Writer, level string) (*slog.Logger, error) {
	parsed, err := ParseLevel(level)
	if err != nil {
		return nil, err
	}
	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{Level: parsed, ReplaceAttr: sanitize})
	return slog.New(&contextHandler{next: handler}), nil
}
