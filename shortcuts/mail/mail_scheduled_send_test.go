// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"bytes"
	"encoding/base64"
	"testing"
	"time"

	"github.com/zalando/go-keyring"

	"github.com/larksuite/cli/internal/auth"
	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/httpmock"
)

// mailShortcutTestFactoryWithSendScope creates a test factory with the send scope included.
func mailShortcutTestFactoryWithSendScope(t *testing.T) (*cmdutil.Factory, *bytes.Buffer, *bytes.Buffer, *httpmock.Registry) {
	t.Helper()
	keyring.MockInit()
	t.Setenv("HOME", t.TempDir())

	cfg := mailTestConfig()
	token := &auth.StoredUAToken{
		UserOpenId:       cfg.UserOpenId,
		AppId:            cfg.AppID,
		AccessToken:      "test-user-access-token",
		RefreshToken:     "test-refresh-token",
		ExpiresAt:        time.Now().Add(1 * time.Hour).UnixMilli(),
		RefreshExpiresAt: time.Now().Add(24 * time.Hour).UnixMilli(),
		Scope:            "mail:user_mailbox.messages:write mail:user_mailbox.messages:read mail:user_mailbox.message:modify mail:user_mailbox.message:readonly mail:user_mailbox.message.address:read mail:user_mailbox.message.subject:read mail:user_mailbox.message.body:read mail:user_mailbox:readonly mail:user_mailbox.message:send",
		GrantedAt:        time.Now().Add(-1 * time.Hour).UnixMilli(),
	}
	if err := auth.SetStoredToken(token); err != nil {
		t.Fatalf("SetStoredToken() error = %v", err)
	}
	t.Cleanup(func() {
		_ = auth.RemoveStoredToken(cfg.AppID, cfg.UserOpenId)
	})

	return cmdutil.TestFactory(t, cfg)
}

// stubScheduledSendEndpoints registers HTTP stubs for profile, draft create, and draft send.
func stubScheduledSendEndpoints(reg *httpmock.Registry) {
	// Profile
	reg.Register(&httpmock.Stub{
		URL: "/user_mailboxes/me/profile",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"primary_email_address": "me@example.com",
			},
		},
	})

	// Draft create
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"draft_id": "draft_sched001",
			},
		},
	})

	// Draft send (for confirm-send)
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts/draft_sched001/send",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"message_id": "msg_sched001",
				"thread_id":  "thread_sched001",
			},
		},
	})
}

// stubScheduledSendEndpointsNoSend registers stubs for profile and draft create only (no send).
func stubScheduledSendEndpointsNoSend(reg *httpmock.Registry) {
	reg.Register(&httpmock.Stub{
		URL: "/user_mailboxes/me/profile",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"primary_email_address": "me@example.com",
			},
		},
	})
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"draft_id": "draft_sched001",
			},
		},
	})
}

// stubReplyScheduledSendEndpoints registers stubs for reply/reply-all/forward with scheduled send.
func stubReplyScheduledSendEndpoints(reg *httpmock.Registry) {
	reg.Register(&httpmock.Stub{
		URL: "/user_mailboxes/me/profile",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"primary_email_address": "me@example.com",
			},
		},
	})

	reg.Register(&httpmock.Stub{
		URL: "/user_mailboxes/me/messages/msg_orig001",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"message": map[string]interface{}{
					"message_id":      "msg_orig001",
					"thread_id":       "thread_orig001",
					"smtp_message_id": "<msg_orig001@example.com>",
					"subject":         "Original Subject",
					"head_from":       map[string]interface{}{"mail_address": "sender@example.com", "name": "Sender"},
					"to":              []map[string]interface{}{{"mail_address": "me@example.com", "name": "Me"}},
					"cc":              []interface{}{},
					"bcc":             []interface{}{},
					"body_html":       base64.URLEncoding.EncodeToString([]byte("<p>Original body</p>")),
					"body_plain_text": base64.URLEncoding.EncodeToString([]byte("Original body")),
					"internal_date":   "1704067200000",
					"attachments":     []interface{}{},
				},
			},
		},
	})

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"draft_id": "draft_sched001",
			},
		},
	})

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts/draft_sched001/send",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"message_id": "msg_sched001",
				"thread_id":  "thread_sched001",
			},
		},
	})
}

// stubReplyScheduledSendEndpointsNoSend registers stubs without draft send (for draft-only tests).
func stubReplyScheduledSendEndpointsNoSend(reg *httpmock.Registry) {
	reg.Register(&httpmock.Stub{
		URL: "/user_mailboxes/me/profile",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"primary_email_address": "me@example.com",
			},
		},
	})

	reg.Register(&httpmock.Stub{
		URL: "/user_mailboxes/me/messages/msg_orig001",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"message": map[string]interface{}{
					"message_id":      "msg_orig001",
					"thread_id":       "thread_orig001",
					"smtp_message_id": "<msg_orig001@example.com>",
					"subject":         "Original Subject",
					"head_from":       map[string]interface{}{"mail_address": "sender@example.com", "name": "Sender"},
					"to":              []map[string]interface{}{{"mail_address": "me@example.com", "name": "Me"}},
					"cc":              []interface{}{},
					"bcc":             []interface{}{},
					"body_html":       base64.URLEncoding.EncodeToString([]byte("<p>Original body</p>")),
					"body_plain_text": base64.URLEncoding.EncodeToString([]byte("Original body")),
					"internal_date":   "1704067200000",
					"attachments":     []interface{}{},
				},
			},
		},
	})

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"draft_id": "draft_sched001",
			},
		},
	})
}

func futureTimeStr() string {
	return time.Now().Add(2 * time.Hour).Format(time.RFC3339)
}

// ---------------------------------------------------------------------------
// +send scheduled send tests
// ---------------------------------------------------------------------------

func TestSend_ScheduledSend_DraftOnly(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)
	stubScheduledSendEndpointsNoSend(reg)

	sendTime := futureTimeStr()
	err := runMountedMailShortcut(t, MailSend, []string{
		"+send",
		"--to", "recipient@example.com",
		"--subject", "Scheduled Test",
		"--body", "<p>Hello</p>",
		"--send-time", sendTime,
	}, f, stdout)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["draft_id"] != "draft_sched001" {
		t.Fatalf("expected draft_id=draft_sched001, got %v", data["draft_id"])
	}
}

func TestSend_ScheduledSend_WithConfirmSend(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactoryWithSendScope(t)
	stubScheduledSendEndpoints(reg)

	sendTime := futureTimeStr()
	err := runMountedMailShortcut(t, MailSend, []string{
		"+send",
		"--to", "recipient@example.com",
		"--subject", "Scheduled Test",
		"--body", "<p>Hello</p>",
		"--send-time", sendTime,
		"--confirm-send",
	}, f, stdout)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["status"] != "scheduled" {
		t.Fatalf("expected status=scheduled, got %v", data["status"])
	}
	if data["message_id"] != "msg_sched001" {
		t.Fatalf("expected message_id=msg_sched001, got %v", data["message_id"])
	}
}

func TestSend_ScheduledSend_InvalidTime(t *testing.T) {
	f, stdout, _, _ := mailShortcutTestFactory(t)

	err := runMountedMailShortcut(t, MailSend, []string{
		"+send",
		"--to", "recipient@example.com",
		"--subject", "Scheduled Test",
		"--body", "<p>Hello</p>",
		"--send-time", "not-a-date",
	}, f, stdout)

	if err == nil {
		t.Fatal("expected error for invalid send-time")
	}
}

func TestSend_ScheduledSend_TooSoon(t *testing.T) {
	f, stdout, _, _ := mailShortcutTestFactory(t)

	tooSoon := time.Now().Add(1 * time.Minute).Format(time.RFC3339)
	err := runMountedMailShortcut(t, MailSend, []string{
		"+send",
		"--to", "recipient@example.com",
		"--subject", "Scheduled Test",
		"--body", "<p>Hello</p>",
		"--send-time", tooSoon,
	}, f, stdout)

	if err == nil {
		t.Fatal("expected error for send-time too soon")
	}
}

// ---------------------------------------------------------------------------
// +reply scheduled send tests
// ---------------------------------------------------------------------------

func TestReply_ScheduledSend_WithConfirmSend(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactoryWithSendScope(t)
	stubReplyScheduledSendEndpoints(reg)

	sendTime := futureTimeStr()
	err := runMountedMailShortcut(t, MailReply, []string{
		"+reply",
		"--message-id", "msg_orig001",
		"--body", "<p>Reply body</p>",
		"--send-time", sendTime,
		"--confirm-send",
	}, f, stdout)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["status"] != "scheduled" {
		t.Fatalf("expected status=scheduled, got %v", data["status"])
	}
}

func TestReply_ScheduledSend_DraftOnly(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)
	stubReplyScheduledSendEndpointsNoSend(reg)

	sendTime := futureTimeStr()
	err := runMountedMailShortcut(t, MailReply, []string{
		"+reply",
		"--message-id", "msg_orig001",
		"--body", "<p>Reply body</p>",
		"--send-time", sendTime,
	}, f, stdout)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["draft_id"] != "draft_sched001" {
		t.Fatalf("expected draft_id=draft_sched001, got %v", data["draft_id"])
	}
}

// ---------------------------------------------------------------------------
// +reply-all scheduled send tests
// ---------------------------------------------------------------------------

func TestReplyAll_ScheduledSend_WithConfirmSend(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactoryWithSendScope(t)
	stubReplyScheduledSendEndpoints(reg)

	sendTime := futureTimeStr()
	err := runMountedMailShortcut(t, MailReplyAll, []string{
		"+reply-all",
		"--message-id", "msg_orig001",
		"--body", "<p>Reply all body</p>",
		"--send-time", sendTime,
		"--confirm-send",
	}, f, stdout)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["status"] != "scheduled" {
		t.Fatalf("expected status=scheduled, got %v", data["status"])
	}
}

func TestReplyAll_ScheduledSend_DraftOnly(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)
	stubReplyScheduledSendEndpointsNoSend(reg)

	sendTime := futureTimeStr()
	err := runMountedMailShortcut(t, MailReplyAll, []string{
		"+reply-all",
		"--message-id", "msg_orig001",
		"--body", "<p>Reply all body</p>",
		"--send-time", sendTime,
	}, f, stdout)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["draft_id"] != "draft_sched001" {
		t.Fatalf("expected draft_id=draft_sched001, got %v", data["draft_id"])
	}
}

// ---------------------------------------------------------------------------
// +forward scheduled send tests
// ---------------------------------------------------------------------------

func TestForward_ScheduledSend_WithConfirmSend(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactoryWithSendScope(t)
	stubReplyScheduledSendEndpoints(reg)

	sendTime := futureTimeStr()
	err := runMountedMailShortcut(t, MailForward, []string{
		"+forward",
		"--message-id", "msg_orig001",
		"--to", "forward-to@example.com",
		"--send-time", sendTime,
		"--confirm-send",
	}, f, stdout)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["status"] != "scheduled" {
		t.Fatalf("expected status=scheduled, got %v", data["status"])
	}
}

func TestForward_ScheduledSend_DraftOnly(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)
	stubReplyScheduledSendEndpointsNoSend(reg)

	sendTime := futureTimeStr()
	err := runMountedMailShortcut(t, MailForward, []string{
		"+forward",
		"--message-id", "msg_orig001",
		"--to", "forward-to@example.com",
		"--send-time", sendTime,
	}, f, stdout)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := decodeShortcutEnvelopeData(t, stdout)
	if data["draft_id"] != "draft_sched001" {
		t.Fatalf("expected draft_id=draft_sched001, got %v", data["draft_id"])
	}
}
