package ydb

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
)

// Database defines the interface for YDB database operations
type Database interface {
	// Query executes a query and returns the result set
	Query(ctx context.Context, sql string, params ...table.ParameterOption) (result.Result, error)

	// Exec executes a query that doesn't return results
	Exec(ctx context.Context, sql string, params ...table.ParameterOption) error

	// DoTx executes a function within a transaction
	DoTx(ctx context.Context, fn func(ctx context.Context, tx table.TransactionActor) error) error

	// Close closes the database connection
	Close(ctx context.Context) error
}
