package user

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type loggingPool struct {
	storage.DBPool
	err error
}

func (p loggingPool) QueryRow(context.Context, string, ...any) pgx.Row { return loggingRow{err: p.err} }

type loggingRow struct{ err error }

func (r loggingRow) Scan(...any) error { return r.err }

func TestDatabaseLogsCarryRequestAndLayer(t *testing.T) {
	var output bytes.Buffer
	log, err := logger.New(&output, logger.LevelDebug)
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("database unavailable")
	repo := NewUserRepository(loggingPool{err: failure}, log)
	ctx := logger.WithUserID(logger.WithRequestID(context.Background(), "db-request"), "caller")
	if _, err := repo.FindByID(ctx, uuid.New()); !errors.Is(err, failure) {
		t.Fatalf("error=%v", err)
	}
	entries := strings.Split(strings.TrimSpace(output.String()), "\n")
	count := len(entries)
	for _, entry := range entries {
		for _, field := range []string{"layer=repository", "module=user", "request_id=db-request", "user_id=caller"} {
			if !strings.Contains(entry, field) {
				t.Fatalf("missing database log context: %s", entry)
			}
		}
	}
	if count != 2 {
		t.Fatalf("got %d records, want database operation and error", count)
	}
}
