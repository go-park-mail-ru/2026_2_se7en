package storage

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDatabaseURLMissing = errors.New("DATABASE_URL is empty")

type DBPool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type DB struct {
	log *slog.Logger

	pool *pgxpool.Pool
}

func (db *DB) Pool() DBPool {
	return db.pool
}

func Connect(ctx context.Context, connString string, baseLog *slog.Logger) (*DB, error) {
	log := logger.Scope(baseLog, logger.LayerStorage, logger.ModuleDatabase)

	if connString == "" {
		return nil, ErrDatabaseURLMissing
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.ErrorContext(ctx, "Connect failed", logger.Error(err))
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		log.ErrorContext(ctx, "Ping failed", logger.Error(err))
		pool.Close()
		return nil, err
	}

	return &DB{
		log:  logger.Scope(baseLog, logger.LayerStorage, logger.ModuleDatabase),
		pool: pool}, nil
}

func Close(db *DB) {
	if db != nil && db.pool != nil {
		db.pool.Close()
		db.log.Info("database pool closed")
	}
}
