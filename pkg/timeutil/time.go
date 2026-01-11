package timeutil

import (
	"time"
)

// NowUTC returns current time in UTC
func NowUTC() time.Time {
	return time.Now().UTC()
}

// ToUTC converts any time to UTC
func ToUTC(t time.Time) time.Time {
	return t.UTC()
}

// FormatForMessage formats time for Telegram messages
func FormatForMessage(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05 UTC")
}

// FormatShort formats time in short readable format
func FormatShort(t time.Time) string {
	return t.UTC().Format("Jan 2 15:04 UTC")
}

// IsExpired checks if a deadline has passed
func IsExpired(deadline time.Time) bool {
	return time.Now().After(deadline)
}

// MinutesUntil returns minutes until a time (negative if past)
func MinutesUntil(t time.Time) int {
	duration := time.Until(t)
	return int(duration.Minutes())
}

// AddMinutes adds minutes to a time
func AddMinutes(t time.Time, minutes int) time.Time {
	return t.Add(time.Duration(minutes) * time.Minute)
}

// SubtractMinutes subtracts minutes from a time
func SubtractMinutes(t time.Time, minutes int) time.Time {
	return t.Add(-time.Duration(minutes) * time.Minute)
}

// DurationInMinutes returns duration in minutes
func DurationInMinutes(d time.Duration) int {
	return int(d.Minutes())
}

// ToUnixMillis converts time to Unix milliseconds
func ToUnixMillis(t time.Time) int64 {
	return t.UnixMilli()
}

// ToUnixSeconds converts time to Unix seconds
func ToUnixSeconds(t time.Time) int64 {
	return t.Unix()
}

// FromUnixMillis converts Unix milliseconds to time
func FromUnixMillis(ms int64) time.Time {
	return time.UnixMilli(ms).UTC()
}

// FromUnixSeconds converts Unix seconds to time
func FromUnixSeconds(s int64) time.Time {
	return time.Unix(s, 0).UTC()
}

// CalculateDecisionDeadline calculates when to ask user for decision
func CalculateDecisionDeadline(reviewStartTime time.Time, shiftMinutes int) time.Time {
	return reviewStartTime.Add(-time.Duration(shiftMinutes) * time.Minute)
}

// CalculateNonWhitelistCancelTime calculates when to auto-cancel non-whitelisted review
func CalculateNonWhitelistCancelTime(delayMinutes int) time.Time {
	return time.Now().Add(time.Duration(delayMinutes) * time.Minute)
}

// ShouldShiftSlot checks if slot should be shifted
func ShouldShiftSlot(slotStartTime time.Time, thresholdMinutes int) bool {
	thresholdFromNow := time.Now().Add(time.Duration(thresholdMinutes) * time.Minute)
	return thresholdFromNow.After(slotStartTime) || thresholdFromNow.Equal(slotStartTime)
}

// CalculateSlotDuration returns slot duration in minutes
func CalculateSlotDuration(start, end time.Time) int {
	return int(end.Sub(start).Minutes())
}
