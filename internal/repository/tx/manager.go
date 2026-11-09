package tx

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
	"github.com/go-park-mail-ru/2026_2_se7en/internal/tools/logger"
	"github.com/jackc/pgx/v5"
)

type ManagerImpl struct {
	log *slog.Logger

	pool storage.DBPool
}

func NewManager(pool storage.DBPool, baseLog *slog.Logger) *ManagerImpl {
	return &ManagerImpl{
		log:  logger.Scope(baseLog, logger.LayerRepository, logger.ModuleTransaction),
		pool: pool}
}

func (m *ManagerImpl) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	m.log.DebugContext(ctx, "database operation", logger.KeyOperation, "WithinTransaction")

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		m.log.ErrorContext(ctx, "WithinTransaction failed", logger.Error(err))

		return err
	}

	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)

		defer cancel()

		if err := tx.Rollback(rollbackCtx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			m.log.ErrorContext(rollbackCtx, "Rollback failed", logger.Error(err))
		}
	}()

	ctx = WithTx(ctx, tx)
	if err := fn(ctx); err != nil {
		m.log.ErrorContext(ctx, "WithinTransaction failed", logger.Error(err))

		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		m.log.ErrorContext(ctx, "Commit failed", logger.Error(err))
	}
	return err
}
