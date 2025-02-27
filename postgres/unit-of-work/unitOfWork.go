package unit_of_work

import (
	"context"
	"fmt"
	"package/postgres/executor"
)

type UnitOfWorkImpl struct {
	Executor *executor.Executor
}

func NewUnitOfWork(
	executor *executor.Executor,
) *UnitOfWorkImpl {
	return &UnitOfWorkImpl{
		Executor: executor,
	}
}

func (u *UnitOfWorkImpl) BeginWithTx(ctx context.Context) (context.Context, error) {
	const op = "postgres.UnitOfWork.BeginWithTx"

	tx, err := u.Executor.Begin(ctx)
	if err != nil {
		return ctx, fmt.Errorf("%s: %w: %v", op, ErrTxStartFailed, err)
	}

	ctx = u.Executor.InjectTx(ctx, tx)

	return ctx, nil
}

func (u *UnitOfWorkImpl) BeginWithBatch(ctx context.Context) (context.Context, error) {
	const op = "postgres.UnitOfWork.BeginWithBatch"

	batch := u.Executor.NewBatch()

	ctx = u.Executor.InjectBatch(ctx, batch)

	return ctx, nil
}

func (u *UnitOfWorkImpl) BeginWithTxAndBatch(ctx context.Context) (context.Context, error) {
	const op = "postgres.UnitOfWork.BeginWithTxAndBatch"

	tx, err := u.Executor.Begin(ctx)
	if err != nil {
		return ctx, fmt.Errorf("%s: %w: %v", op, ErrTxStartFailed, err)
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
		results := tx.SendBatch(ctx, batchExecutor.Batch)
		for i := 0; i < batchExecutor.Batch.Len(); i++ {
			_, err := results.Exec()
			if err != nil {
				return fmt.Errorf("%s: %w: %v", op, ErrExecBatch, err)
			}
		}

		if err := results.Close(); err != nil {
			return fmt.Errorf("%s: %w: %v", op, ErrClosingBatch, err)
		}

	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w: %v", op, ErrCommitTx, err)
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
		return fmt.Errorf("%s: %w: %v", op, ErrRollbackTx, err)
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
