package tx

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ManagerImpl struct {
	pool *pgxpool.Pool
}

func NewManager(pool *pgxpool.Pool) *ManagerImpl {
	return &ManagerImpl{pool: pool}
}

func (m *ManagerImpl) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)

		defer cancel()

		_ = tx.Rollback(rollbackCtx)
	}()

	ctx = WithTx(ctx, tx)
	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
