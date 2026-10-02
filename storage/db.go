package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	conn *pgxpool.Pool
}

func Connect(ctx context.Context, connString string) (*DB, error) {
	if connString == "" {
		return nil, errors.New("DATABASE_URL is empty")
	}

	conn, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

func Close(db *DB) {
	if db != nil || db.conn != nil {
		db.conn.Close()
	}
}
