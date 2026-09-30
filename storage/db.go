package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	conn *pgx.Conn
}

func Connect(ctx context.Context, connString string) (*DB, error) {
	if connString == "" {
		return nil, errors.New("DATABASE_URL is empty")
	}

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

func Close(ctx context.Context, db *DB) error {
	if db == nil || db.conn == nil {
		return nil
	}

	return db.conn.Close(ctx)
}
