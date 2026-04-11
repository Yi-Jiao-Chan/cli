// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"context"
	"fmt"

	"github.com/larksuite/cli/shortcuts/common"
)

var MailCancelScheduledSend = common.Shortcut{
	Service:     "mail",
	Command:     "+cancel-scheduled-send",
	Description: "Cancel a scheduled email that has not been sent yet. Requires --message-id of the scheduled message.",
	Risk:        "write",
	Scopes:      []string{"mail:user_mailbox.message:modify", "mail:user_mailbox:readonly"},
	AuthTypes:   []string{"user"},
	Flags: []common.Flag{
		{Name: "mailbox", Desc: "Mailbox email address (default: me)"},
		{Name: "message-id", Desc: "Required. The message ID of the scheduled email to cancel.", Required: true},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		mailboxID := resolveMailboxID(runtime)
		messageID := runtime.Str("message-id")
		return common.NewDryRunAPI().
			Desc("Cancel a scheduled email send").
			POST(mailboxPath(mailboxID, "drafts", messageID, "cancel_scheduled_send"))
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		messageID := runtime.Str("message-id")
		if messageID == "" {
			return fmt.Errorf("--message-id is required")
		}
		return nil
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		mailboxID := resolveMailboxID(runtime)
		messageID := runtime.Str("message-id")

		_, err := runtime.CallAPI("POST", mailboxPath(mailboxID, "drafts", messageID, "cancel_scheduled_send"), nil, nil)
		if err != nil {
			return fmt.Errorf("failed to cancel scheduled send: %w", err)
		}
		runtime.Out(map[string]interface{}{
			"message_id": messageID,
			"status":     "cancelled",
		}, nil)
		return nil
	},
}
