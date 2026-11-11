package session

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/usecase/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
)

func TestSessionRepository(t *testing.T) {
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	repo := NewSessionRepository(pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx := context.Background()
	userID, sessionID := uuid.New(), uuid.New()
	now, expiry := time.Now(), time.Now().Add(time.Hour)

	pool.ExpectQuery("insert into session").WithArgs(userID, expiry).WillReturnRows(
		pgxmock.NewRows([]string{"id", "user_id", "expires_at", "created_at", "updated_at"}).AddRow(sessionID, userID, expiry, now, now),
	)
	created, err := repo.CreateSession(ctx, userID, expiry)
	if err != nil || created.ID != sessionID {
		t.Fatalf("CreateSession() = %#v, %v", created, err)
	}

	pool.ExpectExec("delete from session").WithArgs(sessionID).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := repo.DeleteSession(ctx, sessionID); err != nil {
		t.Fatal(err)
	}
	pool.ExpectQuery("select s.user_id").WithArgs(sessionID).WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(userID))
	gotUser, err := repo.GetSessionUser(ctx, sessionID)
	if err != nil || gotUser.UserID != userID {
		t.Fatalf("GetSessionUser() = %#v, %v", gotUser, err)
	}
	pool.ExpectQuery("select s.user_id").WithArgs(sessionID).WillReturnError(pgx.ErrNoRows)
	if _, err := repo.GetSessionUser(ctx, sessionID); err != apperrors.ErrSessionNotFound {
		t.Fatalf("not found error = %v", err)
	}
	pool.ExpectExec("update session").WithArgs(sessionID, expiry).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repo.ExtendSession(ctx, sessionID, expiry); err != nil {
		t.Fatal(err)
	}
	pool.ExpectExec("update session").WithArgs(sessionID, expiry).WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	if err := repo.ExtendSession(ctx, sessionID, expiry); err != apperrors.ErrSessionNotFound {
		t.Fatalf("missing extend error = %v", err)
	}
	if err := pool.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
