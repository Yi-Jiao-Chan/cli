// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/larksuite/cli/errs"
	"github.com/larksuite/cli/shortcuts/common"
	draftpkg "github.com/larksuite/cli/shortcuts/mail/draft"
	"github.com/larksuite/cli/shortcuts/mail/emlbuilder"
	"github.com/larksuite/cli/shortcuts/mail/signature"
)

// signatureFlag is the common flag definition for --signature-id, shared by all compose shortcuts.
var signatureFlag = common.Flag{
	Name: "signature-id",
	Desc: "Optional. Signature ID to append after body content. Overrides the mailbox's default send/reply signature, which is appended automatically when omitted. Run `mail +signature` to list available signatures.",
}

// noSignatureFlag is the common flag definition for --no-signature, shared by
// all compose shortcuts. When set, no signature is appended — neither an
// explicit --signature-id nor the mailbox's auto-resolved default signature.
var noSignatureFlag = common.Flag{
	Name: "no-signature",
	Type: "bool",
	Desc: "Do not append any signature, including the account's default send/reply signature.",
}

// signatureKind selects which default signature to resolve for an address:
// the send-mail default or the reply default.
type signatureKind int

const (
	sigKindSend  signatureKind = iota // send / draft-create / forward
	sigKindReply                      // reply / reply-all
)

// resolveDefaultSignatureID returns the default signature ID for the sender
// address, or "" when there is no default. It is non-fatal by design: on any
// lookup error it returns "" so the caller silently degrades to "no signature"
// rather than failing the send (the user did not explicitly request a
// signature). senderEmail should be the resolved sender address (alias-aware);
// pass "" to fall back to the primary address' default.
func resolveDefaultSignatureID(runtime *common.RuntimeContext, mailboxID, senderEmail string, kind signatureKind) string {
	resp, err := signature.ListAll(runtime, mailboxID) // shares the per-mailbox processCache with resolveSignature
	if err != nil || resp == nil {
		// Non-fatal: an auto-resolution failure must not block the send.
		fmt.Fprintf(runtime.IO().ErrOut, "warning: could not resolve default signature: %v\n", err)
		return ""
	}
	pick := func(u signature.SignatureUsage) string {
		if kind == sigKindReply {
			return u.ReplySignatureID
		}
		return u.SendMailSignatureID
	}
	// 1) Exact match on the sender address (case-insensitive), to honour
	//    alias / send_as scenarios.
	if senderEmail != "" {
		for _, u := range resp.Usages {
			if strings.EqualFold(u.EmailAddress, senderEmail) {
				return normalizeSigID(pick(u))
			}
		}
	}
	// 2) Fall back to the first usage (primary address default).
	if len(resp.Usages) > 0 {
		return normalizeSigID(pick(resp.Usages[0]))
	}
	return ""
}

// normalizeSigID treats "0" and blank as "no default", mirroring the
// is_send_default / is_reply_default normalisation in mail_signature.go.
func normalizeSigID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" || id == "0" {
		return ""
	}
	return id
}

// signatureResult holds the pre-processed signature data ready for HTML injection.
type signatureResult struct {
	ID              string
	RenderedContent string
	Images          []draftpkg.SignatureImage
}

// resolveSignature fetches, interpolates, and downloads images for a signature.
// fromEmail is the --from address (may be an alias); used to match the correct
// sender identity for template interpolation. Pass "" to use the primary address.
func resolveSignature(ctx context.Context, runtime *common.RuntimeContext, mailboxID, signatureID, fromEmail string) (*signatureResult, error) {
	if signatureID == "" {
		return nil, nil
	}

	sig, err := signature.Get(runtime, mailboxID, signatureID)
	if err != nil {
		return nil, err
	}

	// Resolve sender info for template interpolation.
	lang := resolveLang(runtime)
	senderName, senderEmail := resolveSenderInfo(runtime, mailboxID, fromEmail)
	rendered := signature.InterpolateTemplate(sig, lang, senderName, senderEmail)

	// Download signature inline images. The file_key field contains a
	// direct download URL provided by the mail backend.
	var images []draftpkg.SignatureImage
	for _, img := range sig.Images {
		if img.DownloadURL == "" || img.CID == "" {
			continue
		}
		data, ct, err := downloadSignatureImage(runtime, img.DownloadURL, img.ImageName)
		if err != nil {
			return nil, mailDecorateProblemMessage(err, "failed to download signature image %s", img.ImageName)
		}
		images = append(images, draftpkg.SignatureImage{
			CID:         img.CID,
			ContentType: ct,
			FileName:    img.ImageName,
			Data:        data,
		})
	}

	return &signatureResult{
		ID:              sig.ID,
		RenderedContent: rendered,
		Images:          images,
	}, nil
}

// injectSignatureIntoBody inserts signature HTML into the body, placing
// it right after the user-authored region and before any system-managed
// tail (large attachment card or quote block). Any existing signature is
// removed first. Returns the new full HTML body.
//
// Delegates to draftpkg.PlaceSignatureBeforeSystemTail for the actual
// placement, sharing a single source of truth with the edit-time
// insert_signature op so both paths yield identical structure.
func injectSignatureIntoBody(bodyHTML string, sig *signatureResult) string {
	if sig == nil {
		return bodyHTML
	}
	sigBlock := draftpkg.SignatureSpacing() + draftpkg.BuildSignatureHTML(sig.ID, sig.RenderedContent)
	return draftpkg.PlaceSignatureBeforeSystemTail(bodyHTML, sigBlock)
}

// addSignatureImagesToBuilder adds signature inline images to the EML builder.
func addSignatureImagesToBuilder(bld emlbuilder.Builder, sig *signatureResult) emlbuilder.Builder {
	if sig == nil {
		return bld
	}
	for _, img := range sig.Images {
		cid := normalizeInlineCID(img.CID)
		if cid == "" {
			continue
		}
		bld = bld.AddInline(img.Data, img.ContentType, img.FileName, cid)
	}
	return bld
}

// resolveSenderInfo fetches send_as addresses and returns the name/email
// for signature interpolation. If fromEmail is non-empty, it matches
// that address in the sendable list (for alias/send_as scenarios);
// otherwise falls back to the first (primary) address.
func resolveSenderInfo(runtime *common.RuntimeContext, mailboxID, fromEmail string) (name, email string) {
	data, err := runtime.CallAPITyped("GET", mailboxPath(mailboxID, "settings", "send_as"), nil, nil)
	if err != nil {
		return "", ""
	}
	addrs, ok := data["sendable_addresses"].([]interface{})
	if !ok || len(addrs) == 0 {
		return "", ""
	}
	// If fromEmail is specified, find the matching address.
	if fromEmail != "" {
		for _, a := range addrs {
			m, ok := a.(map[string]interface{})
			if !ok {
				continue
			}
			e, _ := m["email_address"].(string)
			if strings.EqualFold(e, fromEmail) {
				n, _ := m["name"].(string)
				return n, e
			}
		}
	}
	// Fall back to the first sendable address (primary).
	first, ok := addrs[0].(map[string]interface{})
	if !ok {
		return "", ""
	}
	n, _ := first["name"].(string)
	e, _ := first["email_address"].(string)
	return n, e
}

// downloadSignatureImage downloads a signature image by its direct URL.
// Security: enforces https, does not send Bearer token (URL is pre-signed),
// uses context timeout, and limits response size. Aligned with
// downloadAttachmentContent in helpers.go.
func downloadSignatureImage(runtime *common.RuntimeContext, downloadURL, filename string) ([]byte, string, error) {
	u, err := url.Parse(downloadURL)
	if err != nil {
		return nil, "", mailInvalidResponseError("signature image download: invalid URL: %v", err).WithCause(err)
	}
	if u.Scheme != "https" {
		return nil, "", mailInvalidResponseError("signature image download: URL must use https (got %q)", u.Scheme)
	}
	if u.Host == "" {
		return nil, "", mailInvalidResponseError("signature image download: URL has no host")
	}

	httpClient, err := runtime.Factory.HttpClient()
	if err != nil {
		return nil, "", errs.NewInternalError(errs.SubtypeSDKError, "signature image download: %v", err).WithCause(err)
	}
	ctx, cancel := context.WithTimeout(runtime.Ctx(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, "", errs.NewInternalError(errs.SubtypeSDKError, "signature image download: %v", err).WithCause(err)
	}
	// Do NOT send Authorization: the download URL is pre-signed.

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, "", errs.NewNetworkError(errs.SubtypeNetworkTransport, "signature image download: %v", err).WithCause(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if resp.StatusCode >= 500 {
			return nil, "", errs.NewNetworkError(errs.SubtypeNetworkServer, "signature image download: HTTP %d: %s", resp.StatusCode, string(body)).
				WithCode(resp.StatusCode).
				WithRetryable()
		}
		subtype := errs.SubtypeUnknown
		if resp.StatusCode == http.StatusNotFound {
			subtype = errs.SubtypeNotFound
		}
		return nil, "", errs.NewAPIError(subtype, "signature image download: HTTP %d: %s", resp.StatusCode, string(body)).WithCode(resp.StatusCode)
	}

	const maxSize = 10 * 1024 * 1024
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return nil, "", errs.NewNetworkError(errs.SubtypeNetworkTransport, "signature image download: read body: %v", err).WithCause(err)
	}
	if len(data) > maxSize {
		return nil, "", mailFailedPreconditionError("signature image download: file exceeds 10MB limit")
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" || ct == "application/octet-stream" {
		ct = contentTypeFromFilename(filename)
	}

	return data, ct, nil
}

func contentTypeFromFilename(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".bmp":
		return "image/bmp"
	default:
		return "application/octet-stream"
	}
}

// signatureCIDs returns the CID list from a signatureResult, for inline CID validation.
func signatureCIDs(sig *signatureResult) []string {
	if sig == nil {
		return nil
	}
	cids := make([]string, 0, len(sig.Images))
	for _, img := range sig.Images {
		cid := normalizeInlineCID(img.CID)
		if cid != "" {
			cids = append(cids, cid)
		}
	}
	return cids
}

// validateSignatureFlags returns an error if both --no-signature and
// --signature-id are set: opting out of all signatures while also naming an
// explicit one is contradictory. Plain-text mode with a signature is allowed —
// the signature is downgraded to plain text and appended (see
// appendPlainTextSignature).
func validateSignatureFlags(noSignature bool, signatureID string) error {
	if noSignature && signatureID != "" {
		return mailValidationError("--no-signature and --signature-id are mutually exclusive").
			WithParams(
				mailInvalidParam("--no-signature", "mutually exclusive with --signature-id"),
				mailInvalidParam("--signature-id", "mutually exclusive with --no-signature"),
			)
	}
	return nil
}

// htmlBoundaryTagRe matches the block / line-break tags whose boundaries map to
// a newline when downgrading signature HTML to plain text.
var htmlBoundaryTagRe = regexp.MustCompile(`(?i)</?(?:br|div|p|tr)(?:\s[^>]*)?/?>`)

// htmlImgTagRe matches <img ...> tags, which are dropped entirely in plain text
// (plain-text mail carries no inline images).
var htmlImgTagRe = regexp.MustCompile(`(?i)<img(?:\s[^>]*)?/?>`)

// htmlAnyTagRe matches any remaining tag, removed without a substitute.
var htmlAnyTagRe = regexp.MustCompile(`<[^>]*>`)

// excessNewlinesRe collapses runs of 2+ newlines into a single newline.
// Block / line-break boundaries each map to one newline, so adjacent boundary
// tags (e.g. <br/><div>) would otherwise leave blank lines; signatures render
// as a tidy single-newline-separated block.
var excessNewlinesRe = regexp.MustCompile(`\n{2,}`)

// signatureToPlainText downgrades interpolated signature HTML to plain text:
//   - <br> / <div> / <p> / <tr> boundaries become newlines
//   - <img> is dropped (plain-text mail has no inline images)
//   - all other tags are stripped
//   - basic HTML entities are decoded
//   - runs of blank lines are collapsed
func signatureToPlainText(renderedHTML string) string {
	s := renderedHTML
	// Drop images first so their boundary attributes don't leak as text.
	s = htmlImgTagRe.ReplaceAllString(s, "")
	// Map block / line-break boundaries to newlines.
	s = htmlBoundaryTagRe.ReplaceAllString(s, "\n")
	// Strip every remaining tag.
	s = htmlAnyTagRe.ReplaceAllString(s, "")
	// Decode basic HTML entities (&amp; &lt; &gt; &quot; &#34; &nbsp; ...).
	s = html.UnescapeString(s)
	// Normalise newlines and trim trailing whitespace per line.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lines[i] = strings.TrimRight(ln, " \t")
	}
	s = strings.Join(lines, "\n")
	// Collapse runs of blank lines so the signature is a tidy single-newline
	// separated block, then trim the edges.
	s = excessNewlinesRe.ReplaceAllString(s, "\n")
	return strings.TrimSpace(s)
}

// appendPlainTextSignature appends a plain-text rendition of the signature to
// the plain-text body, separated by a blank line. Returns textBody unchanged
// when there is no signature or the downgraded signature is empty.
func appendPlainTextSignature(textBody string, sig *signatureResult) string {
	if sig == nil {
		return textBody
	}
	txt := signatureToPlainText(sig.RenderedContent)
	if strings.TrimSpace(txt) == "" {
		return textBody
	}
	const sep = "\n\n" // align with the client's plain-text signature spacing (one blank line)
	return strings.TrimRight(textBody, "\n") + sep + txt
}
