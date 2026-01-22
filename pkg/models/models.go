package models

// Review request statuses
const (
	StatusUnknownProjectReview        = "UNKNOWN_PROJECT_REVIEW"
	StatusKnownProjectReview          = "KNOWN_PROJECT_REVIEW"
	StatusWhitelisted                 = "WHITELISTED"
	StatusNotWhitelisted              = "NOT_WHITELISTED"
	StatusNeedToApprove               = "NEED_TO_APPROVE"
	StatusWaitingForApprove           = "WAITING_FOR_APPROVE"
	StatusApproved                    = "APPROVED"
	StatusCancelled                   = "CANCELLED"
	StatusAutoCancelled               = "AUTO_CANCELLED"
	StatusAutoCancelledNotWhitelisted = "AUTO_CANCELLED_NOT_WHITELISTED"
)

// Whitelist entry types
const (
	EntryTypeFamily  = "FAMILY"
	EntryTypeProject = "PROJECT"
)

// User status values
const (
	UserStatusActive   = "ACTIVE"
	UserStatusInactive = "INACTIVE"
)

// Slot types from calendar
const (
	SlotTypeFreeTime = "FREE_TIME"
	SlotTypeBooking  = "BOOKING"
)

// Intermediate states are mutable
func IsIntermediateStatus(status string) bool {
	switch status {
	case StatusUnknownProjectReview,
		StatusKnownProjectReview,
		StatusWhitelisted,
		StatusNotWhitelisted,
		StatusNeedToApprove,
		StatusWaitingForApprove:
		return true
	default:
		return false
	}
}

// Final states are immutable
func IsFinalStatus(status string) bool {
	switch status {
	case StatusApproved,
		StatusCancelled,
		StatusAutoCancelled,
		StatusAutoCancelledNotWhitelisted:
		return true
	default:
		return false
	}
}

// User represents a reviewer in the users table
type User struct {
	ReviewerLogin     string  `db:"reviewer_login"`
	Status            string  `db:"status"`
	TelegramChatID    int64   `db:"telegram_chat_id"`
	CreatedAt         uint32  `db:"created_at"`
	LastAuthSuccessAt *uint32 `db:"last_auth_success_at"`
	LastAuthFailureAt *uint32 `db:"last_auth_failure_at"`
}

// UserSettings represents per-user configuration
type UserSettings struct {
	ReviewerLogin                  string `db:"reviewer_login"`
	ResponseDeadlineShiftMinutes   int32  `db:"response_deadline_shift_minutes"`
	NonWhitelistCancelDelayMinutes int32  `db:"non_whitelist_cancel_delay_minutes"`
	NotifyWhitelistTimeout         bool   `db:"notify_whitelist_timeout"`
	NotifyNonWhitelistCancel       bool   `db:"notify_non_whitelist_cancel"`
	SlotShiftThresholdMinutes      int32  `db:"slot_shift_threshold_minutes"`
	SlotShiftDurationMinutes       int32  `db:"slot_shift_duration_minutes"`
	CleanupDurationsMinutes        int32  `db:"cleanup_durations_minutes"`
}

// DefaultUserSettings returns default user settings
func DefaultUserSettings(reviewerLogin string) *UserSettings {
	return &UserSettings{
		ReviewerLogin:                  reviewerLogin,
		ResponseDeadlineShiftMinutes:   20,
		NonWhitelistCancelDelayMinutes: 5,
		NotifyWhitelistTimeout:         true,
		NotifyNonWhitelistCancel:       true,
		SlotShiftThresholdMinutes:      25,
		SlotShiftDurationMinutes:       15,
		CleanupDurationsMinutes:        15,
	}
}

// ProjectFamily represents a project in the project_families table
type ProjectFamily struct {
	FamilyLabel string `db:"family_label"`
	ProjectName string `db:"project_name"`
}

// WhitelistEntry represents an entry in user_project_whitelist
type WhitelistEntry struct {
	ReviewerLogin string `db:"reviewer_login"`
	EntryType     string `db:"entry_type"`
	Name          string `db:"name"`
}

// ReviewRequest represents a review request in the review_requests table
type ReviewRequest struct {
	ID                   string  `db:"id"`
	ReviewerLogin        string  `db:"reviewer_login"`
	NotificationID       *string `db:"notification_id"`
	ProjectName          *string `db:"project_name"`
	FamilyLabel          *string `db:"family_label"`
	ReviewStartTime      uint32  `db:"review_start_time"`
	CalendarSlotID       string  `db:"calendar_slot_id"`
	BookingID            string  `db:"booking_id"`
	DecisionDeadline     *uint32 `db:"decision_deadline"`
	NonWhitelistCancelAt *uint32 `db:"non_whitelist_cancel_at"`
	TelegramMessageID    *string `db:"telegram_message_id"`
	Status               string  `db:"status"`
	CreatedAt            uint32  `db:"created_at"`
	DecidedAt            *uint32 `db:"decided_at"`
}

// CalendarSlot represents a time slot from the calendar API
type CalendarSlot struct {
	ID    string
	Start int64
	End   int64
	Type  string
}

// CalendarBooking represents a booking from the calendar API
type CalendarBooking struct {
	ID          string
	EventSlotID string
	StartTime   int64
	EndTime     int64
	ProjectName string
}

// TelegramCallbackData represents the parsed callback data from Telegram
type TelegramCallbackData struct {
	Action          string
	ReviewRequestID string
}

// LockboxPayload represents the structure of secrets in Lockbox
type LockboxPayload struct {
	Version int                   `json:"version"`
	Users   map[string]UserTokens `json:"users"`
}

// UserTokens represents access and refresh tokens for a user stored in YDB
type UserTokens struct {
	ReviewerLogin string `json:"reviewer_login"`
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	CreatedAt     uint32 `json:"created_at"`
	UpdatedAt     uint32 `json:"updated_at"`
	IssueTime     int64  `db:"issue_time" json:"issue_time"`   // Unix timestamp when token was issued
	ExpiryTime    int64  `db:"expiry_time" json:"expiry_time"` // Unix timestamp when token expires
}

// TokenResponse represents the authentication response from s21 platform
type TokenResponse struct {
	Error            string `json:"error"`
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	IssueTime        int64  `json:"issue_time"`  // Unix timestamp when token was issued
	ExpiryTime       int64  `json:"expiry_time"` // Unix timestamp when token expires
}

// Validation errors
var (
	ErrInvalidStatus     = "invalid review status"
	ErrInvalidEntryType  = "invalid whitelist entry type"
	ErrInvalidUserStatus = "invalid user status"
	ErrInvalidReviewID   = "invalid review request ID"
)

// IsValidStatus checks if a status string is valid
func IsValidStatus(status string) bool {
	switch status {
	case StatusUnknownProjectReview,
		StatusKnownProjectReview,
		StatusWhitelisted,
		StatusNotWhitelisted,
		StatusNeedToApprove,
		StatusWaitingForApprove,
		StatusApproved,
		StatusCancelled,
		StatusAutoCancelled,
		StatusAutoCancelledNotWhitelisted:
		return true
	default:
		return false
	}
}

// IsValidEntryType checks if an entry type is valid
func IsValidEntryType(entryType string) bool {
	return entryType == EntryTypeFamily || entryType == EntryTypeProject
}

// IsValidUserStatus checks if a user status is valid
func IsValidUserStatus(status string) bool {
	return status == UserStatusActive || status == UserStatusInactive
}
