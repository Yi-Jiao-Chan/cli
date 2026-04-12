// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"context"
	"fmt"
	"io"

	"github.com/larksuite/cli/shortcuts/common"
)

var MailCancelScheduledSend = common.Shortcut{
	Service:     "mail",
	Command:     "+cancel-scheduled-send",
	Description: "Cancel a scheduled draft send. The message will be moved back to drafts.",
	Risk:        "write",
	Scopes:      []string{"mail:user_mailbox.message:send"},
	AuthTypes:   []string{"user"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "mailbox", Default: "me", Desc: "Mailbox email address (default: me)"},
		{Name: "message-id", Desc: "Required. The message ID (messageBizID) of the scheduled message", Required: true},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		mailboxID := resolveMailboxID(runtime)
		messageID := runtime.Str("message-id")
		return common.NewDryRunAPI().
			Desc("Cancel a scheduled email send — the message will be moved back to drafts").
			POST(mailboxPath(mailboxID, "messages", messageID, "cancel_scheduled_send"))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		mailboxID := resolveMailboxID(runtime)
		messageID := runtime.Str("message-id")
		if messageID == "" {
			return fmt.Errorf("--message-id is required")
		}

		hintIdentityFirst(runtime, mailboxID)

		path := mailboxPath(mailboxID, "messages", messageID, "cancel_scheduled_send")
		result, err := runtime.CallAPI("POST", path, nil, nil)
		if err != nil {
			return fmt.Errorf("cancel scheduled send failed: %w", err)
		}

		out := map[string]interface{}{
			"message_id": messageID,
			"status":     "cancelled",
			"result":     result,
		}
		runtime.OutFormat(out, nil, func(w io.Writer) {
			fmt.Fprintln(w, "Scheduled send cancelled successfully.")
			fmt.Fprintf(w, "message_id: %s\n", messageID)
			fmt.Fprintln(w, "The message has been moved back to drafts.")
		})
		return nil
	},
}
