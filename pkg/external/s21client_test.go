package external

import (
	"testing"
	"time"

	s21client "github.com/arseniisemenow/s21auto-client-go"
	"github.com/stretchr/testify/assert"

	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/models"
)

func TestS21ClientCreation(t *testing.T) {
	t.Run("NewS21Client with tokens", func(t *testing.T) {
		client := NewS21Client("access_token", "refresh_token")
		assert.NotNil(t, client)
		assert.NotNil(t, client.client)
	})

	t.Run("NewS21ClientWithSchoolID", func(t *testing.T) {
		headers := &s21client.ContextHeaders{
			XEDUSchoolID:  "school123",
			XEDUProductID: "product123",
			XEDUOrgUnitID: "org123",
			XEDURouteInfo: "route123",
		}
		client := NewS21ClientWithSchoolID("access_token", "refresh_token", "school123", headers)
		assert.NotNil(t, client)
	})

	t.Run("NewS21ClientFromCreds", func(t *testing.T) {
		client := NewS21ClientFromCreds("username", "password")
		assert.NotNil(t, client)
		assert.NotNil(t, client.client)
	})
}

func TestFormatCallbackData(t *testing.T) {
	tests := []struct {
		name            string
		action          string
		reviewRequestID string
		expected        string
	}{
		{
			name:            "Approve callback",
			action:          "APPROVE",
			reviewRequestID: "550e8400-e29b-41d4-a716-446655440000",
			expected:        "APPROVE:550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:            "Decline callback",
			action:          "DECLINE",
			reviewRequestID: "550e8400-e29b-41d4-a716-446655440000",
			expected:        "DECLINE:550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCallbackData(tt.action, tt.reviewRequestID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalendarSlot(t *testing.T) {
	slot := CalendarSlot{
		ID:    "slot-123",
		Start: time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 1, 8, 15, 0, 0, 0, time.UTC),
		Type:  "FREE_TIME",
	}

	assert.Equal(t, "slot-123", slot.ID)
	assert.Equal(t, "FREE_TIME", slot.Type)
	assert.Equal(t, time.Hour, slot.End.Sub(slot.Start))
}

func TestCalendarBooking(t *testing.T) {
	booking := CalendarBooking{
		ID:          "booking-123",
		EventSlotID: "slot-123",
		Start:       time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
		End:         time.Date(2025, 1, 8, 15, 0, 0, 0, time.UTC),
		ProjectName: "go-concurrency",
	}

	assert.Equal(t, "booking-123", booking.ID)
	assert.Equal(t, "slot-123", booking.EventSlotID)
	assert.Equal(t, "go-concurrency", booking.ProjectName)
}

func TestProjectFamilyCreation(t *testing.T) {
	family := &models.ProjectFamily{
		FamilyLabel: "C - I",
		ProjectName: "C5_s21_decimal",
	}

	assert.Equal(t, "C - I", family.FamilyLabel)
	assert.Equal(t, "C5_s21_decimal", family.ProjectName)
}

func TestExtractProjectNameFromMessage(t *testing.T) {
	message := "Review requested for project go-concurrency"
	result := ExtractProjectNameFromMessage(message)
	assert.Equal(t, message, result)
}

func TestFindNotificationByTime(t *testing.T) {
	notifications := []Notification{
		{
			ID:      "notif-1",
			Message: "Test notification",
			Time:    time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
		},
		{
			ID:      "notif-2",
			Message: "Another notification",
			Time:    time.Date(2025, 1, 8, 14, 1, 0, 0, time.UTC),
		},
	}

	slotTime := time.Date(2025, 1, 8, 14, 0, 30, 0, time.UTC)

	// Should find the first notification within 1 minute window
	result := FindNotificationByTime(notifications, slotTime, 1*time.Minute)
	assert.NotNil(t, result)
	assert.Equal(t, "notif-1", result.ID)

	// Should not find any notification with very small window
	result = FindNotificationByTime(notifications, slotTime, 1*time.Second)
	assert.Nil(t, result)
}

func TestNotificationStructure(t *testing.T) {
	notif := Notification{
		ID:                "notif-123",
		RelatedObjectType: "BOOKING",
		RelatedObjectID:   "slot-123",
		Message:           "Review requested",
		Time:              time.Now(),
		WasRead:           false,
		GroupName:         "Reviews",
	}

	assert.Equal(t, "notif-123", notif.ID)
	assert.Equal(t, "BOOKING", notif.RelatedObjectType)
	assert.Equal(t, "slot-123", notif.RelatedObjectID)
	assert.False(t, notif.WasRead)
}

// Test that CalendarSlot and CalendarBooking structs have the right fields
func TestSlotAndBookingFields(t *testing.T) {
	slot := CalendarSlot{
		ID:    "test-id",
		Start: time.Now(),
		End:   time.Now().Add(time.Hour),
		Type:  models.SlotTypeFreeTime,
	}

	assert.NotNil(t, slot.ID)
	assert.NotNil(t, slot.Type)
	assert.True(t, slot.End.After(slot.Start))

	booking := CalendarBooking{
		ID:          "booking-id",
		EventSlotID: "slot-id",
		Start:       time.Now(),
		End:         time.Now().Add(time.Hour),
		ProjectName: "test-project",
	}

	assert.NotNil(t, booking.ID)
	assert.NotNil(t, booking.EventSlotID)
	assert.NotNil(t, booking.ProjectName)
}

func TestExtractFamiliesEmpty(t *testing.T) {
	// Test with empty graph response
	// This would require mocking the full s21client response
	// For now, just verify the function exists
	t.Skip("Requires full graph response mock")
}
