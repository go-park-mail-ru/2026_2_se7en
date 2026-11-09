package session

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SessionRepositoryImpl struct {
	pool storage.DBPool
}

func NewSessionRepository(pool storage.DBPool) *SessionRepositoryImpl {
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
	const query = `select s.user_id from session s join app_user u on u.id = s.user_id and u.deleted_at is null where s.id = $1 and s.expires_at > now()`
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
	const query = `update session s set expires_at = $2, updated_at = now() where s.id = $1 and s.expires_at > now() and exists (select 1 from app_user u where u.id = s.user_id and u.deleted_at is null)`
	tag, err := r.pool.Exec(ctx, query, sessionID, expiresAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrSessionNotFound
	}
	return nil
}
