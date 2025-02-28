package uow

import (
	"context"
	"fmt"
	"package/postgres/executor"
)

type UnitOfWork interface {
	BeginWithTxAndBatch(ctx context.Context) (context.Context, error)
	BeginWithTx(ctx context.Context) (context.Context, error)
	Rollback(ctx context.Context) error
	GracefulRollback(ctx context.Context, err *error)
	Commit(ctx context.Context) error
}

type UnitOfWorkImpl struct {
	Executor *executor.Manager
}

func NewUnitOfWork(
	executor *executor.Manager,
) *UnitOfWorkImpl {
	return &UnitOfWorkImpl{
		Executor: executor,
	}
}

func (u *UnitOfWorkImpl) BeginWithTx(ctx context.Context) (context.Context, error) {
	const op = "postgres.UnitOfWork.BeginWithTx"

	tx, err := u.Executor.Begin(ctx)
	if err != nil {
		return ctx, fmt.Errorf("%s: %w: %w", op, ErrTxStartFailed, err)
	}

	ctx = u.Executor.InjectTx(ctx, tx)

	return ctx, nil
}

func (u *UnitOfWorkImpl) BeginWithTxAndBatch(ctx context.Context) (context.Context, error) {
	const op = "postgres.UnitOfWork.BeginWithTxAndBatch"

	tx, err := u.Executor.Begin(ctx)
	if err != nil {
		return ctx, fmt.Errorf("%s: %w: %w", op, ErrTxStartFailed, err)
	}

	batch := u.Executor.NewBatch()

	ctx = u.Executor.InjectTx(ctx, tx)
	ctx = u.Executor.InjectBatch(ctx, batch)

	return ctx, nil
}

func (u *UnitOfWorkImpl) Commit(ctx context.Context) error {
	const op = "postgres.UnitOfWork.Commit"

	tx, ok := u.Executor.ExtractTx(ctx)
	if !ok {
		return fmt.Errorf("%s: %w", op, ErrNoCommitTx)
	}

	if batchExecutor, ok := u.Executor.ExtractBatch(ctx); ok {
		if err := func() error {
			results := tx.SendBatch(ctx, batchExecutor.Batch)
			defer results.Close()

			if batchExecutor.Batch.Len() > 0 {
				for i := 0; i < batchExecutor.Batch.Len(); i++ {
					_, err := results.Exec()
					if err != nil {
						return fmt.Errorf("%s: %w: %w", op, ErrExecBatch, err)
					}
				}
			}
			return nil
		}(); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w: %w", op, ErrCommitTx, err)
	}

	return nil
}

func (u *UnitOfWorkImpl) Rollback(ctx context.Context) error {
	const op = "postgres.UnitOfWork.Rollback"

	tx, ok := u.Executor.ExtractTx(ctx)
	if !ok {
		return fmt.Errorf("%s: %w", op, ErrNoRollbackTx)
	}

	if err := tx.Rollback(ctx); err != nil {
		return fmt.Errorf("%s: %w: %w", op, ErrRollbackTx, err)
	}

	return nil
}

func (u *UnitOfWorkImpl) GracefulRollback(ctx context.Context, err *error) {
	const op = "postgres.UnitOfWork.GracefulRollback"

	if r := recover(); r != nil {
		_ = u.Rollback(ctx)
		panic(r)
	}

	if err != nil && *err != nil {
		_ = u.Rollback(ctx)
	}
}
