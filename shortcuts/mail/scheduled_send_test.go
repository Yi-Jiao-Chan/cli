// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/larksuite/cli/internal/output"
)

func TestParseAndValidateSendTime_Empty(t *testing.T) {
	result, err := parseAndValidateSendTime("")
	if err != nil {
		t.Fatalf("expected no error for empty string, got: %v", err)
	}
	if result != "" {
		t.Fatalf("expected empty result for empty string, got: %q", result)
	}
}

func TestParseAndValidateSendTime_ValidRFC3339(t *testing.T) {
	future := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	result, err := parseAndValidateSendTime(future)
	if err != nil {
		t.Fatalf("expected no error for valid RFC 3339 time, got: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result for valid RFC 3339 time")
	}
}

func TestParseAndValidateSendTime_ValidWithTimezone(t *testing.T) {
	future := time.Now().Add(2 * time.Hour)
	// Use a specific timezone offset
	loc := time.FixedZone("CST", 8*60*60)
	timeStr := future.In(loc).Format(time.RFC3339)
	result, err := parseAndValidateSendTime(timeStr)
	if err != nil {
		t.Fatalf("expected no error for valid time with timezone, got: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestParseAndValidateSendTime_WithoutTimezoneDefaultsToUTC(t *testing.T) {
	// Use a time far enough in the future to pass the 5-minute check
	future := time.Now().UTC().Add(1 * time.Hour)
	timeStr := future.Format("2006-01-02T15:04:05")
	result, err := parseAndValidateSendTime(timeStr)
	if err != nil {
		t.Fatalf("expected no error for time without timezone, got: %v", err)
	}
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	// The result should be valid RFC 3339
	if _, parseErr := time.Parse(time.RFC3339, result); parseErr != nil {
		t.Fatalf("result %q is not valid RFC 3339: %v", result, parseErr)
	}
}

func TestParseAndValidateSendTime_InvalidFormat(t *testing.T) {
	_, err := parseAndValidateSendTime("not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
	var exitErr *output.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if !strings.Contains(exitErr.Error(), "Invalid time format") {
		t.Errorf("expected 'Invalid time format' in error, got: %v", exitErr)
	}
}

func TestParseAndValidateSendTime_TooSoon(t *testing.T) {
	// Time that is only 1 minute in the future — should fail the 5-minute check
	soon := time.Now().Add(1 * time.Minute).Format(time.RFC3339)
	_, err := parseAndValidateSendTime(soon)
	if err == nil {
		t.Fatal("expected error for time too soon, got nil")
	}
	var exitErr *output.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if !strings.Contains(exitErr.Error(), "at least 5 minutes") {
		t.Errorf("expected '5 minutes' in error, got: %v", exitErr)
	}
}

func TestParseAndValidateSendTime_PastTime(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	_, err := parseAndValidateSendTime(past)
	if err == nil {
		t.Fatal("expected error for past time, got nil")
	}
}

func TestFormatScheduledTimeHuman_ValidTime(t *testing.T) {
	future := time.Now().Add(2 * time.Hour).Format(time.RFC3339)
	result := formatScheduledTimeHuman(future)
	if !strings.Contains(result, "in ") {
		t.Errorf("expected relative time description, got: %q", result)
	}
	if !strings.Contains(result, future) {
		t.Errorf("expected original time in output, got: %q", result)
	}
}

func TestFormatScheduledTimeHuman_InvalidTime(t *testing.T) {
	result := formatScheduledTimeHuman("not-a-date")
	if result != "not-a-date" {
		t.Errorf("expected passthrough for invalid time, got: %q", result)
	}
}

func TestFormatScheduledTimeHuman_DaysAway(t *testing.T) {
	future := time.Now().Add(72 * time.Hour).Format(time.RFC3339)
	result := formatScheduledTimeHuman(future)
	if !strings.Contains(result, "days") {
		t.Errorf("expected 'days' in output, got: %q", result)
	}
}

func TestFormatScheduledTimeHuman_MinutesAway(t *testing.T) {
	future := time.Now().Add(30 * time.Minute).Format(time.RFC3339)
	result := formatScheduledTimeHuman(future)
	if !strings.Contains(result, "minutes") {
		t.Errorf("expected 'minutes' in output, got: %q", result)
	}
}
