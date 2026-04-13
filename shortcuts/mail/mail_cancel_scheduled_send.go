// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"context"
	"fmt"
	"net/url"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/shortcuts/common"
)

var MailCancelScheduledSend = common.Shortcut{
	Service:     "mail",
	Command:     "+cancel-scheduled-send",
	Description: "Cancel a scheduled email send. The email will be restored as a draft.",
	Risk:        "write",
	Scopes:      []string{"mail:user_mailbox.message:send"},
	AuthTypes:   []string{"user"},
	HasFormat:   true,
	Flags: []common.Flag{
		{Name: "message-id", Desc: "Message ID of the scheduled email to cancel (required)", Required: true},
		{Name: "user-mailbox-id", Desc: "User mailbox ID (default: me)"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		if runtime.Str("message-id") == "" {
			return output.ErrValidation("--message-id is required")
		}
		return nil
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		messageID := runtime.Str("message-id")
		userMailboxID := runtime.Str("user-mailbox-id")
		if userMailboxID == "" {
			userMailboxID = "me"
		}
		return common.NewDryRunAPI().
			Desc("Cancel scheduled send — message will be restored as draft").
			POST(fmt.Sprintf("/open-apis/mail/v1/user_mailboxes/%s/messages/%s/cancel_scheduled_send",
				url.PathEscape(userMailboxID),
				url.PathEscape(messageID),
			))
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		messageID := runtime.Str("message-id")
		userMailboxID := runtime.Str("user-mailbox-id")
		if userMailboxID == "" {
			userMailboxID = "me"
		}

		path := fmt.Sprintf("/open-apis/mail/v1/user_mailboxes/%s/messages/%s/cancel_scheduled_send",
			url.PathEscape(userMailboxID),
			url.PathEscape(messageID),
		)

		_, err := runtime.CallAPI("POST", path, nil, nil)
		if err != nil {
			return output.ErrWithHint(output.ExitAPI, "api_error",
				fmt.Sprintf("Failed to cancel scheduled send for message %s", messageID),
				"Ensure the message ID is correct and the email has not already been sent.",
			)
		}

		runtime.Out(map[string]interface{}{
			"message_id":        messageID,
			"status":            "cancelled",
			"restored_as_draft": true,
		}, nil)

		fmt.Fprintf(runtime.IO().ErrOut,
			"tip: the message has been restored as a draft. Use lark-cli mail +draft-edit --id %s to edit.\n",
			sanitizeForTerminal(messageID))

		return nil
	},
}
