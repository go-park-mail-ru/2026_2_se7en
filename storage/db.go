package storage

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

type DB struct {
	conn *pgx.Conn
}

func Connect(ctx context.Context, connString string) (*DB, error) {
	if conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL")); err != nil {
		return &DB{conn: conn}, nil
	} else {
		return nil, err
	}
}

func Close(ctx context.Context, db *DB) error {
	return db.conn.Close(ctx)
}
