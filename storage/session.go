package storage

import (
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

	sqlDeleteSession = `
		delete from session
		where id = $1
	`
)

func (db *DB) CreateSession(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*models.Session, error) {
	var session models.Session
	err := db.conn.QueryRow(ctx, sqlCreateSession, userID, expiresAt).
		Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (db *DB) FindSessionByID(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := db.conn.QueryRow(ctx, sqlFindSessionByID, sessionID).
		Scan(&session.ID, &session.UserID, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (db *DB) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := db.conn.Exec(ctx, sqlDeleteSession, sessionID)
	return err
}
