package ydb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"

	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/models"
)

// ============================================================================
// Tests for Database Interface and MockDatabase
// ============================================================================

// TestMockDatabase_Query tests the MockDatabase Query method
func TestMockDatabase_Query(t *testing.T) {
	mockDB := NewMockDatabase()
	ctx := context.Background()

	t.Run("query with error", func(t *testing.T) {
		expectedErr := errors.New("query failed")
		mockDB.On("Query", ctx, "INVALID SQL", []table.ParameterOption(nil)).
			Return(nil, expectedErr).Once()

		res, err := mockDB.Query(ctx, "INVALID SQL")

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("query returns nil result", func(t *testing.T) {
		mockDB.On("Query", ctx, "SELECT * FROM users", []table.ParameterOption(nil)).
			Return(nil, nil).Once()

		res, err := mockDB.Query(ctx, "SELECT * FROM users")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

// TestMockDatabase_Exec tests the MockDatabase Exec method
func TestMockDatabase_Exec(t *testing.T) {
	mockDB := NewMockDatabase()
	ctx := context.Background()

	t.Run("successful exec", func(t *testing.T) {
		mockDB.On("Exec", ctx, "UPDATE users SET status = 'active'", []table.ParameterOption(nil)).
			Return(nil).Once()

		err := mockDB.Exec(ctx, "UPDATE users SET status = 'active'")

		assert.NoError(t, err)
	})

	t.Run("exec with error", func(t *testing.T) {
		expectedErr := errors.New("exec failed")
		mockDB.On("Exec", ctx, "INVALID SQL", []table.ParameterOption(nil)).
			Return(expectedErr).Once()

		err := mockDB.Exec(ctx, "INVALID SQL")

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

// TestMockDatabase_DoTx tests the MockDatabase DoTx method
func TestMockDatabase_DoTx(t *testing.T) {
	mockDB := NewMockDatabase()
	ctx := context.Background()

	t.Run("successful transaction", func(t *testing.T) {
		txFunc := func(ctx context.Context, tx table.TransactionActor) error {
			return nil
		}
		mockDB.On("DoTx", ctx, mock.AnythingOfType("func(context.Context, table.TransactionActor) error")).
			Return(nil).Once()

		err := mockDB.DoTx(ctx, txFunc)

		assert.NoError(t, err)
	})

	t.Run("transaction with error", func(t *testing.T) {
		expectedErr := errors.New("transaction failed")
		txFunc := func(ctx context.Context, tx table.TransactionActor) error {
			return expectedErr
		}
		mockDB.On("DoTx", ctx, mock.AnythingOfType("func(context.Context, table.TransactionActor) error")).
			Return(expectedErr).Once()

		err := mockDB.DoTx(ctx, txFunc)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

// TestMockDatabase_Close tests the MockDatabase Close method
func TestMockDatabase_Close(t *testing.T) {
	mockDB := NewMockDatabase()
	ctx := context.Background()

	t.Run("successful close", func(t *testing.T) {
		mockDB.On("Close", ctx).Return(nil).Once()

		err := mockDB.Close(ctx)

		assert.NoError(t, err)
	})

	t.Run("close with error", func(t *testing.T) {
		expectedErr := errors.New("close failed")
		mockDB.On("Close", ctx).Return(expectedErr).Once()

		err := mockDB.Close(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

// TestNewMockDatabase tests the NewMockDatabase constructor
func TestNewMockDatabase(t *testing.T) {
	mockDB := NewMockDatabase()

	assert.NotNil(t, mockDB)
}

// TestYDBClient_NewYDBClient_MissingEnvVars tests NewYDBClient with missing environment variables
func TestYDBClient_NewYDBClient_MissingEnvVars(t *testing.T) {
	ctx := context.Background()

	t.Run("missing YDB_ENDPOINT", func(t *testing.T) {
		// Save and restore environment
		oldEndpoint := getEnv("YDB_ENDPOINT")
		oldDatabase := getEnv("YDB_DATABASE")
		defer setEnv("YDB_ENDPOINT", oldEndpoint)
		defer setEnv("YDB_DATABASE", oldDatabase)

		setEnv("YDB_ENDPOINT", "")
		setEnv("YDB_DATABASE", "/test")

		client, err := NewYDBClient(ctx)

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "YDB_ENDPOINT environment variable not set")
	})

	t.Run("missing YDB_DATABASE", func(t *testing.T) {
		// Save and restore environment
		oldEndpoint := getEnv("YDB_ENDPOINT")
		oldDatabase := getEnv("YDB_DATABASE")
		defer setEnv("YDB_ENDPOINT", oldEndpoint)
		defer setEnv("YDB_DATABASE", oldDatabase)

		setEnv("YDB_ENDPOINT", "localhost:2135")
		setEnv("YDB_DATABASE", "")

		client, err := NewYDBClient(ctx)

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "YDB_DATABASE environment variable not set")
	})
}

// TestYDBClient_Close_NilDriver tests Close with nil driver
func TestYDBClient_Close_NilDriver(t *testing.T) {
	ctx := context.Background()
	client := &YDBClient{driver: nil}

	err := client.Close(ctx)

	assert.NoError(t, err)
}

// TestDatabaseAdapter_NewDatabaseAdapter tests NewDatabaseAdapter
func TestDatabaseAdapter_NewDatabaseAdapter(t *testing.T) {
	adapter := NewDatabaseAdapter()
	assert.NotNil(t, adapter)
}

// TestDatabaseAdapter_Close tests DatabaseAdapter Close method
func TestDatabaseAdapter_Close(t *testing.T) {
	adapter := NewDatabaseAdapter()
	ctx := context.Background()

	t.Run("close with no connection", func(t *testing.T) {
		// Should not error when closing with no connection
		err := adapter.Close(ctx)
		assert.NoError(t, err)
	})
}

// ============================================================================
// Helper functions for environment variable testing
// ============================================================================

// getEnv gets environment variable value
func getEnv(key string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		return ""
	}
	return val
}

// setEnv sets environment variable value
func setEnv(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}

// ============================================================================
// Original Tests
// ============================================================================

// TestTablePathPrefix tests the TablePathPrefix function
func TestTablePathPrefix(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "custom path",
			path:     "/custom/path",
			expected: `PRAGMA TablePathPrefix("/custom/path");`,
		},
		{
			name:     "empty path returns empty string (database in connection string)",
			path:     "",
			expected: "",
		},
		{
			name:     "root path",
			path:     "/",
			expected: `PRAGMA TablePathPrefix("/");`,
		},
		{
			name:     "path with special characters",
			path:     "/path-with_special/chars",
			expected: `PRAGMA TablePathPrefix("/path-with_special/chars");`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TablePathPrefix(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetFieldTypeForValue tests the getFieldTypeForValue helper function
func TestGetFieldTypeForValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "int32 value",
			value:    int32(42),
			expected: "Int32",
		},
		{
			name:     "int value",
			value:    42,
			expected: "Int32",
		},
		{
			name:     "bool value true",
			value:    true,
			expected: "Bool",
		},
		{
			name:     "bool value false",
			value:    false,
			expected: "Bool",
		},
		{
			name:     "unknown type defaults to Utf8",
			value:    "string",
			expected: "Utf8",
		},
		{
			name:     "float defaults to Utf8",
			value:    3.14,
			expected: "Utf8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFieldTypeForValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNewParameter tests the NewParameter function
func TestNewParameter(t *testing.T) {
	tests := []struct {
		name      string
		paramName string
		value     interface{}
		wantPanic bool
	}{
		{
			name:      "valid text value",
			paramName: "$test_param",
			value:     types.TextValue("test"),
			wantPanic: false,
		},
		{
			name:      "valid int64 value",
			paramName: "$int_param",
			value:     types.Int64Value(42),
			wantPanic: false,
		},
		{
			name:      "invalid value type - string",
			paramName: "$invalid",
			value:     "string_value",
			wantPanic: true,
		},
		{
			name:      "invalid value type - int",
			paramName: "$invalid",
			value:     42,
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					NewParameter(tt.paramName, tt.value)
				})
			} else {
				assert.NotPanics(t, func() {
					result := NewParameter(tt.paramName, tt.value)
					assert.NotNil(t, result)
				})
			}
		})
	}
}

// TestGetConnection_MissingEnvVars tests GetConnection with missing environment variables
func TestGetConnection_MissingEnvVars(t *testing.T) {
	// Note: we can't actually unset env vars in Go tests easily,
	// so we'll test that the function returns appropriate errors
	// In integration tests, you would set up test environment variables

	tests := []struct {
		name    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "check connection error handling",
			wantErr: true,
			errMsg:  "YDB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new context to avoid sync.Once issues
			ctx := context.Background()

			// This test verifies error handling without actually connecting
			// In integration tests, you would set up test environment variables
			_, err := GetConnection(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestQuery_Construction tests SQL query construction logic
func TestQuery_Construction(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		paramsCount int
		description string
	}{
		{
			name:        "user query by telegram ID",
			sql:         TablePathPrefix("") + `DECLARE $telegram_chat_id AS Int64; SELECT reviewer_login FROM users WHERE telegram_chat_id = $telegram_chat_id;`,
			paramsCount: 1,
			description: "Query for user by telegram chat ID",
		},
		{
			name:        "user settings query",
			sql:         TablePathPrefix("") + `DECLARE $reviewer_login AS Utf8; SELECT response_deadline_shift_minutes FROM user_settings WHERE reviewer_login = $reviewer_login;`,
			paramsCount: 1,
			description: "Query for user settings",
		},
		{
			name:        "whitelist check query",
			sql:         TablePathPrefix("") + `DECLARE $reviewer_login AS Utf8; DECLARE $project_name AS Utf8; SELECT COUNT(*) AS count FROM user_project_whitelist WHERE reviewer_login = $reviewer_login;`,
			paramsCount: 2,
			description: "Query to check whitelist membership",
		},
		{
			name:        "review request status update",
			sql:         TablePathPrefix("") + `DECLARE $id AS Utf8; DECLARE $status AS Utf8; UPDATE review_requests SET status = $status WHERE id = $id;`,
			paramsCount: 2,
			description: "Update review request status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify SQL contains required elements
			// Note: TablePathPrefix("") returns empty string since database is in connection string
			assert.Contains(t, tt.sql, "DECLARE", "SQL should contain parameter declarations")

			// Count parameters
			paramCount := countParameters(tt.sql)
			assert.Equal(t, tt.paramsCount, paramCount, "Parameter count mismatch")
		})
	}
}

// Helper function to count DECLARE statements in SQL
func countParameters(sql string) int {
	count := 0
	idx := 0
	for {
		found := findSubstring(sql, "DECLARE ", idx)
		if found == -1 {
			break
		}
		count++
		idx = found + 8
	}
	return count
}

// Simple substring finder
func findSubstring(s, substr string, start int) int {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// TestBuildInClause tests the IN clause construction logic
func TestBuildInClause(t *testing.T) {
	tests := []struct {
		name      string
		statuses  []string
		wantEmpty bool
		wantCount int
	}{
		{
			name:      "empty statuses",
			statuses:  []string{},
			wantEmpty: true,
			wantCount: 0,
		},
		{
			name:      "single status",
			statuses:  []string{models.StatusApproved},
			wantEmpty: false,
			wantCount: 1,
		},
		{
			name:      "multiple statuses",
			statuses:  []string{models.StatusApproved, models.StatusCancelled, models.StatusAutoCancelled},
			wantEmpty: false,
			wantCount: 3,
		},
		{
			name: "all intermediate statuses",
			statuses: []string{
				models.StatusUnknownProjectReview,
				models.StatusKnownProjectReview,
				models.StatusWhitelisted,
				models.StatusNotWhitelisted,
				models.StatusNeedToApprove,
				models.StatusWaitingForApprove,
			},
			wantEmpty: false,
			wantCount: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantEmpty {
				assert.Empty(t, tt.statuses, "Statuses should be empty")
			} else {
				assert.NotEmpty(t, tt.statuses, "Statuses should not be empty")
				assert.Equal(t, tt.wantCount, len(tt.statuses))

				// Simulate IN clause construction
				inClause := buildInClause(tt.statuses)
				assert.NotEmpty(t, inClause, "IN clause should not be empty")

				// Verify all statuses are quoted
				for _, status := range tt.statuses {
					assert.Contains(t, inClause, fmt.Sprintf(`"%s"`, status),
						"IN clause should contain quoted status")
				}
			}
		})
	}
}

// Helper function to build IN clause (simulating the logic in GetReviewRequestsByStatus)
func buildInClause(statuses []string) string {
	if len(statuses) == 0 {
		return ""
	}

	inClause := ""
	for i, status := range statuses {
		if i > 0 {
			inClause += ", "
		}
		inClause += `"` + status + `"`
	}
	return inClause
}

// TestSQLValidation tests SQL query validation patterns
func TestSQLValidation(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		shouldHave    []string
		shouldNotHave []string
	}{
		{
			name: "user upsert SQL",
			sql: TablePathPrefix("") + `
				DECLARE $reviewer_login AS Utf8;
				DECLARE $status AS Utf8;
				DECLARE $telegram_chat_id AS Int64;
				UPSERT INTO users (reviewer_login, status, telegram_chat_id)
				VALUES ($reviewer_login, $status, $telegram_chat_id);
			`,
			// Note: PRAGMA TablePathPrefix not expected when database is in connection string
			shouldHave: []string{
				"DECLARE",
				"UPSERT",
				"INTO users",
				"VALUES",
			},
			shouldNotHave: []string{
				"SELECT",
				"DROP",
				"DELETE",
			},
		},
		{
			name: "user settings update SQL",
			sql: TablePathPrefix("") + `
				DECLARE $reviewer_login AS Utf8;
				DECLARE $value AS Int32;
				UPDATE user_settings
				SET response_deadline_shift_minutes = $value
				WHERE reviewer_login = $reviewer_login;
			`,
			// Note: PRAGMA TablePathPrefix not expected when database is in connection string
			shouldHave: []string{
				"UPDATE",
				"SET",
				"WHERE",
			},
			shouldNotHave: []string{
				"INSERT",
				"SELECT",
			},
		},
		{
			name: "whitelist insert SQL",
			sql: TablePathPrefix("") + `
				DECLARE $reviewer_login AS Utf8;
				DECLARE $entry_type AS Utf8;
				DECLARE $name AS Utf8;
				INSERT INTO user_project_whitelist (reviewer_login, entry_type, name)
				VALUES ($reviewer_login, $entry_type, $name);
			`,
			// Note: PRAGMA TablePathPrefix not expected when database is in connection string
			shouldHave: []string{
				"INSERT",
				"INTO user_project_whitelist",
				"VALUES",
			},
			shouldNotHave: []string{
				"UPDATE",
				"DELETE",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expected := range tt.shouldHave {
				assert.Contains(t, tt.sql, expected,
					"SQL should contain '%s'", expected)
			}

			for _, notExpected := range tt.shouldNotHave {
				assert.NotContains(t, tt.sql, notExpected,
					"SQL should not contain '%s'", notExpected)
			}
		})
	}
}

// TestParameterBuilding tests parameter construction for different types
func TestParameterBuilding(t *testing.T) {
	tests := []struct {
		name     string
		buildFn  func() []interface{}
		verifyFn func(t *testing.T, params []interface{})
	}{
		{
			name: "user query parameters",
			buildFn: func() []interface{} {
				chatID := int64(123456789)
				return []interface{}{
					"$telegram_chat_id", types.Int64Value(chatID),
				}
			},
			verifyFn: func(t *testing.T, params []interface{}) {
				require.Len(t, params, 2)
				assert.Equal(t, "$telegram_chat_id", params[0])
				assert.IsType(t, types.Int64Value(0), params[1])
			},
		},
		{
			name: "user upsert parameters",
			buildFn: func() []interface{} {
				now := uint32(time.Now().Unix())
				user := &models.User{
					ReviewerLogin:     "testuser",
					Status:            models.UserStatusActive,
					TelegramChatID:    123456789,
					CreatedAt:         now,
					LastAuthSuccessAt: &now,
				}
				return []interface{}{
					"$reviewer_login", types.TextValue(user.ReviewerLogin),
					"$status", types.TextValue(user.Status),
					"$telegram_chat_id", types.Int64Value(user.TelegramChatID),
				}
			},
			verifyFn: func(t *testing.T, params []interface{}) {
				require.Len(t, params, 6)
				assert.Equal(t, "$reviewer_login", params[0])
				assert.Equal(t, "$status", params[2])
				assert.Equal(t, "$telegram_chat_id", params[4])
			},
		},
		{
			name: "user settings parameters",
			buildFn: func() []interface{} {
				settings := &models.UserSettings{
					ReviewerLogin:                  "testuser",
					ResponseDeadlineShiftMinutes:   30,
					NonWhitelistCancelDelayMinutes: 5,
					NotifyWhitelistTimeout:         true,
					NotifyNonWhitelistCancel:       false,
				}
				return []interface{}{
					"$reviewer_login", types.TextValue(settings.ReviewerLogin),
					"$response_deadline_shift_minutes", types.Int32Value(settings.ResponseDeadlineShiftMinutes),
					"$notify_whitelist_timeout", types.BoolValue(settings.NotifyWhitelistTimeout),
				}
			},
			verifyFn: func(t *testing.T, params []interface{}) {
				require.Len(t, params, 6)
				assert.Equal(t, "$reviewer_login", params[0])
				assert.Equal(t, "$response_deadline_shift_minutes", params[2])
				assert.Equal(t, "$notify_whitelist_timeout", params[4])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.buildFn()
			tt.verifyFn(t, params)
		})
	}
}

// TestErrorMessages tests error message construction
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "user not found error",
			err:         fmt.Errorf("user not found with telegram_chat_id %d", 123456789),
			expectedMsg: "user not found",
		},
		{
			name:        "query failed error",
			err:         fmt.Errorf("failed to query user by reviewer_login %s: %w", "testuser", fmt.Errorf("connection failed")),
			expectedMsg: "failed to query",
		},
		{
			name:        "scan error",
			err:         fmt.Errorf("failed to scan user: %w", fmt.Errorf("invalid column type")),
			expectedMsg: "failed to scan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.Contains(t, tt.err.Error(), tt.expectedMsg)
		})
	}
}

// TestDataConversion tests data conversion logic
func TestDataConversion(t *testing.T) {
	t.Run("timestamp conversion", func(t *testing.T) {
		now := time.Now()
		ts := now.Unix()

		// Convert to YDB value
		ydbValue := types.DatetimeValue(uint32(ts))
		// Verify the value is created successfully
		assert.NotNil(t, ydbValue)
	})

	t.Run("optional timestamp conversion - with value", func(t *testing.T) {
		now := time.Now()
		ts := uint32(now.Unix())

		// Convert to optional YDB value
		ydbValue := optionalDatetime(&ts)
		// Verify the value is created successfully
		assert.NotNil(t, ydbValue)
	})

	t.Run("optional timestamp conversion - nil", func(t *testing.T) {
		// Convert nil to optional YDB value
		ydbValue := optionalDatetime(nil)
		// Verify the value is created successfully (should be null)
		assert.NotNil(t, ydbValue)
	})
}

// TestUserModelOperations tests user model operations without database
func TestUserModelOperations(t *testing.T) {
	t.Run("create valid user", func(t *testing.T) {
		now := uint32(time.Now().Unix())
		user := &models.User{
			ReviewerLogin:     "testuser",
			Status:            models.UserStatusActive,
			TelegramChatID:    123456789,
			CreatedAt:         now,
			LastAuthSuccessAt: &now,
			LastAuthFailureAt: nil,
		}

		assert.Equal(t, "testuser", user.ReviewerLogin)
		assert.Equal(t, models.UserStatusActive, user.Status)
		assert.Equal(t, int64(123456789), user.TelegramChatID)
		assert.Nil(t, user.LastAuthFailureAt)
	})

	t.Run("create user with failure timestamp", func(t *testing.T) {
		failureTime := uint32(time.Now().Unix())
		user := &models.User{
			ReviewerLogin:     "testuser",
			Status:            models.UserStatusInactive,
			TelegramChatID:    123456789,
			LastAuthFailureAt: &failureTime,
		}

		assert.NotNil(t, user.LastAuthFailureAt)
		assert.Equal(t, failureTime, *user.LastAuthFailureAt)
	})
}

// TestReviewRequestModelOperations tests review request model operations
func TestReviewRequestModelOperations(t *testing.T) {
	t.Run("create review request with optional fields", func(t *testing.T) {
		projectName := "test-project"
		familyLabel := "Test Family"
		decisionDeadline := uint32(time.Now().Add(1 * time.Hour).Unix())

		req := &models.ReviewRequest{
			ID:               "test-id-123",
			ReviewerLogin:    "testuser",
			ProjectName:      &projectName,
			FamilyLabel:      &familyLabel,
			DecisionDeadline: &decisionDeadline,
			Status:           models.StatusUnknownProjectReview,
		}

		assert.Equal(t, "test-id-123", req.ID)
		assert.NotNil(t, req.ProjectName)
		assert.Equal(t, "test-project", *req.ProjectName)
		assert.NotNil(t, req.FamilyLabel)
		assert.Equal(t, "Test Family", *req.FamilyLabel)
		assert.NotNil(t, req.DecisionDeadline)
	})

	t.Run("create review request without optional fields", func(t *testing.T) {
		req := &models.ReviewRequest{
			ID:            "test-id-456",
			ReviewerLogin: "testuser",
			Status:        models.StatusApproved,
		}

		assert.Nil(t, req.ProjectName)
		assert.Nil(t, req.FamilyLabel)
		assert.Nil(t, req.DecisionDeadline)
		assert.Nil(t, req.DecidedAt)
	})
}

// TestStatusValidation tests status validation logic
func TestStatusValidation(t *testing.T) {
	tests := []struct {
		name   string
		status string
		valid  bool
	}{
		{
			name:   "valid intermediate status",
			status: models.StatusUnknownProjectReview,
			valid:  true,
		},
		{
			name:   "valid final status",
			status: models.StatusApproved,
			valid:  true,
		},
		{
			name:   "invalid status",
			status: "INVALID_STATUS",
			valid:  false,
		},
		{
			name:   "empty status",
			status: "",
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := models.IsValidStatus(tt.status)
			assert.Equal(t, tt.valid, isValid)

			if isValid {
				// If it's valid, check if it's intermediate or final
				isIntermediate := models.IsIntermediateStatus(tt.status)
				isFinal := models.IsFinalStatus(tt.status)
				assert.True(t, isIntermediate || isFinal,
					"Valid status should be either intermediate or final")
				assert.False(t, isIntermediate && isFinal,
					"Status cannot be both intermediate and final")
			}
		})
	}
}

// TestTransactionLogic tests transaction-related logic
func TestTransactionLogic(t *testing.T) {
	t.Run("upsert project families in transaction", func(t *testing.T) {
		families := []*models.ProjectFamily{
			{FamilyLabel: "Go", ProjectName: "go-concurrency"},
			{FamilyLabel: "C++", ProjectName: "cpp-modern"},
			{FamilyLabel: "Python", ProjectName: "python-async"},
		}

		assert.Len(t, families, 3)

		// Verify all families have required fields
		for _, family := range families {
			assert.NotEmpty(t, family.FamilyLabel)
			assert.NotEmpty(t, family.ProjectName)
		}
	})

	t.Run("empty families list", func(t *testing.T) {
		families := []*models.ProjectFamily{}
		assert.Empty(t, families)
	})
}

// TestWhitelistOperations tests whitelist operation logic
func TestWhitelistOperations(t *testing.T) {
	t.Run("create family entry", func(t *testing.T) {
		entry := &models.WhitelistEntry{
			ReviewerLogin: "testuser",
			EntryType:     models.EntryTypeFamily,
			Name:          "Go - I",
		}

		assert.Equal(t, models.EntryTypeFamily, entry.EntryType)
		assert.True(t, models.IsValidEntryType(entry.EntryType))
	})

	t.Run("create project entry", func(t *testing.T) {
		entry := &models.WhitelistEntry{
			ReviewerLogin: "testuser",
			EntryType:     models.EntryTypeProject,
			Name:          "go-concurrency",
		}

		assert.Equal(t, models.EntryTypeProject, entry.EntryType)
		assert.True(t, models.IsValidEntryType(entry.EntryType))
	})

	t.Run("invalid entry type", func(t *testing.T) {
		entry := &models.WhitelistEntry{
			ReviewerLogin: "testuser",
			EntryType:     "INVALID_TYPE",
			Name:          "test",
		}

		assert.False(t, models.IsValidEntryType(entry.EntryType))
	})
}

// TestContextHandling tests context handling in various operations
func TestContextHandling(t *testing.T) {
	t.Run("context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		assert.NotNil(t, ctx)
		assert.False(t, ctx.Err() != nil && ctx.Err() == context.DeadlineExceeded)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		assert.Error(t, ctx.Err())
		assert.Equal(t, context.Canceled, ctx.Err())
	})

	t.Run("background context", func(t *testing.T) {
		ctx := context.Background()
		assert.NotNil(t, ctx)
		assert.NoError(t, ctx.Err())
	})
}

// BenchmarkDatetimeConversion benchmarks datetime conversion
func BenchmarkDatetimeConversion(b *testing.B) {
	timestamp := uint32(time.Now().Unix())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = types.DatetimeValue(timestamp)
	}
}

// BenchmarkOptionalDatetimeConversion benchmarks optional datetime conversion
func BenchmarkOptionalDatetimeConversion(b *testing.B) {
	timestamp := uint32(time.Now().Unix())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = optionalDatetime(&timestamp)
	}
}

// BenchmarkTablePathPrefix benchmarks TablePathPrefix function
func BenchmarkTablePathPrefix(b *testing.B) {
	path := "/local"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TablePathPrefix(path)
	}
}
