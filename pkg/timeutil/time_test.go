package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNowUTC(t *testing.T) {
	now := NowUTC()
	assert.Equal(t, time.UTC, now.Location(), "NowUTC() should return UTC time")
}

func TestToUTC(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
	}{
		{
			name:  "UTC time remains UTC",
			input: time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC),
		},
		{
			name:  "Local time converts to UTC",
			input: time.Date(2025, 1, 8, 14, 30, 0, 0, time.Local),
		},
	}

	// Tests requiring location loading
	t.Run("New York time converts to UTC", func(t *testing.T) {
		loc := mustLoadLocation(t, "America/New_York")
		input := time.Date(2025, 1, 8, 14, 30, 0, 0, loc)
		result := ToUTC(input)
		assert.Equal(t, time.UTC, result.Location(), "ToUTC() should return UTC time")
	})

	t.Run("Tokyo time converts to UTC", func(t *testing.T) {
		loc := mustLoadLocation(t, "Asia/Tokyo")
		input := time.Date(2025, 1, 8, 14, 30, 0, 0, loc)
		result := ToUTC(input)
		assert.Equal(t, time.UTC, result.Location(), "ToUTC() should return UTC time")
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToUTC(tt.input)
			assert.Equal(t, time.UTC, result.Location(), "ToUTC() should return UTC time")
		})
	}
}

func TestFormatForMessage(t *testing.T) {
	tests := []struct {
		name      string
		input     time.Time
		expected  string
	}{
		{
			name:     "Standard date time",
			input:    time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC),
			expected: "2025-01-08 14:30:00 UTC",
		},
		{
			name:     "Midnight",
			input:    time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC),
			expected: "2025-01-08 00:00:00 UTC",
		},
		{
			name:     "End of day",
			input:    time.Date(2025, 1, 8, 23, 59, 59, 0, time.UTC),
			expected: "2025-01-08 23:59:59 UTC",
		},
		{
			name:     "With nanoseconds (should be truncated)",
			input:    time.Date(2025, 1, 8, 14, 30, 0, 123456789, time.UTC),
			expected: "2025-01-08 14:30:00 UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatForMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatShort(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Standard date time",
			input:    time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC),
			expected: "Jan 8 14:30 UTC",
		},
		{
			name:     "Different month",
			input:    time.Date(2025, 12, 25, 10, 15, 0, 0, time.UTC),
			expected: "Dec 25 10:15 UTC",
		},
		{
			name:     "Single digit hour",
			input:    time.Date(2025, 6, 5, 9, 5, 0, 0, time.UTC),
			expected: "Jun 5 09:05 UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatShort(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsExpired(t *testing.T) {
	tests := []struct {
		name     string
		deadline time.Time
		expected bool
	}{
		{
			name:     "Past deadline",
			deadline: time.Now().Add(-1 * time.Hour),
			expected: true,
		},
		{
			name:     "Future deadline",
			deadline: time.Now().Add(1 * time.Hour),
			expected: false,
		},
		{
			name:     "Just past (1 second)",
			deadline: time.Now().Add(-1 * time.Second),
			expected: true,
		},
		{
			name:     "Just future (1 second)",
			deadline: time.Now().Add(1 * time.Second),
			expected: false,
		},
		{
			name:     "Slightly in future",
			deadline: time.Now().Add(10 * time.Millisecond),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExpired(tt.deadline)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMinutesUntil(t *testing.T) {
	tests := []struct {
		name           string
		targetTime     time.Time
		minExpected    int
		maxExpected    int
	}{
		{
			name:        "30 minutes in future",
			targetTime:  time.Now().Add(30 * time.Minute),
			minExpected: 29,
			maxExpected: 31,
		},
		{
			name:        "30 minutes in past",
			targetTime:  time.Now().Add(-30 * time.Minute),
			minExpected: -31,
			maxExpected: -29,
		},
		{
			name:        "1 minute in future",
			targetTime:  time.Now().Add(1 * time.Minute),
			minExpected: 0,
			maxExpected: 2,
		},
		{
			name:        "1 hour in future",
			targetTime:  time.Now().Add(1 * time.Hour),
			minExpected: 59,
			maxExpected: 61,
		},
		{
			name:        "Zero duration",
			targetTime:  time.Now(),
			minExpected: -1,
			maxExpected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MinutesUntil(tt.targetTime)
			assert.GreaterOrEqual(t, result, tt.minExpected)
			assert.LessOrEqual(t, result, tt.maxExpected)
		})
	}
}

func TestAddMinutes(t *testing.T) {
	tests := []struct {
		name     string
		base     time.Time
		minutes  int
		expected time.Time
	}{
		{
			name:     "Add positive minutes",
			base:     time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			minutes:  30,
			expected: time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC),
		},
		{
			name:     "Add negative minutes",
			base:     time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			minutes:  -30,
			expected: time.Date(2025, 1, 8, 13, 30, 0, 0, time.UTC),
		},
		{
			name:     "Add zero minutes",
			base:     time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			minutes:  0,
			expected: time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
		},
		{
			name:     "Cross hour boundary",
			base:     time.Date(2025, 1, 8, 14, 45, 0, 0, time.UTC),
			minutes:  30,
			expected: time.Date(2025, 1, 8, 15, 15, 0, 0, time.UTC),
		},
		{
			name:     "Cross day boundary",
			base:     time.Date(2025, 1, 8, 23, 30, 0, 0, time.UTC),
			minutes:  60,
			expected: time.Date(2025, 1, 9, 0, 30, 0, 0, time.UTC),
		},
		{
			name:     "Large positive value",
			base:     time.Date(2025, 1, 8, 12, 0, 0, 0, time.UTC),
			minutes:  1440, // 24 hours
			expected: time.Date(2025, 1, 9, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddMinutes(tt.base, tt.minutes)
			assert.True(t, result.Equal(tt.expected), "AddMinutes() = %v, want %v", result, tt.expected)
		})
	}
}

func TestSubtractMinutes(t *testing.T) {
	tests := []struct {
		name     string
		base     time.Time
		minutes  int
		expected time.Time
	}{
		{
			name:     "Subtract positive minutes",
			base:     time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC),
			minutes:  30,
			expected: time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
		},
		{
			name:     "Subtract zero minutes",
			base:     time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC),
			minutes:  0,
			expected: time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC),
		},
		{
			name:     "Cross hour boundary",
			base:     time.Date(2025, 1, 8, 14, 15, 0, 0, time.UTC),
			minutes:  30,
			expected: time.Date(2025, 1, 8, 13, 45, 0, 0, time.UTC),
		},
		{
			name:     "Cross day boundary",
			base:     time.Date(2025, 1, 9, 0, 30, 0, 0, time.UTC),
			minutes:  60,
			expected: time.Date(2025, 1, 8, 23, 30, 0, 0, time.UTC),
		},
		{
			name:     "Large value",
			base:     time.Date(2025, 1, 9, 12, 0, 0, 0, time.UTC),
			minutes:  1440, // 24 hours
			expected: time.Date(2025, 1, 8, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubtractMinutes(tt.base, tt.minutes)
			assert.True(t, result.Equal(tt.expected), "SubtractMinutes() = %v, want %v", result, tt.expected)
		})
	}
}

func TestDurationInMinutes(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected int
	}{
		{
			name:     "90 minutes",
			duration: 90 * time.Minute,
			expected: 90,
		},
		{
			name:     "1 hour",
			duration: 1 * time.Hour,
			expected: 60,
		},
		{
			name:     "Zero duration",
			duration: 0,
			expected: 0,
		},
		{
			name:     "Fractional minute rounds down",
			duration: 90*time.Minute + 30*time.Second,
			expected: 90,
		},
		{
			name:     "Large duration",
			duration: 24 * time.Hour,
			expected: 1440,
		},
		{
			name:     "Negative duration",
			duration: -30 * time.Minute,
			expected: -30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DurationInMinutes(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToUnixMillis(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		wantPos  bool
	}{
		{
			name:    "2025 date",
			input:   time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC),
			wantPos: true,
		},
		{
			name:    "Unix epoch",
			input:   time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			wantPos: false,
		},
		{
			name:    "Before epoch",
			input:   time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC),
			wantPos: false,
		},
		{
			name:    "Future date",
			input:   time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			wantPos: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToUnixMillis(tt.input)
			if tt.wantPos {
				assert.Greater(t, result, int64(0))
			}
			// Verify round-trip conversion
			roundTrip := FromUnixMillis(result)
			assert.Equal(t, tt.input.UnixMilli(), roundTrip.UnixMilli())
		})
	}
}

func TestToUnixSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		wantPos  bool
	}{
		{
			name:    "2025 date",
			input:   time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC),
			wantPos: true,
		},
		{
			name:    "Unix epoch",
			input:   time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			wantPos: false,
		},
		{
			name:    "Before epoch",
			input:   time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC),
			wantPos: false,
		},
		{
			name:    "Future date",
			input:   time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			wantPos: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToUnixSeconds(tt.input)
			if tt.wantPos {
				assert.Greater(t, result, int64(0))
			}
			// Verify round-trip conversion
			roundTrip := FromUnixSeconds(result)
			assert.Equal(t, tt.input.Unix(), roundTrip.Unix())
		})
	}
}

func TestFromUnixMillis(t *testing.T) {
	tests := []struct {
		name        string
		millis      int64
		expected    time.Time
		expectedLoc *time.Location
	}{
		{
			name:        "Standard timestamp",
			millis:      1736340600000, // 2025-01-08 14:30:00 UTC
			expectedLoc: time.UTC,
		},
		{
			name:        "Unix epoch",
			millis:      0,
			expectedLoc: time.UTC,
		},
		{
			name:        "Negative (before epoch)",
			millis:      -86400000, // 1 day before epoch
			expectedLoc: time.UTC,
		},
		{
			name:        "Large positive value",
			millis:      1735689600000, // 2025-01-01 00:00:00 UTC
			expectedLoc: time.UTC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromUnixMillis(tt.millis)
			assert.Equal(t, tt.expectedLoc, result.Location(), "FromUnixMillis() should return UTC time")
			// Verify round-trip
			backToMillis := ToUnixMillis(result)
			assert.Equal(t, tt.millis, backToMillis)
		})
	}
}

func TestFromUnixSeconds(t *testing.T) {
	tests := []struct {
		name        string
		seconds     int64
		expectedLoc *time.Location
	}{
		{
			name:        "Standard timestamp",
			seconds:     1736340600, // 2025-01-08 14:30:00 UTC
			expectedLoc: time.UTC,
		},
		{
			name:        "Unix epoch",
			seconds:     0,
			expectedLoc: time.UTC,
		},
		{
			name:        "Negative (before epoch)",
			seconds:     -86400, // 1 day before epoch
			expectedLoc: time.UTC,
		},
		{
			name:        "Large positive value",
			seconds:     1735689600, // 2025-01-01 00:00:00 UTC
			expectedLoc: time.UTC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromUnixSeconds(tt.seconds)
			assert.Equal(t, tt.expectedLoc, result.Location(), "FromUnixSeconds() should return UTC time")
			// Verify round-trip
			backToSeconds := ToUnixSeconds(result)
			assert.Equal(t, tt.seconds, backToSeconds)
		})
	}
}

func TestCalculateDecisionDeadline(t *testing.T) {
	tests := []struct {
		name             string
		reviewStartTime  time.Time
		shiftMinutes     int
		expected         time.Time
	}{
		{
			name:            "Standard shift",
			reviewStartTime: time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			shiftMinutes:    20,
			expected:        time.Date(2025, 1, 8, 13, 40, 0, 0, time.UTC),
		},
		{
			name:            "Zero shift",
			reviewStartTime: time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			shiftMinutes:    0,
			expected:        time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
		},
		{
			name:            "Large shift (1 hour)",
			reviewStartTime: time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			shiftMinutes:    60,
			expected:        time.Date(2025, 1, 8, 13, 0, 0, 0, time.UTC),
		},
		{
			name:            "Cross day boundary",
			reviewStartTime: time.Date(2025, 1, 8, 0, 30, 0, 0, time.UTC),
			shiftMinutes:    60,
			expected:        time.Date(2025, 1, 7, 23, 30, 0, 0, time.UTC),
		},
		{
			name:            "Small shift",
			reviewStartTime: time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			shiftMinutes:    5,
			expected:        time.Date(2025, 1, 8, 13, 55, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDecisionDeadline(tt.reviewStartTime, tt.shiftMinutes)
			assert.True(t, result.Equal(tt.expected), "CalculateDecisionDeadline() = %v, want %v", result, tt.expected)
		})
	}
}

func TestCalculateNonWhitelistCancelTime(t *testing.T) {
	tests := []struct {
		name        string
		delayMinutes int
		checkFuture  bool
	}{
		{
			name:        "5 minute delay",
			delayMinutes: 5,
			checkFuture:  true,
		},
		{
			name:        "Zero delay",
			delayMinutes: 0,
			checkFuture:  true,
		},
		{
			name:        "Large delay",
			delayMinutes: 60,
			checkFuture:  true,
		},
		{
			name:        "Small delay",
			delayMinutes: 1,
			checkFuture:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			result := CalculateNonWhitelistCancelTime(tt.delayMinutes)
			after := time.Now()

			// Result should be approximately delayMinutes from now
			// Allow 1 second tolerance for test execution time
			expectedMin := before.Add(time.Duration(tt.delayMinutes) * time.Minute).Add(-1 * time.Second)
			expectedMax := after.Add(time.Duration(tt.delayMinutes) * time.Minute).Add(1 * time.Second)

			assert.GreaterOrEqual(t, result, expectedMin)
			assert.LessOrEqual(t, result, expectedMax)
		})
	}
}

func TestShouldShiftSlot(t *testing.T) {
	tests := []struct {
		name             string
		slotStartTime    time.Time
		thresholdMinutes int
		expected         bool
	}{
		{
			name:             "Slot within threshold (should shift)",
			slotStartTime:    time.Now().Add(20 * time.Minute),
			thresholdMinutes: 25,
			expected:         true,
		},
		{
			name:             "Slot at threshold (should shift)",
			slotStartTime:    time.Now().Add(25 * time.Minute),
			thresholdMinutes: 25,
			expected:         true,
		},
		{
			name:             "Slot just beyond threshold",
			slotStartTime:    time.Now().Add(26 * time.Minute),
			thresholdMinutes: 25,
			expected:         false,
		},
		{
			name:             "Slot far in future",
			slotStartTime:    time.Now().Add(30 * time.Minute),
			thresholdMinutes: 25,
			expected:         false,
		},
		{
			name:             "Slot in past",
			slotStartTime:    time.Now().Add(-10 * time.Minute),
			thresholdMinutes: 25,
			expected:         true,
		},
		{
			name:             "Zero threshold",
			slotStartTime:    time.Now().Add(1 * time.Minute),
			thresholdMinutes: 0,
			expected:         false,
		},
		{
			name:             "Large threshold",
			slotStartTime:    time.Now().Add(2 * time.Hour),
			thresholdMinutes: 180,
			expected:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldShiftSlot(tt.slotStartTime, tt.thresholdMinutes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateSlotDuration(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "90 minute slot",
			start:    time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 8, 15, 30, 0, 0, time.UTC),
			expected: 90,
		},
		{
			name:     "1 hour slot",
			start:    time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 8, 15, 0, 0, 0, time.UTC),
			expected: 60,
		},
		{
			name:     "30 minute slot",
			start:    time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC),
			expected: 30,
		},
		{
			name:     "Cross day boundary",
			start:    time.Date(2025, 1, 8, 23, 30, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 9, 1, 0, 0, 0, time.UTC),
			expected: 90,
		},
		{
			name:     "Zero duration",
			start:    time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			expected: 0,
		},
		{
			name:     "With seconds (fractional minute rounds down)",
			start:    time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 8, 15, 30, 30, 0, time.UTC),
			expected: 90,
		},
		{
			name:     "Negative duration (end before start)",
			start:    time.Date(2025, 1, 8, 15, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC),
			expected: -60,
		},
		{
			name:     "Full day",
			start:    time.Date(2025, 1, 8, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 1, 9, 0, 0, 0, 0, time.UTC),
			expected: 1440,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateSlotDuration(tt.start, tt.end)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to load location, panics on error
func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	require.NoError(t, err)
	return loc
}
