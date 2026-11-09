package storage

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/apperrors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBPool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type DB struct {
	pool *pgxpool.Pool
}

func (db *DB) Pool() DBPool {
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
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &DB{pool: pool}, nil
}

func Close(db *DB) {
	if db != nil && db.pool != nil {
		db.pool.Close()
	}
}
