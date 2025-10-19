package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDateRange parses a date range in the format "dd/mm/yy-dd/mm/yy" or single date "dd/mm/yy"
// Also accepts "dd/mm-dd/mm" or single date "dd/mm" (uses current year)
// Returns startDate and endDate in YYYY-MM-DD format
func ParseDateRange(dateRange string) (string, string, error) {
	parts := strings.Split(dateRange, "-")
	if len(parts) == 0 || len(parts) > 2 {
		return "", "", fmt.Errorf("invalid date format. Expected 'dd/mm' or 'dd/mm/yy' or 'dd/mm-dd/mm' or 'dd/mm/yy-dd/mm/yy'")
	}

	// Parse start date
	startDate, err := parseDate(parts[0])
	if err != nil {
		return "", "", fmt.Errorf("invalid start date: %w", err)
	}

	// If only one date provided, use it as both start and end
	if len(parts) == 1 {
		return startDate, startDate, nil
	}

	// Parse end date
	endDate, err := parseDate(parts[1])
	if err != nil {
		return "", "", fmt.Errorf("invalid end date: %w", err)
	}

	return startDate, endDate, nil
}

// parseDate converts dd/mm/yy or dd/mm to YYYY-MM-DD format
// If year is not provided, uses current year
func parseDate(dateStr string) (string, error) {
	dateStr = strings.TrimSpace(dateStr)
	parts := strings.Split(dateStr, "/")

	if len(parts) != 2 && len(parts) != 3 {
		return "", fmt.Errorf("invalid date format. Expected dd/mm or dd/mm/yy")
	}

	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid day: %s", parts[0])
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid month: %s", parts[1])
	}

	var year int
	if len(parts) == 3 {
		// Year provided
		year, err = strconv.Atoi(parts[2])
		if err != nil {
			return "", fmt.Errorf("invalid year: %s", parts[2])
		}

		// Convert 2-digit year to 4-digit year
		if year < 50 {
			year += 2000
		} else if year < 100 {
			year += 1900
		}
	} else {
		// No year provided, use current year
		year = time.Now().Year()
	}

	// Validate date
	if day < 1 || day > 31 || month < 1 || month > 12 {
		return "", fmt.Errorf("invalid date: %d/%d/%d", day, month, year)
	}

	// Format as YYYY-MM-DD
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day), nil
}

// FormatDateForDisplay converts YYYY-MM-DD to dd/mm/yy for display
func FormatDateForDisplay(dateStr string) (string, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %s", dateStr)
	}

	year := t.Year() % 100 // Get last 2 digits
	return fmt.Sprintf("%02d/%02d/%02d", t.Day(), t.Month(), year), nil
}
