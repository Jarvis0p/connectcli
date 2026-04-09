package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseHHMMPeriod parses "hh:mm" as hours and minutes into a duration (e.g. "00:10" → 10m, "01:30" → 1h30m).
func ParseHHMMPeriod(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty period")
	}
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid period %q: expected hh:mm", s)
	}
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours in period: %w", err)
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes in period: %w", err)
	}
	if hours < 0 || minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("invalid period: hours must be >= 0, minutes 0-59")
	}
	d := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
	if d < time.Minute {
		return 0, fmt.Errorf("period must be at least 00:01 (1 minute)")
	}
	return d, nil
}

// FormatDurationAsHHMM formats a duration as hh:mm (rounded to nearest minute).
func FormatDurationAsHHMM(d time.Duration) string {
	d = d.Round(time.Minute)
	totalMin := int(d / time.Minute)
	if totalMin < 0 {
		totalMin = 0
	}
	h := totalMin / 60
	m := totalMin % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}
