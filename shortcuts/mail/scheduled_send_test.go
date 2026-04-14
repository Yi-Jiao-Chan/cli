// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"testing"
	"time"
)

func TestParseAndValidateSendTime_Empty(t *testing.T) {
	got, err := parseAndValidateSendTime("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestParseAndValidateSendTime_ValidRFC3339(t *testing.T) {
	future := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	got, err := parseAndValidateSendTime(future)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestParseAndValidateSendTime_ValidRFC3339WithTimezone(t *testing.T) {
	future := time.Now().Add(2 * time.Hour).In(time.FixedZone("CST", 8*3600)).Format(time.RFC3339)
	got, err := parseAndValidateSendTime(future)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestParseAndValidateSendTime_WithoutTimezoneDefaultsToUTC(t *testing.T) {
	future := time.Now().UTC().Add(1 * time.Hour).Format("2006-01-02T15:04:05")
	got, err := parseAndValidateSendTime(future)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty result")
	}
	// Parse the result to verify it's valid RFC 3339
	parsed, err := time.Parse(time.RFC3339, got)
	if err != nil {
		t.Fatalf("result is not valid RFC 3339: %v", err)
	}
	// Should be in UTC since no timezone was provided
	if parsed.Location().String() != "UTC" {
		t.Fatalf("expected UTC timezone, got %s", parsed.Location())
	}
}

func TestParseAndValidateSendTime_InvalidFormat(t *testing.T) {
	testCases := []string{
		"not-a-date",
		"2026-04-14",
		"2026/04/14T09:00:00",
		"April 14, 2026",
		"1234567890",
	}
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			_, err := parseAndValidateSendTime(tc)
			if err == nil {
				t.Fatalf("expected error for input %q, got nil", tc)
			}
		})
	}
}

func TestParseAndValidateSendTime_TooSoon(t *testing.T) {
	// 1 minute in the future - should fail (minimum 5 minutes)
	tooSoon := time.Now().Add(1 * time.Minute).Format(time.RFC3339)
	_, err := parseAndValidateSendTime(tooSoon)
	if err == nil {
		t.Fatal("expected error for time too soon in the future")
	}
}

func TestParseAndValidateSendTime_PastTime(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	_, err := parseAndValidateSendTime(past)
	if err == nil {
		t.Fatal("expected error for past time")
	}
}

func TestFormatScheduledTimeHuman_ValidTime(t *testing.T) {
	future := time.Now().Add(2 * time.Hour).Format(time.RFC3339)
	result := formatScheduledTimeHuman(future)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	if result == future {
		t.Fatal("expected formatted result, got raw time")
	}
}

func TestFormatScheduledTimeHuman_InvalidTime(t *testing.T) {
	result := formatScheduledTimeHuman("not-a-date")
	if result != "not-a-date" {
		t.Fatalf("expected raw input returned, got %q", result)
	}
}

func TestFormatScheduledTimeHuman_Minutes(t *testing.T) {
	future := time.Now().Add(30 * time.Minute).Format(time.RFC3339)
	result := formatScheduledTimeHuman(future)
	if result == future {
		t.Fatal("expected formatted result with relative time")
	}
}

func TestFormatScheduledTimeHuman_Days(t *testing.T) {
	future := time.Now().Add(48 * time.Hour).Format(time.RFC3339)
	result := formatScheduledTimeHuman(future)
	if result == future {
		t.Fatal("expected formatted result with relative time")
	}
}
