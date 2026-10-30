package session

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepositoryImpl {
	return &SessionRepositoryImpl{pool: pool}
}

func (r *SessionRepositoryImpl) CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error) {
	const query = `insert into session (user_id, expires_at) values ($1, $2) returning id, user_id, expires_at, created_at, updated_at`
	var db SessionDB
	err := r.pool.QueryRow(ctx, query, userID, expiresAt).Scan(&db.ID, &db.UserID, &db.ExpiresAt, &db.CreatedAt, &db.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return SessionDBToModel(&db), nil
}

func (r *SessionRepositoryImpl) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	const query = `delete from session where id = $1`
	_, err := r.pool.Exec(ctx, query, sessionID)
	return err
}

func (r *SessionRepositoryImpl) GetSessionUser(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	const query = `select user_id from session where id = $1 and expires_at > now()`
	var userID uuid.UUID
	err := r.pool.QueryRow(ctx, query, sessionID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, apperrors.ErrSessionNotFound
	}
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (r *SessionRepositoryImpl) ExtendSession(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	const query = `update session set expires_at = $2, updated_at = now() where id = $1 and expires_at > now()`
	tag, err := r.pool.Exec(ctx, query, sessionID, expiresAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrSessionNotFound
	}
	return nil
}
