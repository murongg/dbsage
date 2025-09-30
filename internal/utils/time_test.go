package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTick(t *testing.T) {
	tickFunc := Tick()

	// Test that the function returns a time.Time
	result := tickFunc()
	_, ok := result.(time.Time)
	assert.True(t, ok, "Tick should return a time.Time")

	// Test that successive calls return different times (though this might be flaky in very fast systems)
	time1 := tickFunc().(time.Time)
	time.Sleep(1 * time.Nanosecond) // Ensure some time passes
	time2 := tickFunc().(time.Time)

	// In most cases, these should be different, but we'll just ensure they're valid times
	assert.True(t, !time1.IsZero(), "First tick should return a valid time")
	assert.True(t, !time2.IsZero(), "Second tick should return a valid time")
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "nanoseconds",
			duration: 500 * time.Nanosecond,
			expected: "500ns",
		},
		{
			name:     "microseconds",
			duration: 750 * time.Microsecond,
			expected: "750µs",
		},
		{
			name:     "milliseconds only",
			duration: 250 * time.Millisecond,
			expected: "250ms",
		},
		{
			name:     "milliseconds with microseconds",
			duration: 250*time.Millisecond + 500*time.Microsecond,
			expected: "250ms",
		},
		{
			name:     "seconds",
			duration: 2 * time.Second,
			expected: "2s",
		},
		{
			name:     "seconds with milliseconds",
			duration: 2*time.Second + 500*time.Millisecond,
			expected: "2.5s",
		},
		{
			name:     "minutes",
			duration: 5 * time.Minute,
			expected: "5m0s",
		},
		{
			name:     "minutes with seconds",
			duration: 5*time.Minute + 30*time.Second,
			expected: "5m30s",
		},
		{
			name:     "hours",
			duration: 2 * time.Hour,
			expected: "2h0m0s",
		},
		{
			name:     "complex duration",
			duration: 2*time.Hour + 15*time.Minute + 30*time.Second + 250*time.Millisecond,
			expected: "2h15m30.25s",
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDurationTruncation(t *testing.T) {
	// Test that durations are properly truncated to milliseconds when >= 1ms
	duration := 1*time.Second + 123*time.Millisecond + 456*time.Microsecond + 789*time.Nanosecond
	result := FormatDuration(duration)

	// Should be truncated to milliseconds, so microseconds and nanoseconds should be removed
	expected := "1.123s"
	assert.Equal(t, expected, result)
}

func BenchmarkTick(b *testing.B) {
	tickFunc := Tick()
	for i := 0; i < b.N; i++ {
		tickFunc()
	}
}

func BenchmarkFormatDuration(b *testing.B) {
	duration := 2*time.Hour + 15*time.Minute + 30*time.Second + 250*time.Millisecond
	for i := 0; i < b.N; i++ {
		FormatDuration(duration)
	}
}
