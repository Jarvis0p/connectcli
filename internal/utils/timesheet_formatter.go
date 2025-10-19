package utils

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// TimesheetShift represents a single shift entry
type TimesheetShift struct {
	Date          string
	Type          string
	StartTime     time.Time
	EndTime       time.Time
	TotalHours    float64
	EmployeeNotes string
	ActualDate    time.Time // For proper sorting
}

// TimesheetDayEntry represents a day's timesheet entry
type TimesheetDayEntry struct {
	Date   string `json:"date"`
	Shifts []struct {
		PunchIn struct {
			TimestampWithTimezone struct {
				Timestamp int64 `json:"timestamp"`
			} `json:"timestampWithTimezone"`
		} `json:"punchIn"`
		PunchOut struct {
			TimestampWithTimezone struct {
				Timestamp int64 `json:"timestamp"`
			} `json:"timestampWithTimezone"`
		} `json:"punchOut"`
		PunchTag struct {
			Name string `json:"name"`
		} `json:"punchTag"`
		EmployeeNotes string `json:"employeeNotes"`
	} `json:"shifts"`
}

// TimesheetEntry represents a timesheet entry
type TimesheetEntry struct {
	TimeSheetDayEntries []TimesheetDayEntry `json:"timeSheetDayEntries"`
}

// TimesheetData represents the timesheet data structure
type TimesheetData struct {
	UserTimeSheets struct {
		TimeSheetEntries []TimesheetEntry `json:"timeSheetEntries"`
	} `json:"userTimeSheets"`
}

// ParseTimesheetData parses the timesheet response and extracts shifts
func ParseTimesheetData(rawData map[string]interface{}) ([]TimesheetShift, error) {
	var shifts []TimesheetShift

	// Navigate through the nested structure
	userTimeSheets, ok := rawData["userTimeSheets"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("userTimeSheets not found in response")
	}

	timeSheetEntries, ok := userTimeSheets["timeSheetEntries"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("timeSheetEntries not found in response")
	}

	for _, entryInterface := range timeSheetEntries {
		entry, ok := entryInterface.(map[string]interface{})
		if !ok {
			continue
		}

		timeSheetDayEntries, ok := entry["timeSheetDayEntries"].([]interface{})
		if !ok {
			continue
		}

		for _, dayEntryInterface := range timeSheetDayEntries {
			dayEntry, ok := dayEntryInterface.(map[string]interface{})
			if !ok {
				continue
			}

			date, ok := dayEntry["date"].(string)
			if !ok {
				continue
			}

			shiftsInterface, ok := dayEntry["shifts"].([]interface{})
			if !ok {
				continue
			}

			for _, shiftInterface := range shiftsInterface {
				shift, ok := shiftInterface.(map[string]interface{})
				if !ok {
					continue
				}

				// Extract punch tag name
				punchTag, ok := shift["punchTag"].(map[string]interface{})
				if !ok {
					continue
				}
				tagName, ok := punchTag["name"].(string)
				if !ok {
					tagName = "Unknown"
				}

				// Extract punch in time
				punchIn, ok := shift["punchIn"].(map[string]interface{})
				if !ok {
					continue
				}
				timestampWithTimezone, ok := punchIn["timestampWithTimezone"].(map[string]interface{})
				if !ok {
					continue
				}
				startTimestamp, ok := timestampWithTimezone["timestamp"].(float64)
				if !ok {
					continue
				}

				// Extract punch out time
				punchOut, ok := shift["punchOut"].(map[string]interface{})
				if !ok {
					continue
				}
				punchOutTimestampWithTimezone, ok := punchOut["timestampWithTimezone"].(map[string]interface{})
				if !ok {
					continue
				}
				endTimestamp, ok := punchOutTimestampWithTimezone["timestamp"].(float64)
				if !ok {
					continue
				}

				// Extract employee notes
				employeeNotes, _ := shift["employeeNotes"].(string)

				// Convert timestamps to time.Time
				startTime := time.Unix(int64(startTimestamp), 0)
				endTime := time.Unix(int64(endTimestamp), 0)

				// Calculate duration
				duration := endTime.Sub(startTime)
				totalHours := duration.Hours()

				// Parse date for display
				parsedDate, err := time.Parse("2006-01-02", date)
				if err != nil {
					continue
				}

				shifts = append(shifts, TimesheetShift{
					Date:          parsedDate.Format("Mon 1/2"),
					Type:          tagName,
					StartTime:     startTime,
					EndTime:       endTime,
					TotalHours:    totalHours,
					EmployeeNotes: employeeNotes,
					ActualDate:    parsedDate,
				})
			}
		}
	}

	// Sort shifts by date and start time
	sort.Slice(shifts, func(i, j int) bool {
		if !shifts[i].ActualDate.Equal(shifts[j].ActualDate) {
			return shifts[i].ActualDate.Before(shifts[j].ActualDate)
		}
		return shifts[i].StartTime.Before(shifts[j].StartTime)
	})

	return shifts, nil
}

// FormatTimesheetTable formats the shifts into a table display
func FormatTimesheetTable(shifts []TimesheetShift, verbose bool) string {
	if len(shifts) == 0 {
		return "No shifts found for the specified date range."
	}

	var builder strings.Builder

	// Header
	builder.WriteString("Date           Type                      Start Time           End Time             Duration    Notes\n")
	builder.WriteString("----           ----                      ----------           --------             --------    -----\n")

	// Rows
	for _, shift := range shifts {
		startTimeStr := shift.StartTime.Format("03:04 PM")
		endTimeStr := shift.EndTime.Format("03:04 PM")

		// Format duration as HH:MM
		hours := int(shift.TotalHours)
		minutes := int((shift.TotalHours - float64(hours)) * 60)
		durationStr := fmt.Sprintf("%02d:%02d", hours, minutes)

		// Handle notes based on verbose mode
		var notes string
		if verbose {
			notes = shift.EmployeeNotes
		} else {
			notes = truncateString(shift.EmployeeNotes, 50)
		}

		// Format the row with proper alignment including notes column
		row := fmt.Sprintf("%-14s %-25s %-20s %-20s %-10s %s\n",
			shift.Date,
			truncateString(shift.Type, 24),
			startTimeStr,
			endTimeStr,
			durationStr,
			notes)

		builder.WriteString(row)
	}

	return builder.String()
}

// truncateString truncates a string to the specified length and adds "..." if needed
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
