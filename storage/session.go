package storage

import (
	apperrors "app/app_errors"
	"app/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	sqlCreateSession = `
		insert into session (user_id, expires_at)
	 	values ($1, $2)
		returning (id, user_id, expires_at, created_at, updated_at)
	`

	sqlFindSessionByID = `
		select * 
		from session
		where id = $1
			and expires_at > now()
	`

	sqlGetSessionUser = `
		select user_id
		from session
		where id = $1
			and expires_at > now()
	`

	sqlExtendSession = `
		update session
		set expires_at = $2, updated_at = now()
		where id = $1
			and expires_at > now()
	`

	sqlDeleteSession = `
		delete from session
		where id = $1
	`
)

func (db *DB) CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error) {
	var session models.Session
	err := db.pool.QueryRow(ctx, sqlCreateSession, userID, expiresAt).
		Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (db *DB) FindSessionByID(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := db.pool.QueryRow(ctx, sqlFindSessionByID, sessionID).
		Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (db *DB) GetSessionUser(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	err := db.pool.QueryRow(ctx, sqlGetSessionUser, sessionID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, apperrors.ErrSessionNotFound
	}
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func (db *DB) ExtendSession(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	tag, err := db.pool.Exec(ctx, sqlExtendSession, sessionID, expiresAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrSessionNotFound
	}
	return nil
}

func (db *DB) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := db.pool.Exec(ctx, sqlDeleteSession, sessionID)
	return err
}
