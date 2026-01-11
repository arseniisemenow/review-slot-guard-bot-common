package ydb

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
)

var (
	initOnce sync.Once
)

// InitSchema creates all tables if they don't exist
// Should be called once at application startup
func InitSchema(ctx context.Context) error {
	var initErr error
	initOnce.Do(func() {
		initErr = createTables(ctx)
	})
	return initErr
}

// createTables creates all required YDB tables
func createTables(ctx context.Context) error {
	database := os.Getenv("YDB_DATABASE")
	if database == "" {
		return fmt.Errorf("YDB_DATABASE environment variable not set")
	}

	// Ensure database path starts with /
	if !strings.HasPrefix(database, "/") {
		database = "/" + database
	}

	logger := log.New(os.Stdout, "[YDB_SCHEMA] ", log.LstdFlags)

	// Create tables
	tables := []struct {
		name    string
		schema  string
	}{
		{
			name: "users",
			schema: `
				CREATE TABLE users (
					reviewer_login Utf8,
					status Utf8,
					telegram_chat_id Int64,
					created_at Datetime,
					last_auth_success_at Datetime,
					last_auth_failure_at Datetime,
					PRIMARY KEY (reviewer_login)
				)
			`,
		},
		{
			name: "user_settings",
			schema: `
				CREATE TABLE user_settings (
					reviewer_login Utf8,
					response_deadline_shift_minutes Int32,
					non_whitelist_cancel_delay_minutes Int32,
					notify_whitelist_timeout Bool,
					notify_non_whitelist_cancel Bool,
					slot_shift_threshold_minutes Int32,
					slot_shift_duration_minutes Int32,
					cleanup_durations_minutes Int32,
					PRIMARY KEY (reviewer_login)
				)
			`,
		},
		{
			name: "user_project_whitelist",
			schema: `
				CREATE TABLE user_project_whitelist (
					reviewer_login Utf8,
					entry_type Utf8,
					name Utf8,
					PRIMARY KEY (reviewer_login, entry_type, name)
				)
			`,
		},
		{
			name: "project_families",
			schema: `
				CREATE TABLE project_families (
					family_label Utf8,
					project_name Utf8,
					PRIMARY KEY (family_label, project_name)
				)
			`,
		},
		{
			name: "review_requests",
			schema: `
				CREATE TABLE review_requests (
					id Utf8,
					reviewer_login Utf8,
					notification_id Utf8,
					project_name Utf8,
					family_label Utf8,
					review_start_time Datetime,
					calendar_slot_id Utf8,
					decision_deadline Datetime,
					non_whitelist_cancel_at Datetime,
					telegram_message_id Utf8,
					status Utf8,
					created_at Datetime,
					decided_at Datetime,
					PRIMARY KEY (id)
				)
			`,
		},
	}

	driver, err := GetConnection(ctx)
	if err != nil {
		return fmt.Errorf("failed to get YDB connection: %w", err)
	}

	for _, tbl := range tables {
		tablePath := database + "/" + tbl.name
		if err := createTableIfNotExists(ctx, driver, tablePath, tbl.schema, database, logger); err != nil {
			return fmt.Errorf("failed to create table %s: %w", tbl.name, err)
		}
	}

	logger.Println("All tables initialized successfully")
	return nil
}

// createTableIfNotExists creates a table if it doesn't already exist
func createTableIfNotExists(ctx context.Context, driver *ydb.Driver, tablePath, schema, database string, logger *log.Logger) error {
	// First, check if table exists by trying to describe it
	err := driver.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, err := s.DescribeTable(ctx, tablePath)
		return err
	})

	if err == nil {
		// Table exists
		logger.Printf("Table already exists: %s", tablePath)
		return nil
	}

	// Table doesn't exist, create it
	logger.Printf("Creating table: %s", tablePath)

	// Build the query with PRAGMA TablePathPrefix
	query := fmt.Sprintf("PRAGMA TablePathPrefix(\"%s\");\n%s", database, schema)

	err = driver.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		return s.ExecuteSchemeQuery(ctx, query)
	})

	if err != nil {
		return fmt.Errorf("failed to execute schema query: %w", err)
	}

	logger.Printf("Table created successfully: %s", tablePath)
	return nil
}
