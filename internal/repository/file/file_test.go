package file

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
)

func TestFileRepositoryCreate(t *testing.T) {
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := NewFileRepository(pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	now := time.Now()
	id := uuid.New()
	pool.ExpectQuery("insert into file").WithArgs("/file.png", "avatar.png", "image/png").WillReturnRows(
		pgxmock.NewRows([]string{"id", "url", "filename", "type", "created_at", "updated_at"}).AddRow(id, "/file.png", "avatar.png", "image/png", now, now),
	)
	file, err := repo.Create(context.Background(), "/file.png", "avatar.png", "image/png")
	if err != nil || file.ID != id || file.Type != "image/png" {
		t.Fatalf("Create() = %#v, %v", file, err)
	}

	pool.ExpectQuery("insert into file").WithArgs("u", "f", "t").WillReturnError(pgx.ErrNoRows)
	if _, err := repo.Create(context.Background(), "u", "f", "t"); err == nil {
		t.Fatal("expected database error")
	}
	if err := pool.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
