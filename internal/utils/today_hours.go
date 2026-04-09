package utils

import (
	"strconv"
	"time"

	"connectcli/internal/api"
	"connectcli/internal/credentials"
)

// TotalHoursTodayIncludingOpenShift sums completed shift hours for today (Asia/Kolkata calendar day)
// plus the current open punch duration (openElapsed). If a parsed shift matches openPunchID, its
// stored duration is replaced by openElapsed so the ongoing punch is counted correctly.
func TotalHoursTodayIncludingOpenShift(creds *credentials.Credentials, punchClockObjectID int, loc *time.Location, openPunchID string, openElapsed time.Duration) (float64, error) {
	today := time.Now().In(loc).Format("2006-01-02")

	tc := api.NewTimesheetClient()
	resp, err := tc.FetchTimesheet(creds, strconv.Itoa(punchClockObjectID), today, today)
	if err != nil {
		return 0, err
	}
	if len(resp.Data.RawData) == 0 {
		return openElapsed.Hours(), nil
	}

	shifts, err := ParseTimesheetData(resp.Data.RawData)
	if err != nil {
		return openElapsed.Hours(), nil
	}

	var sum float64
	foundOpen := false
	for _, s := range shifts {
		if s.ActualDate.Format("2006-01-02") != today {
			continue
		}
		if openPunchID != "" && s.PunchId == openPunchID {
			foundOpen = true
			sum += openElapsed.Hours()
			continue
		}
		sum += s.TotalHours
	}
	if openPunchID != "" && !foundOpen {
		sum += openElapsed.Hours()
	}
	return sum, nil
}
