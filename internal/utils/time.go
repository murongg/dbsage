package utils

import (
	"time"
)

// Tick generates a tick command for animation updates
func Tick() func() interface{} {
	return func() interface{} {
		return time.Now()
	}
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return d.String()
	}
	if d < time.Second {
		return d.Truncate(time.Millisecond).String()
	}
	return d.Truncate(time.Millisecond).String()
}
