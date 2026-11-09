package tx

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/storage"
)

type ManagerImpl struct {
	pool storage.DBPool
}

func NewManager(pool storage.DBPool) *ManagerImpl {
	return &ManagerImpl{pool: pool}
}

func (m *ManagerImpl) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		_ = tx.Rollback(rollbackCtx)
	}()

	ctx = WithTx(ctx, tx)
	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
