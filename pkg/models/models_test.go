package models

import (
	"testing"
)

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{"UnknownProjectReview", StatusUnknownProjectReview},
		{"KnownProjectReview", StatusKnownProjectReview},
		{"Whitelisted", StatusWhitelisted},
		{"NotWhitelisted", StatusNotWhitelisted},
		{"NeedToApprove", StatusNeedToApprove},
		{"WaitingForApprove", StatusWaitingForApprove},
		{"Approved", StatusApproved},
		{"Cancelled", StatusCancelled},
		{"AutoCancelled", StatusAutoCancelled},
		{"AutoCancelledNotWhitelisted", StatusAutoCancelledNotWhitelisted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status == "" {
				t.Errorf("Status constant %s is empty", tt.name)
			}
		})
	}
}

func TestEntryTypeConstants(t *testing.T) {
	if EntryTypeFamily != "FAMILY" {
		t.Errorf("EntryTypeFamily = %s, want FAMILY", EntryTypeFamily)
	}
	if EntryTypeProject != "PROJECT" {
		t.Errorf("EntryTypeProject = %s, want PROJECT", EntryTypeProject)
	}
}

func TestUserStatusConstants(t *testing.T) {
	if UserStatusActive != "ACTIVE" {
		t.Errorf("UserStatusActive = %s, want ACTIVE", UserStatusActive)
	}
	if UserStatusInactive != "INACTIVE" {
		t.Errorf("UserStatusInactive = %s, want INACTIVE", UserStatusInactive)
	}
}

func TestIsIntermediateStatus(t *testing.T) {
	intermediateStates := []string{
		StatusUnknownProjectReview,
		StatusKnownProjectReview,
		StatusWhitelisted,
		StatusNotWhitelisted,
		StatusNeedToApprove,
		StatusWaitingForApprove,
	}

	for _, status := range intermediateStates {
		if !IsIntermediateStatus(status) {
			t.Errorf("IsIntermediateStatus(%s) should return true", status)
		}
	}

	finalStates := []string{
		StatusApproved,
		StatusCancelled,
		StatusAutoCancelled,
		StatusAutoCancelledNotWhitelisted,
	}

	for _, status := range finalStates {
		if IsIntermediateStatus(status) {
			t.Errorf("IsIntermediateStatus(%s) should return false", status)
		}
	}
}

func TestIsFinalStatus(t *testing.T) {
	finalStates := []string{
		StatusApproved,
		StatusCancelled,
		StatusAutoCancelled,
		StatusAutoCancelledNotWhitelisted,
	}

	for _, status := range finalStates {
		if !IsFinalStatus(status) {
			t.Errorf("IsFinalStatus(%s) should return true", status)
		}
	}

	intermediateStates := []string{
		StatusUnknownProjectReview,
		StatusKnownProjectReview,
		StatusWhitelisted,
		StatusNotWhitelisted,
		StatusNeedToApprove,
		StatusWaitingForApprove,
	}

	for _, status := range intermediateStates {
		if IsFinalStatus(status) {
			t.Errorf("IsFinalStatus(%s) should return false", status)
		}
	}
}

func TestDefaultUserSettings(t *testing.T) {
	settings := DefaultUserSettings("testuser")

	if settings.ReviewerLogin != "testuser" {
		t.Errorf("ReviewerLogin = %s, want testuser", settings.ReviewerLogin)
	}
	if settings.ResponseDeadlineShiftMinutes != 20 {
		t.Errorf("ResponseDeadlineShiftMinutes = %d, want 20", settings.ResponseDeadlineShiftMinutes)
	}
	if settings.NonWhitelistCancelDelayMinutes != 5 {
		t.Errorf("NonWhitelistCancelDelayMinutes = %d, want 5", settings.NonWhitelistCancelDelayMinutes)
	}
	if !settings.NotifyWhitelistTimeout {
		t.Errorf("NotifyWhitelistTimeout should be true")
	}
	if !settings.NotifyNonWhitelistCancel {
		t.Errorf("NotifyNonWhitelistCancel should be true")
	}
	if settings.SlotShiftThresholdMinutes != 25 {
		t.Errorf("SlotShiftThresholdMinutes = %d, want 25", settings.SlotShiftThresholdMinutes)
	}
	if settings.SlotShiftDurationMinutes != 15 {
		t.Errorf("SlotShiftDurationMinutes = %d, want 15", settings.SlotShiftDurationMinutes)
	}
	if settings.CleanupDurationsMinutes != 15 {
		t.Errorf("CleanupDurationsMinutes = %d, want 15", settings.CleanupDurationsMinutes)
	}
}

func TestIsValidStatus(t *testing.T) {
	validStatuses := []string{
		StatusUnknownProjectReview,
		StatusKnownProjectReview,
		StatusWhitelisted,
		StatusNotWhitelisted,
		StatusNeedToApprove,
		StatusWaitingForApprove,
		StatusApproved,
		StatusCancelled,
		StatusAutoCancelled,
		StatusAutoCancelledNotWhitelisted,
	}

	for _, status := range validStatuses {
		if !IsValidStatus(status) {
			t.Errorf("IsValidStatus(%s) should return true", status)
		}
	}

	invalidStatuses := []string{
		"",
		"INVALID",
		"unknown",
		"PENDING",
	}

	for _, status := range invalidStatuses {
		if IsValidStatus(status) {
			t.Errorf("IsValidStatus(%s) should return false", status)
		}
	}
}

func TestIsValidEntryType(t *testing.T) {
	if !IsValidEntryType(EntryTypeFamily) {
		t.Errorf("IsValidEntryType(FAMILY) should return true")
	}
	if !IsValidEntryType(EntryTypeProject) {
		t.Errorf("IsValidEntryType(PROJECT) should return true")
	}
	if IsValidEntryType("INVALID") {
		t.Errorf("IsValidEntryType(INVALID) should return false")
	}
}

func TestIsValidUserStatus(t *testing.T) {
	if !IsValidUserStatus(UserStatusActive) {
		t.Errorf("IsValidUserStatus(ACTIVE) should return true")
	}
	if !IsValidUserStatus(UserStatusInactive) {
		t.Errorf("IsValidUserStatus(INACTIVE) should return true")
	}
	if IsValidUserStatus("INVALID") {
		t.Errorf("IsValidUserStatus(INVALID) should return false")
	}
}
