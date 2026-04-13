// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"fmt"
	"time"

	"github.com/larksuite/cli/internal/output"
)

const (
	// minScheduleLeadTime is the minimum time in the future for scheduled send.
	minScheduleLeadTime = 5 * time.Minute
)

// parseAndValidateSendTime parses and validates the --send-time flag value.
// Returns a Unix timestamp string (seconds since epoch) to pass to the API,
// or an error if invalid. If sendTimeStr is empty, returns ("", nil) indicating
// immediate send.
func parseAndValidateSendTime(sendTimeStr string) (string, error) {
	if sendTimeStr == "" {
		return "", nil
	}

	t, err := time.Parse(time.RFC3339, sendTimeStr)
	if err != nil {
		// Try parsing without timezone offset — default to UTC
		t, err = time.Parse("2006-01-02T15:04:05", sendTimeStr)
		if err != nil {
			return "", output.ErrValidation(
				"Invalid time format for --send-time. Use RFC 3339 format, e.g. 2026-04-14T09:00:00+08:00",
			)
		}
		t = t.UTC()
	}

	if time.Until(t) < minScheduleLeadTime {
		return "", output.ErrValidation(
			"Scheduled time must be at least 5 minutes in the future",
		)
	}

	return fmt.Sprintf("%d", t.Unix()), nil
}

// formatScheduledTimeHuman returns a human-readable scheduled time string
// for pretty output, e.g. "2026-04-14T09:00:00+08:00 (Mon, in 14 hours)"
func formatScheduledTimeHuman(sendTime string) string {
	t, err := time.Parse(time.RFC3339, sendTime)
	if err != nil {
		return sendTime
	}
	dur := time.Until(t)
	var relative string
	switch {
	case dur < time.Hour:
		relative = fmt.Sprintf("in %d minutes", int(dur.Minutes()))
	case dur < 24*time.Hour:
		relative = fmt.Sprintf("in %d hours", int(dur.Hours()))
	default:
		relative = fmt.Sprintf("in %d days", int(dur.Hours()/24))
	}
	return fmt.Sprintf("%s (%s, %s)", sendTime, t.Format("Mon"), relative)
}
