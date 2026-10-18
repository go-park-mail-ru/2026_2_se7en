package storage

import (
	"context"

	"app/apperrors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

func Connect(ctx context.Context, connString string) (*DB, error) {
	if connString == "" {
		return nil, apperrors.ErrDatabaseURLMissing
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	return &DB{pool: pool}, nil
}

func Close(db *DB) {
	if db != nil && db.pool != nil {
		db.pool.Close()
	}
}
