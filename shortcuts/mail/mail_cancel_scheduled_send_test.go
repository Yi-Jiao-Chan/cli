// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"errors"
	"testing"

	"github.com/larksuite/cli/internal/httpmock"
	"github.com/larksuite/cli/internal/output"
)

func TestCancelScheduledSend_Success(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)
	reg.Register(&httpmock.Stub{
		URL: "/user_mailboxes/me/messages/msg_sched_001/cancel_scheduled_send",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{},
		},
	})

	err := runMountedMailShortcut(t, MailCancelScheduledSend, []string{
		"+cancel-scheduled-send", "--message-id", "msg_sched_001",
	}, f, stdout)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["message_id"] != "msg_sched_001" {
		t.Errorf("expected message_id msg_sched_001, got %v", data["message_id"])
	}
	if data["status"] != "cancelled" {
		t.Errorf("expected status cancelled, got %v", data["status"])
	}
	if data["restored_as_draft"] != true {
		t.Errorf("expected restored_as_draft true, got %v", data["restored_as_draft"])
	}
}

func TestCancelScheduledSend_MissingMessageID(t *testing.T) {
	f, stdout, _, _ := mailShortcutTestFactory(t)
	err := runMountedMailShortcut(t, MailCancelScheduledSend, []string{
		"+cancel-scheduled-send",
	}, f, stdout)
	if err == nil {
		t.Fatal("expected error for missing --message-id, got nil")
	}
}

func TestCancelScheduledSend_APIError(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)
	reg.Register(&httpmock.Stub{
		URL: "/user_mailboxes/me/messages/msg_invalid/cancel_scheduled_send",
		Body: map[string]interface{}{
			"code": 99991,
			"msg":  "message not found",
		},
	})

	err := runMountedMailShortcut(t, MailCancelScheduledSend, []string{
		"+cancel-scheduled-send", "--message-id", "msg_invalid",
	}, f, stdout)
	if err == nil {
		t.Fatal("expected error for API failure, got nil")
	}
	var exitErr *output.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
}

func TestCancelScheduledSend_CustomMailboxID(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)
	reg.Register(&httpmock.Stub{
		URL: "/user_mailboxes/mailbox_abc/messages/msg_sched_002/cancel_scheduled_send",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{},
		},
	})

	err := runMountedMailShortcut(t, MailCancelScheduledSend, []string{
		"+cancel-scheduled-send", "--message-id", "msg_sched_002", "--user-mailbox-id", "mailbox_abc",
	}, f, stdout)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["message_id"] != "msg_sched_002" {
		t.Errorf("expected message_id msg_sched_002, got %v", data["message_id"])
	}
}
