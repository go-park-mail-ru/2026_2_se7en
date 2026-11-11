package tx

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/pashagolub/pgxmock/v5"
)

func TestManagerWithinTransaction(t *testing.T) {
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	manager := NewManager(pool, slog.New(slog.NewTextHandler(io.Discard, nil)))

	pool.ExpectBegin()
	pool.ExpectCommit()
	err = manager.WithinTransaction(context.Background(), func(ctx context.Context) error {
		if _, ok := FromContext(ctx); !ok {
			t.Fatal("transaction not added to callback context")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("successful transaction returned %v", err)
	}

	callbackErr := errors.New("callback failed")
	pool.ExpectBegin()
	pool.ExpectRollback()
	err = manager.WithinTransaction(context.Background(), func(context.Context) error { return callbackErr })
	if !errors.Is(err, callbackErr) {
		t.Fatalf("callback error lost: %v", err)
	}

	beginErr := errors.New("begin failed")
	pool.ExpectBegin().WillReturnError(beginErr)
	if err := manager.WithinTransaction(context.Background(), func(context.Context) error { return nil }); err == nil {
		t.Fatal("expected begin error")
	}

	if err := pool.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
