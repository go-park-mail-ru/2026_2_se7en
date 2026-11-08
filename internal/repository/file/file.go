package file

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/tx"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/jackc/pgx/v5"
)

type FileRepositoryImpl struct {
	pool storage.DBPool
}

func NewFileRepository(pool storage.DBPool) *FileRepositoryImpl {
	return &FileRepositoryImpl{pool: pool}
}

func (r *FileRepositoryImpl) Create(ctx context.Context, url, filename, mimeType string) (*models.File, error) {
	const query = `insert into file (url, filename, type) values ($1, $2, $3) returning id, url, filename, type, created_at, updated_at`
	var f models.File

	executor := r.getExecutor(ctx)
	err := executor.QueryRow(ctx, query, url, filename, mimeType).Scan(&f.ID, &f.URL, &f.Filename, &f.Type, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FileRepositoryImpl) getExecutor(ctx context.Context) interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
} {
	if transaction, ok := tx.FromContext(ctx); ok {
		return transaction
	}
	return r.pool
}
