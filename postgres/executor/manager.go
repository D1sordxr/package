package executor

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"package/postgres"
)

// Executor defines the interface for executing database operations.
type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, optionsAndArgs ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, optionsAndArgs ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

// txKey and batchKey are used as keys for storing transactions and batches in the context.
type (
	txKey    struct{}
	batchKey struct{}
)

// Manager of the Executor interface.
// It is used to manage transactions and batches in the context and delegate queries to the appropriate executor.
type Manager struct {
	*postgres.Pool
}

// NewManager creates a new Manager instance with the given Postgres connection pool.
func NewManager(pool *postgres.Pool) *Manager {
	return &Manager{Pool: pool}
}

// InjectTx stores a transaction in the context for later retrieval.
func (e *Manager) InjectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// ExtractTx retrieves a transaction from the context.
// It returns the transaction and a boolean indicating whether a transaction was found.
func (e *Manager) ExtractTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

// NewBatch creates a new BatchExecutor for queueing batch queries.
func (e *Manager) NewBatch() *BatchExecutor {
	return &BatchExecutor{Batch: &pgx.Batch{}}
}

// InjectBatch stores a batch in the context for later retrieval.
func (e *Manager) InjectBatch(ctx context.Context, batch *BatchExecutor) context.Context {
	return context.WithValue(ctx, batchKey{}, batch)
}

// ExtractBatch retrieves a batch from the context, if it exists.
func (e *Manager) ExtractBatch(ctx context.Context) (*BatchExecutor, bool) {
	batch, ok := ctx.Value(batchKey{}).(*BatchExecutor)
	return batch, ok
}

// GetExecutor returns the appropriate executor based on the context.
// If a batch is present in the context, it returns the batch executor.
// If a transaction is present, it returns the transaction.
// Otherwise, it returns a PoolExecutor, which wraps the connection pool.
func (e *Manager) GetExecutor(ctx context.Context) Executor {
	if batch, ok := e.ExtractBatch(ctx); ok {
		return batch
	}

	if tx, ok := e.ExtractTx(ctx); ok {
		return tx
	}

	return &PoolExecutor{Pool: e.Pool}
}
