package ydb

import (
	"context"
	"fmt"
	"os"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"

	yc "github.com/ydb-platform/ydb-go-yc-metadata"
)

// YDBClient implements the Database interface and wraps a YDB driver
type YDBClient struct {
	driver *ydb.Driver
}

// NewYDBClient creates a new YDB client from environment variables
// Requires YDB_ENDPOINT and YDB_DATABASE environment variables to be set
func NewYDBClient(ctx context.Context) (*YDBClient, error) {
	endpoint := os.Getenv("YDB_ENDPOINT")
	database := os.Getenv("YDB_DATABASE")

	if endpoint == "" {
		return nil, fmt.Errorf("YDB_ENDPOINT environment variable not set")
	}
	if database == "" {
		return nil, fmt.Errorf("YDB_DATABASE environment variable not set")
	}

	driver, err := ydb.Open(ctx, endpoint+"/?database="+database,
		yc.WithCredentials(), // Use instance metadata service for authentication
		yc.WithInternalCA(),  // Append Yandex Cloud certificates
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open YDB connection: %w", err)
	}

	return &YDBClient{driver: driver}, nil
}

// Query executes a query and returns the result set
func (c *YDBClient) Query(ctx context.Context, sql string, params ...table.ParameterOption) (result.Result, error) {
	var res result.Result
	err := c.driver.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, r, err := s.Execute(ctx, table.DefaultTxControl(), sql, table.NewQueryParameters(params...))
		if err != nil {
			return err
		}
		res = r
		return nil
	}, table.WithIdempotent())

	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return res, nil
}

// Exec executes a query that doesn't return results
func (c *YDBClient) Exec(ctx context.Context, sql string, params ...table.ParameterOption) error {
	err := c.driver.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, _, err := s.Execute(ctx, table.DefaultTxControl(), sql, table.NewQueryParameters(params...))
		return err
	}, table.WithIdempotent())

	if err != nil {
		return fmt.Errorf("exec execution failed: %w", err)
	}

	return nil
}

// DoTx executes a function within a transaction
func (c *YDBClient) DoTx(ctx context.Context, fn func(ctx context.Context, tx table.TransactionActor) error) error {
	err := c.driver.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		return fn(ctx, tx)
	}, table.WithIdempotent())

	if err != nil {
		return fmt.Errorf("transaction execution failed: %w", err)
	}

	return nil
}

// Close closes the database connection
func (c *YDBClient) Close(ctx context.Context) error {
	if c.driver != nil {
		return c.driver.Close(ctx)
	}
	return nil
}
