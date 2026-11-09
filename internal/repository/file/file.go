package file

import (
	"context"
	"log/slog"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/repository/tx"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/jackc/pgx/v5"
)

type FileRepositoryImpl struct {
	log *slog.Logger

	pool storage.DBPool
}

func NewFileRepository(pool storage.DBPool, baseLog *slog.Logger) *FileRepositoryImpl {
	return &FileRepositoryImpl{
		log:  logger.Scope(baseLog, logger.LayerRepository, logger.ModuleFile),
		pool: pool}
}

func (r *FileRepositoryImpl) Create(ctx context.Context, url, filename, mimeType string) (*models.File, error) {
	r.log.DebugContext(ctx, "database operation", logger.KeyOperation, "Create")

	const query = `insert into file (url, filename, type) values ($1, $2, $3) returning id, url, filename, type, created_at, updated_at`
	var f models.File

	executor := r.getExecutor(ctx)
	err := executor.QueryRow(ctx, query, url, filename, mimeType).Scan(&f.ID, &f.URL, &f.Filename, &f.Type, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		r.log.ErrorContext(ctx, "Create failed", logger.Error(err))

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
