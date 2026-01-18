package ydb

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"

	yc "github.com/ydb-platform/ydb-go-yc-metadata"
)

// GetConnection returns a fresh YDB connection for each call
// This prevents stale session state across function invocations
func GetConnection(ctx context.Context) (*ydb.Driver, error) {
	endpoint := os.Getenv("YDB_ENDPOINT")
	database := os.Getenv("YDB_DATABASE")

	if endpoint == "" {
		return nil, fmt.Errorf("YDB_ENDPOINT environment variable not set")
	}
	if database == "" {
		return nil, fmt.Errorf("YDB_DATABASE environment variable not set")
	}

	connectionString := endpoint + "/?database=" + database

	log.Printf("[YDB] Opening new connection: %s", connectionString)

	db, err := ydb.Open(ctx, connectionString,
		yc.WithCredentials(), // Use instance metadata service for authentication
		yc.WithInternalCA(),  // Append Yandex Cloud certificates
	)

	if err != nil {
		log.Printf("[YDB] Failed to open connection: %v", err)
		return nil, err
	}

	log.Printf("[YDB] Successfully opened connection")
	return db, nil
}

// CloseConnection closes the YDB connection
func CloseConnection(ctx context.Context, db *ydb.Driver) error {
	if db != nil {
		return db.Close(ctx)
	}
	return nil
}

// Query executes a query and returns the result set
func Query(ctx context.Context, sql string, params ...table.ParameterOption) (result.Result, error) {
	driver, err := GetConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get YDB connection: %w", err)
	}
	defer CloseConnection(ctx, driver)

	var res result.Result
	err = driver.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
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
func Exec(ctx context.Context, sql string, params ...table.ParameterOption) error {
	driver, err := GetConnection(ctx)
	if err != nil {
		return fmt.Errorf("failed to get YDB connection: %w", err)
	}
	defer CloseConnection(ctx, driver)

	return driver.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, _, err := s.Execute(ctx, table.DefaultTxControl(), sql, table.NewQueryParameters(params...))
		return err
	}, table.WithIdempotent())
}

// DoTx executes a function within a transaction
func DoTx(ctx context.Context, fn func(ctx context.Context, tx table.TransactionActor) error) error {
	driver, err := GetConnection(ctx)
	if err != nil {
		return fmt.Errorf("failed to get YDB connection: %w", err)
	}
	defer CloseConnection(ctx, driver)

	return driver.Table().DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		return fn(ctx, tx)
	}, table.WithIdempotent())
}

// NewParameter creates a new query parameter
func NewParameter(name string, value any) table.ParameterOption {
	return table.ValueParam(name, value.(types.Value))
}

// TablePathPrefix returns the PRAGMA TablePathPrefix directive
// Returns empty string if path is empty, since the database is already
// set in the connection string
func TablePathPrefix(path string) string {
	if path == "" {
		return "" // No prefix needed when database is set in connection
	}
	return fmt.Sprintf("PRAGMA TablePathPrefix(\"%s\");", path)
}
