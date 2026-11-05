package tx

import (
	"context"

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
	defer func() { _ = tx.Rollback(ctx) }()

	ctx = WithTx(ctx, tx)
	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
