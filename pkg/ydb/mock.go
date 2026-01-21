package ydb

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
)

// MockDatabase is a mock implementation of the Database interface
type MockDatabase struct {
	mock.Mock
}

// NewMockDatabase creates a new MockDatabase instance
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{}
}

// Query executes a query and returns the result set
func (m *MockDatabase) Query(ctx context.Context, sql string, params ...table.ParameterOption) (result.Result, error) {
	args := m.Called(ctx, sql, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(result.Result), args.Error(1)
}

// Exec executes a query that doesn't return results
func (m *MockDatabase) Exec(ctx context.Context, sql string, params ...table.ParameterOption) error {
	args := m.Called(ctx, sql, params)
	return args.Error(0)
}

// DoTx executes a function within a transaction
func (m *MockDatabase) DoTx(ctx context.Context, fn func(ctx context.Context, tx table.TransactionActor) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// Close closes the database connection
func (m *MockDatabase) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
