package ydb

import (
	"context"
	"fmt"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result/named"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"

	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/models"
)

// Helper function to convert Unix timestamp to YDB datetime value
func datetimeValueFromUnix(ts int64) types.Value {
	return types.TimestampValueFromTime(time.Unix(ts, 0))
}

// Helper function to create optional datetime value
func optionalDatetimeValue(ts *int64) types.Value {
	if ts == nil {
		return types.NullValue(types.TypeTimestamp)
	}
	return types.OptionalValue(types.TimestampValueFromTime(time.Unix(*ts, 0)))
}

// GetUserByTelegramChatID retrieves a user by their Telegram chat ID
func GetUserByTelegramChatID(ctx context.Context, telegramChatID int64) (*models.User, error) {
	sql := TablePathPrefix("") + `
		DECLARE $telegram_chat_id AS Int64;

		SELECT reviewer_login, status, telegram_chat_id, created_at, last_auth_success_at, last_auth_failure_at
		FROM users
		WHERE telegram_chat_id = $telegram_chat_id;
	`

	params := []table.ParameterOption{
		table.ValueParam("$telegram_chat_id", types.Int64Value(telegramChatID)),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by telegram_chat_id %d: %w", telegramChatID, err)
	}
	defer res.Close()

	var user models.User
	if res.NextRow() {
		err = res.ScanNamed(
			named.Required("reviewer_login", &user.ReviewerLogin),
			named.Required("status", &user.Status),
			named.Required("telegram_chat_id", &user.TelegramChatID),
			named.Required("created_at", &user.CreatedAt),
			named.Required("last_auth_success_at", &user.LastAuthSuccessAt),
			named.Optional("last_auth_failure_at", &user.LastAuthFailureAt),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		return &user, nil
	}

	return nil, fmt.Errorf("user not found with telegram_chat_id %d", telegramChatID)
}

// GetUserByReviewerLogin retrieves a user by their reviewer login
func GetUserByReviewerLogin(ctx context.Context, reviewerLogin string) (*models.User, error) {
	sql := TablePathPrefix("") + `
		DECLARE $reviewer_login AS Utf8;

		SELECT reviewer_login, status, telegram_chat_id, created_at, last_auth_success_at, last_auth_failure_at
		FROM users
		WHERE reviewer_login = $reviewer_login;
	`

	params := []table.ParameterOption{
		table.ValueParam("$reviewer_login", types.TextValue(reviewerLogin)),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by reviewer_login %s: %w", reviewerLogin, err)
	}
	defer res.Close()

	var user models.User
	if res.NextRow() {
		err = res.ScanNamed(
			named.Required("reviewer_login", &user.ReviewerLogin),
			named.Required("status", &user.Status),
			named.Required("telegram_chat_id", &user.TelegramChatID),
			named.Required("created_at", &user.CreatedAt),
			named.Required("last_auth_success_at", &user.LastAuthSuccessAt),
			named.Optional("last_auth_failure_at", &user.LastAuthFailureAt),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		return &user, nil
	}

	return nil, fmt.Errorf("user not found with reviewer_login %s", reviewerLogin)
}

// UpsertUser inserts or updates a user
func UpsertUser(ctx context.Context, user *models.User) error {
	sql := TablePathPrefix("") + `
		DECLARE $reviewer_login AS Utf8;
		DECLARE $status AS Utf8;
		DECLARE $telegram_chat_id AS Int64;
		DECLARE $created_at AS Timestamp;
		DECLARE $last_auth_success_at AS Timestamp;
		DECLARE $last_auth_failure_at AS Optional<Timestamp>;

		UPSERT INTO users (reviewer_login, status, telegram_chat_id, created_at, last_auth_success_at, last_auth_failure_at)
		VALUES ($reviewer_login, $status, $telegram_chat_id, $created_at, $last_auth_success_at, $last_auth_failure_at);
	`

	params := []table.ParameterOption{
		table.ValueParam("$reviewer_login", types.TextValue(user.ReviewerLogin)),
		table.ValueParam("$status", types.TextValue(user.Status)),
		table.ValueParam("$telegram_chat_id", types.Int64Value(user.TelegramChatID)),
		table.ValueParam("$created_at", datetimeValueFromUnix(user.CreatedAt)),
		table.ValueParam("$last_auth_success_at", datetimeValueFromUnix(user.LastAuthSuccessAt)),
		table.ValueParam("$last_auth_failure_at", optionalDatetimeValue(user.LastAuthFailureAt)),
	}

	return Exec(ctx, sql, params...)
}

// UpdateUserStatus updates a user's status
func UpdateUserStatus(ctx context.Context, reviewerLogin, status string) error {
	sql := TablePathPrefix("") + `
		DECLARE $reviewer_login AS Utf8;
		DECLARE $status AS Utf8;

		UPDATE users
		SET status = $status
		WHERE reviewer_login = $reviewer_login;
	`

	params := []table.ParameterOption{
		table.ValueParam("$reviewer_login", types.TextValue(reviewerLogin)),
		table.ValueParam("$status", types.TextValue(status)),
	}

	return Exec(ctx, sql, params...)
}

// GetActiveUsers retrieves all active users
func GetActiveUsers(ctx context.Context) ([]*models.User, error) {
	sql := TablePathPrefix("") + `
		SELECT reviewer_login, status, telegram_chat_id, created_at, last_auth_success_at, last_auth_failure_at
		FROM users
		WHERE status = "ACTIVE";
	`

	res, err := Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users: %w", err)
	}
	defer res.Close()

	var users []*models.User
	for res.NextRow() {
		var user models.User
		err = res.ScanNamed(
			named.Required("reviewer_login", &user.ReviewerLogin),
			named.Required("status", &user.Status),
			named.Required("telegram_chat_id", &user.TelegramChatID),
			named.Required("created_at", &user.CreatedAt),
			named.Required("last_auth_success_at", &user.LastAuthSuccessAt),
			named.Optional("last_auth_failure_at", &user.LastAuthFailureAt),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// GetUserSettings retrieves settings for a user
func GetUserSettings(ctx context.Context, reviewerLogin string) (*models.UserSettings, error) {
	sql := TablePathPrefix("") + `
		DECLARE $reviewer_login AS Utf8;

		SELECT reviewer_login, response_deadline_shift_minutes, non_whitelist_cancel_delay_minutes,
		       notify_whitelist_timeout, notify_non_whitelist_cancel, slot_shift_threshold_minutes,
		       slot_shift_duration_minutes, cleanup_durations_minutes
		FROM user_settings
		WHERE reviewer_login = $reviewer_login;
	`

	params := []table.ParameterOption{
		table.ValueParam("$reviewer_login", types.TextValue(reviewerLogin)),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user settings for %s: %w", reviewerLogin, err)
	}
	defer res.Close()

	var settings models.UserSettings
	if res.NextRow() {
		err = res.ScanNamed(
			named.Required("reviewer_login", &settings.ReviewerLogin),
			named.Required("response_deadline_shift_minutes", &settings.ResponseDeadlineShiftMinutes),
			named.Required("non_whitelist_cancel_delay_minutes", &settings.NonWhitelistCancelDelayMinutes),
			named.Required("notify_whitelist_timeout", &settings.NotifyWhitelistTimeout),
			named.Required("notify_non_whitelist_cancel", &settings.NotifyNonWhitelistCancel),
			named.Required("slot_shift_threshold_minutes", &settings.SlotShiftThresholdMinutes),
			named.Required("slot_shift_duration_minutes", &settings.SlotShiftDurationMinutes),
			named.Required("cleanup_durations_minutes", &settings.CleanupDurationsMinutes),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user settings: %w", err)
		}
		return &settings, nil
	}

	return nil, fmt.Errorf("user settings not found for %s", reviewerLogin)
}

// CreateDefaultUserSettings inserts default settings for a new user
func CreateDefaultUserSettings(ctx context.Context, reviewerLogin string) error {
	settings := models.DefaultUserSettings(reviewerLogin)
	return UpsertUserSettings(ctx, settings)
}

// UpsertUserSettings inserts or updates user settings
func UpsertUserSettings(ctx context.Context, settings *models.UserSettings) error {
	sql := TablePathPrefix("") + `
		DECLARE $reviewer_login AS Utf8;
		DECLARE $response_deadline_shift_minutes AS Int32;
		DECLARE $non_whitelist_cancel_delay_minutes AS Int32;
		DECLARE $notify_whitelist_timeout AS Bool;
		DECLARE $notify_non_whitelist_cancel AS Bool;
		DECLARE $slot_shift_threshold_minutes AS Int32;
		DECLARE $slot_shift_duration_minutes AS Int32;
		DECLARE $cleanup_durations_minutes AS Int32;

		UPSERT INTO user_settings (
			reviewer_login, response_deadline_shift_minutes, non_whitelist_cancel_delay_minutes,
			notify_whitelist_timeout, notify_non_whitelist_cancel, slot_shift_threshold_minutes,
			slot_shift_duration_minutes, cleanup_durations_minutes
		) VALUES (
			$reviewer_login, $response_deadline_shift_minutes, $non_whitelist_cancel_delay_minutes,
			$notify_whitelist_timeout, $notify_non_whitelist_cancel, $slot_shift_threshold_minutes,
			$slot_shift_duration_minutes, $cleanup_durations_minutes
		);
	`

	params := []table.ParameterOption{
		table.ValueParam("$reviewer_login", types.TextValue(settings.ReviewerLogin)),
		table.ValueParam("$response_deadline_shift_minutes", types.Int32Value(settings.ResponseDeadlineShiftMinutes)),
		table.ValueParam("$non_whitelist_cancel_delay_minutes", types.Int32Value(settings.NonWhitelistCancelDelayMinutes)),
		table.ValueParam("$notify_whitelist_timeout", types.BoolValue(settings.NotifyWhitelistTimeout)),
		table.ValueParam("$notify_non_whitelist_cancel", types.BoolValue(settings.NotifyNonWhitelistCancel)),
		table.ValueParam("$slot_shift_threshold_minutes", types.Int32Value(settings.SlotShiftThresholdMinutes)),
		table.ValueParam("$slot_shift_duration_minutes", types.Int32Value(settings.SlotShiftDurationMinutes)),
		table.ValueParam("$cleanup_durations_minutes", types.Int32Value(settings.CleanupDurationsMinutes)),
	}

	return Exec(ctx, sql, params...)
}

// UpdateUserSetting updates a single user setting field
func UpdateUserSetting(ctx context.Context, reviewerLogin, field string, value any) error {
	sql := fmt.Sprintf(TablePathPrefix("")+`
		DECLARE $reviewer_login AS Utf8;
		DECLARE $value AS %s;

		UPDATE user_settings
		SET %s = $value
		WHERE reviewer_login = $reviewer_login;
	`, getFieldTypeForValue(value), field)

	var paramValue table.ParameterOption
	switch v := value.(type) {
	case int32:
		paramValue = table.ValueParam("$value", types.Int32Value(v))
	case int:
		paramValue = table.ValueParam("$value", types.Int32Value(int32(v)))
	case bool:
		paramValue = table.ValueParam("$value", types.BoolValue(v))
	default:
		return fmt.Errorf("unsupported value type for UpdateUserSetting: %T", value)
	}

	params := []table.ParameterOption{
		table.ValueParam("$reviewer_login", types.TextValue(reviewerLogin)),
		paramValue,
	}

	return Exec(ctx, sql, params...)
}

func getFieldTypeForValue(value any) string {
	switch value.(type) {
	case int32, int:
		return "Int32"
	case bool:
		return "Bool"
	default:
		return "Utf8"
	}
}

// GetUserWhitelist retrieves all whitelist entries for a user
func GetUserWhitelist(ctx context.Context, reviewerLogin string) ([]*models.WhitelistEntry, error) {
	sql := TablePathPrefix("") + `
		DECLARE $reviewer_login AS Utf8;

		SELECT reviewer_login, entry_type, name
		FROM user_project_whitelist
		WHERE reviewer_login = $reviewer_login;
	`

	params := []table.ParameterOption{
		table.ValueParam("$reviewer_login", types.TextValue(reviewerLogin)),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query whitelist for %s: %w", reviewerLogin, err)
	}
	defer res.Close()

	var entries []*models.WhitelistEntry
	for res.NextRow() {
		var entry models.WhitelistEntry
		err = res.ScanNamed(
			named.Required("reviewer_login", &entry.ReviewerLogin),
			named.Required("entry_type", &entry.EntryType),
			named.Required("name", &entry.Name),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan whitelist entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

// AddToWhitelist adds an entry to a user's whitelist
func AddToWhitelist(ctx context.Context, entry *models.WhitelistEntry) error {
	sql := TablePathPrefix("") + `
		DECLARE $reviewer_login AS Utf8;
		DECLARE $entry_type AS Utf8;
		DECLARE $name AS Utf8;

		INSERT INTO user_project_whitelist (reviewer_login, entry_type, name)
		VALUES ($reviewer_login, $entry_type, $name);
	`

	params := []table.ParameterOption{
		table.ValueParam("$reviewer_login", types.TextValue(entry.ReviewerLogin)),
		table.ValueParam("$entry_type", types.TextValue(entry.EntryType)),
		table.ValueParam("$name", types.TextValue(entry.Name)),
	}

	return Exec(ctx, sql, params...)
}

// RemoveFromWhitelist removes an entry from a user's whitelist
func RemoveFromWhitelist(ctx context.Context, reviewerLogin, name string) error {
	sql := TablePathPrefix("") + `
		DECLARE $reviewer_login AS Utf8;
		DECLARE $name AS Utf8;

		DELETE FROM user_project_whitelist
		WHERE reviewer_login = $reviewer_login AND name = $name;
	`

	params := []table.ParameterOption{
		table.ValueParam("$reviewer_login", types.TextValue(reviewerLogin)),
		table.ValueParam("$name", types.TextValue(name)),
	}

	return Exec(ctx, sql, params...)
}

// IsInWhitelist checks if a project or family is in a user's whitelist
func IsInWhitelist(ctx context.Context, reviewerLogin, projectName, familyLabel string) (bool, error) {
	sql := TablePathPrefix("") + `
		DECLARE $reviewer_login AS Utf8;
		DECLARE $project_name AS Utf8;
		DECLARE $family_label AS Utf8;

		SELECT COUNT(*) AS count
		FROM user_project_whitelist
		WHERE reviewer_login = $reviewer_login
		  AND (
		    (entry_type = "PROJECT" AND name = $project_name)
		    OR
		    (entry_type = "FAMILY" AND name = $family_label)
		  );
	`

	params := []table.ParameterOption{
		table.ValueParam("$reviewer_login", types.TextValue(reviewerLogin)),
		table.ValueParam("$project_name", types.TextValue(projectName)),
		table.ValueParam("$family_label", types.TextValue(familyLabel)),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return false, fmt.Errorf("failed to check whitelist: %w", err)
	}
	defer res.Close()

	if res.NextRow() {
		var count uint32
		err = res.ScanNamed(named.Required("count", &count))
		if err != nil {
			return false, fmt.Errorf("failed to scan count: %w", err)
		}
		return count > 0, nil
	}

	return false, nil
}

// GetFamilyLabelForProject looks up a project's family label
func GetFamilyLabelForProject(ctx context.Context, projectName string) (string, error) {
	sql := TablePathPrefix("") + `
		DECLARE $project_name AS Utf8;

		SELECT family_label
		FROM project_families
		WHERE project_name = $project_name;
	`

	params := []table.ParameterOption{
		table.ValueParam("$project_name", types.TextValue(projectName)),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return "", fmt.Errorf("failed to query project family: %w", err)
	}
	defer res.Close()

	if res.NextRow() {
		var familyLabel string
		err = res.ScanNamed(named.Required("family_label", &familyLabel))
		if err != nil {
			return "", fmt.Errorf("failed to scan family label: %w", err)
		}
		return familyLabel, nil
	}

	return "", fmt.Errorf("project %s not found in project_families", projectName)
}

// GetAllProjectFamilies retrieves all project families
func GetAllProjectFamilies(ctx context.Context) ([]*models.ProjectFamily, error) {
	sql := TablePathPrefix("") + `
		SELECT family_label, project_name
		FROM project_families;
	`

	res, err := Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to query project families: %w", err)
	}
	defer res.Close()

	var families []*models.ProjectFamily
	for res.NextRow() {
		var family models.ProjectFamily
		err = res.ScanNamed(
			named.Required("family_label", &family.FamilyLabel),
			named.Required("project_name", &family.ProjectName),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project family: %w", err)
		}
		families = append(families, &family)
	}

	return families, nil
}

// GetProjectsByFamily retrieves all projects in a family
func GetProjectsByFamily(ctx context.Context, familyLabel string) ([]string, error) {
	sql := TablePathPrefix("") + `
		DECLARE $family_label AS Utf8;

		SELECT project_name
		FROM project_families
		WHERE family_label = $family_label;
	`

	params := []table.ParameterOption{
		table.ValueParam("$family_label", types.TextValue(familyLabel)),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects by family: %w", err)
	}
	defer res.Close()

	var projects []string
	for res.NextRow() {
		var projectName string
		err = res.ScanNamed(named.Required("project_name", &projectName))
		if err != nil {
			return nil, fmt.Errorf("failed to scan project name: %w", err)
		}
		projects = append(projects, projectName)
	}

	return projects, nil
}

// UpsertProjectFamilies replaces all project families
func UpsertProjectFamilies(ctx context.Context, families []*models.ProjectFamily) error {
	return DoTx(ctx, func(ctx context.Context, tx table.TransactionActor) error {
		// First, delete all existing entries
		_, err := tx.Execute(ctx, TablePathPrefix("")+`DELETE FROM project_families;`, table.NewQueryParameters())
		if err != nil {
			return fmt.Errorf("failed to clear project_families: %w", err)
		}

		// Then insert all families
		for _, family := range families {
			sql := TablePathPrefix("") + `
				DECLARE $family_label AS Utf8;
				DECLARE $project_name AS Utf8;

				INSERT INTO project_families (family_label, project_name)
				VALUES ($family_label, $project_name);
			`

			params := []table.ParameterOption{
				table.ValueParam("$family_label", types.TextValue(family.FamilyLabel)),
				table.ValueParam("$project_name", types.TextValue(family.ProjectName)),
			}

			_, err = tx.Execute(ctx, sql, table.NewQueryParameters(params...))
			if err != nil {
				return fmt.Errorf("failed to insert project family: %w", err)
			}
		}

		return nil
	})
}

// CreateReviewRequest creates a new review request
func CreateReviewRequest(ctx context.Context, req *models.ReviewRequest) error {
	sql := TablePathPrefix("") + `
		DECLARE $id AS Utf8;
		DECLARE $reviewer_login AS Utf8;
		DECLARE $review_start_time AS Timestamp;
		DECLARE $calendar_slot_id AS Utf8;
		DECLARE $status AS Utf8;
		DECLARE $created_at AS Timestamp;

		INSERT INTO review_requests (id, reviewer_login, review_start_time, calendar_slot_id, status, created_at)
		VALUES ($id, $reviewer_login, $review_start_time, $calendar_slot_id, $status, $created_at);
	`

	params := []table.ParameterOption{
		table.ValueParam("$id", types.TextValue(req.ID)),
		table.ValueParam("$reviewer_login", types.TextValue(req.ReviewerLogin)),
		table.ValueParam("$review_start_time", datetimeValueFromUnix(req.ReviewStartTime)),
		table.ValueParam("$calendar_slot_id", types.TextValue(req.CalendarSlotID)),
		table.ValueParam("$status", types.TextValue(req.Status)),
		table.ValueParam("$created_at", datetimeValueFromUnix(req.CreatedAt)),
	}

	return Exec(ctx, sql, params...)
}

// GetReviewRequestByID retrieves a review request by ID
func GetReviewRequestByID(ctx context.Context, id string) (*models.ReviewRequest, error) {
	sql := TablePathPrefix("") + `
		DECLARE $id AS Utf8;

		SELECT id, reviewer_login, notification_id, project_name, family_label, review_start_time,
		       calendar_slot_id, decision_deadline, non_whitelist_cancel_at, telegram_message_id,
		       status, created_at, decided_at
		FROM review_requests
		WHERE id = $id;
	`

	params := []table.ParameterOption{
		table.ValueParam("$id", types.TextValue(id)),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query review request: %w", err)
	}
	defer res.Close()

	if res.NextRow() {
		return scanReviewRequest(res)
	}

	return nil, fmt.Errorf("review request not found: %s", id)
}

// GetReviewRequestByCalendarSlotID retrieves a review request by calendar slot ID
func GetReviewRequestByCalendarSlotID(ctx context.Context, calendarSlotID string) (*models.ReviewRequest, error) {
	sql := TablePathPrefix("") + `
		DECLARE $calendar_slot_id AS Utf8;

		SELECT id, reviewer_login, notification_id, project_name, family_label, review_start_time,
		       calendar_slot_id, decision_deadline, non_whitelist_cancel_at, telegram_message_id,
		       status, created_at, decided_at
		FROM review_requests
		WHERE calendar_slot_id = $calendar_slot_id;
	`

	params := []table.ParameterOption{
		table.ValueParam("$calendar_slot_id", types.TextValue(calendarSlotID)),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query review request by slot ID: %w", err)
	}
	defer res.Close()

	if res.NextRow() {
		return scanReviewRequest(res)
	}

	return nil, fmt.Errorf("review request not found with calendar_slot_id: %s", calendarSlotID)
}

// GetReviewRequestsByStatus retrieves review requests by status
func GetReviewRequestsByStatus(ctx context.Context, statuses []string) ([]*models.ReviewRequest, error) {
	if len(statuses) == 0 {
		return []*models.ReviewRequest{}, nil
	}

	// Build IN clause
	inClause := ""
	for i, status := range statuses {
		if i > 0 {
			inClause += ", "
		}
		inClause += `"` + status + `"`
	}

	sql := TablePathPrefix("") + fmt.Sprintf(`
		SELECT id, reviewer_login, notification_id, project_name, family_label, review_start_time,
		       calendar_slot_id, decision_deadline, non_whitelist_cancel_at, telegram_message_id,
		       status, created_at, decided_at
		FROM review_requests
		WHERE status IN (%s);
	`, inClause)

	res, err := Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to query review requests by status: %w", err)
	}
	defer res.Close()

	var requests []*models.ReviewRequest
	for res.NextRow() {
		req, err := scanReviewRequest(res)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// GetReviewRequestsByUserAndStatus retrieves review requests for a user with specific statuses
func GetReviewRequestsByUserAndStatus(ctx context.Context, reviewerLogin string, statuses []string) ([]*models.ReviewRequest, error) {
	if len(statuses) == 0 {
		return []*models.ReviewRequest{}, nil
	}

	// Build IN clause
	inClause := ""
	for i, status := range statuses {
		if i > 0 {
			inClause += ", "
		}
		inClause += `"` + status + `"`
	}

	sql := TablePathPrefix("") + fmt.Sprintf(`
		DECLARE $reviewer_login AS Utf8;

		SELECT id, reviewer_login, notification_id, project_name, family_label, review_start_time,
		       calendar_slot_id, decision_deadline, non_whitelist_cancel_at, telegram_message_id,
		       status, created_at, decided_at
		FROM review_requests
		WHERE reviewer_login = $reviewer_login AND status IN (%s);
	`, inClause)

	params := []table.ParameterOption{
		table.ValueParam("$reviewer_login", types.TextValue(reviewerLogin)),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query review requests by user and status: %w", err)
	}
	defer res.Close()

	var requests []*models.ReviewRequest
	for res.NextRow() {
		req, err := scanReviewRequest(res)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// GetExpiredWaitingForApprove retrieves reviews that have passed their decision deadline
func GetExpiredWaitingForApprove(ctx context.Context) ([]*models.ReviewRequest, error) {
	sql := TablePathPrefix("") + `
		DECLARE $now AS Timestamp;

		SELECT id, reviewer_login, notification_id, project_name, family_label, review_start_time,
		       calendar_slot_id, decision_deadline, non_whitelist_cancel_at, telegram_message_id,
		       status, created_at, decided_at
		FROM review_requests
		WHERE status = "WAITING_FOR_APPROVE" AND decision_deadline <= $now;
	`

	params := []table.ParameterOption{
		table.ValueParam("$now", datetimeValueFromUnix(time.Now().Unix())),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired reviews: %w", err)
	}
	defer res.Close()

	var requests []*models.ReviewRequest
	for res.NextRow() {
		req, err := scanReviewRequest(res)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// GetExpiredNotWhitelisted retrieves NOT_WHITELISTED reviews that have passed their cancel time
func GetExpiredNotWhitelisted(ctx context.Context) ([]*models.ReviewRequest, error) {
	sql := TablePathPrefix("") + `
		DECLARE $now AS Timestamp;

		SELECT id, reviewer_login, notification_id, project_name, family_label, review_start_time,
		       calendar_slot_id, decision_deadline, non_whitelist_cancel_at, telegram_message_id,
		       status, created_at, decided_at
		FROM review_requests
		WHERE status = "NOT_WHITELISTED" AND non_whitelist_cancel_at <= $now;
	`

	params := []table.ParameterOption{
		table.ValueParam("$now", datetimeValueFromUnix(time.Now().Unix())),
	}

	res, err := Query(ctx, sql, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired non-whitelisted reviews: %w", err)
	}
	defer res.Close()

	var requests []*models.ReviewRequest
	for res.NextRow() {
		req, err := scanReviewRequest(res)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// UpdateReviewRequestStatus updates a review request's status
func UpdateReviewRequestStatus(ctx context.Context, id, status string, decidedAt *int64) error {
	sql := TablePathPrefix("") + `
		DECLARE $id AS Utf8;
		DECLARE $status AS Utf8;
		DECLARE $decided_at AS Optional<Timestamp>;

		UPDATE review_requests
		SET status = $status, decided_at = $decided_at
		WHERE id = $id;
	`

	params := []table.ParameterOption{
		table.ValueParam("$id", types.TextValue(id)),
		table.ValueParam("$status", types.TextValue(status)),
		table.ValueParam("$decided_at", optionalDatetimeValue(decidedAt)),
	}

	return Exec(ctx, sql, params...)
}

// UpdateReviewRequestWithProjectInfo updates a review request with project info
func UpdateReviewRequestWithProjectInfo(ctx context.Context, id, projectName, familyLabel, notificationID string) error {
	sql := TablePathPrefix("") + `
		DECLARE $id AS Utf8;
		DECLARE $project_name AS Utf8;
		DECLARE $family_label AS Utf8;
		DECLARE $notification_id AS Utf8;
		DECLARE $status AS Utf8;

		UPDATE review_requests
		SET project_name = $project_name,
		    family_label = $family_label,
		    notification_id = $notification_id,
		    status = $status
		WHERE id = $id;
	`

	params := []table.ParameterOption{
		table.ValueParam("$id", types.TextValue(id)),
		table.ValueParam("$project_name", types.TextValue(projectName)),
		table.ValueParam("$family_label", types.TextValue(familyLabel)),
		table.ValueParam("$notification_id", types.TextValue(notificationID)),
		table.ValueParam("$status", types.TextValue(models.StatusKnownProjectReview)),
	}

	return Exec(ctx, sql, params...)
}

// UpdateReviewRequestToWaitingForApprove updates a review request to WAITING_FOR_APPROVE
func UpdateReviewRequestToWaitingForApprove(ctx context.Context, id string, decisionDeadline int64, telegramMessageID string) error {
	sql := TablePathPrefix("") + `
		DECLARE $id AS Utf8;
		DECLARE $decision_deadline AS Timestamp;
		DECLARE $telegram_message_id AS Utf8;
		DECLARE $status AS Utf8;

		UPDATE review_requests
		SET decision_deadline = $decision_deadline,
		    telegram_message_id = $telegram_message_id,
		    status = $status
		WHERE id = $id;
	`

	params := []table.ParameterOption{
		table.ValueParam("$id", types.TextValue(id)),
		table.ValueParam("$decision_deadline", datetimeValueFromUnix(decisionDeadline)),
		table.ValueParam("$telegram_message_id", types.TextValue(telegramMessageID)),
		table.ValueParam("$status", types.TextValue(models.StatusWaitingForApprove)),
	}

	return Exec(ctx, sql, params...)
}

// UpdateReviewRequestToNotWhitelisted updates a review request to NOT_WHITELISTED
func UpdateReviewRequestToNotWhitelisted(ctx context.Context, id string, nonWhitelistCancelAt int64) error {
	sql := TablePathPrefix("") + `
		DECLARE $id AS Utf8;
		DECLARE $non_whitelist_cancel_at AS Timestamp;
		DECLARE $status AS Utf8;

		UPDATE review_requests
		SET non_whitelist_cancel_at = $non_whitelist_cancel_at,
		    status = $status
		WHERE id = $id;
	`

	params := []table.ParameterOption{
		table.ValueParam("$id", types.TextValue(id)),
		table.ValueParam("$non_whitelist_cancel_at", datetimeValueFromUnix(nonWhitelistCancelAt)),
		table.ValueParam("$status", types.TextValue(models.StatusNotWhitelisted)),
	}

	return Exec(ctx, sql, params...)
}

// scanReviewRequest scans a review request from a result set
func scanReviewRequest(res result.Result) (*models.ReviewRequest, error) {
	var req models.ReviewRequest
	var notificationID, projectName, familyLabel, telegramMessageID string
	var decisionDeadline, nonWhitelistCancelAt int64
	var decidedAt *int64

	err := res.ScanNamed(
		named.Required("id", &req.ID),
		named.Required("reviewer_login", &req.ReviewerLogin),
		named.Optional("notification_id", &notificationID),
		named.Optional("project_name", &projectName),
		named.Optional("family_label", &familyLabel),
		named.Required("review_start_time", &req.ReviewStartTime),
		named.Required("calendar_slot_id", &req.CalendarSlotID),
		named.Optional("decision_deadline", &decisionDeadline),
		named.Optional("non_whitelist_cancel_at", &nonWhitelistCancelAt),
		named.Optional("telegram_message_id", &telegramMessageID),
		named.Required("status", &req.Status),
		named.Required("created_at", &req.CreatedAt),
		named.Optional("decided_at", &decidedAt),
	)
	if err != nil {
		return nil, err
	}

	// Convert optional string fields
	if notificationID != "" {
		req.NotificationID = &notificationID
	}
	if projectName != "" {
		req.ProjectName = &projectName
	}
	if familyLabel != "" {
		req.FamilyLabel = &familyLabel
	}
	if telegramMessageID != "" {
		req.TelegramMessageID = &telegramMessageID
	}
	if decisionDeadline != 0 {
		req.DecisionDeadline = &decisionDeadline
	}
	if nonWhitelistCancelAt != 0 {
		req.NonWhitelistCancelAt = &nonWhitelistCancelAt
	}
	req.DecidedAt = decidedAt

	return &req, nil
}
