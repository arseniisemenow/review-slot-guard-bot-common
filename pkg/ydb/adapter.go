package ydb

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
)

// DatabaseAdapter implements the Database interface using global functions
type DatabaseAdapter struct{}

// NewDatabaseAdapter creates a new DatabaseAdapter
func NewDatabaseAdapter() *DatabaseAdapter {
	return &DatabaseAdapter{}
}

// Query executes a query and returns the result set
func (d *DatabaseAdapter) Query(ctx context.Context, sql string, params ...table.ParameterOption) (result.Result, error) {
	return Query(ctx, sql, params...)
}

// Exec executes a query that doesn't return results
func (d *DatabaseAdapter) Exec(ctx context.Context, sql string, params ...table.ParameterOption) error {
	return Exec(ctx, sql, params...)
}

// DoTx executes a function within a transaction
func (d *DatabaseAdapter) DoTx(ctx context.Context, fn func(ctx context.Context, tx table.TransactionActor) error) error {
	return DoTx(ctx, fn)
}

// Close closes the database connection
func (d *DatabaseAdapter) Close(ctx context.Context) error {
	return CloseConnection(ctx)
}
