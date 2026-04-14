// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"testing"

	"github.com/larksuite/cli/internal/httpmock"
)

func TestCancelScheduledSend_Success(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/messages/msg_sched123/cancel_scheduled_send",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{},
		},
	})

	err := runMountedMailShortcut(t, MailCancelScheduledSend, []string{
		"+cancel-scheduled-send",
		"--message-id", "msg_sched123",
	}, f, stdout)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["message_id"] != "msg_sched123" {
		t.Fatalf("expected message_id=msg_sched123, got %v", data["message_id"])
	}
	if data["status"] != "cancelled" {
		t.Fatalf("expected status=cancelled, got %v", data["status"])
	}
	if data["restored_as_draft"] != true {
		t.Fatalf("expected restored_as_draft=true, got %v", data["restored_as_draft"])
	}
}

func TestCancelScheduledSend_WithMailboxID(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/mbx_custom/messages/msg_sched456/cancel_scheduled_send",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{},
		},
	})

	err := runMountedMailShortcut(t, MailCancelScheduledSend, []string{
		"+cancel-scheduled-send",
		"--message-id", "msg_sched456",
		"--user-mailbox-id", "mbx_custom",
	}, f, stdout)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["message_id"] != "msg_sched456" {
		t.Fatalf("expected message_id=msg_sched456, got %v", data["message_id"])
	}
}

func TestCancelScheduledSend_MissingMessageID(t *testing.T) {
	f, stdout, _, _ := mailShortcutTestFactory(t)

	err := runMountedMailShortcut(t, MailCancelScheduledSend, []string{
		"+cancel-scheduled-send",
	}, f, stdout)

	if err == nil {
		t.Fatal("expected error for missing --message-id")
	}
}

func TestCancelScheduledSend_APIError(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/messages/msg_bad/cancel_scheduled_send",
		Body: map[string]interface{}{
			"code": 99991,
			"msg":  "message not found",
		},
	})

	err := runMountedMailShortcut(t, MailCancelScheduledSend, []string{
		"+cancel-scheduled-send",
		"--message-id", "msg_bad",
	}, f, stdout)

	if err == nil {
		t.Fatal("expected error for API failure")
	}
}
