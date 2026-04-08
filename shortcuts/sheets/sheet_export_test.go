// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package sheets

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/httpmock"
)

func TestSheetExportRejectsOverwriteWithoutFlag(t *testing.T) {
	f, _, _, _ := cmdutil.TestFactory(t, sheetsTestConfig())

	tmpDir := t.TempDir()
	cmdutil.TestChdir(t, tmpDir)

	if err := os.WriteFile("report.xlsx", []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	// The overwrite check happens before any API call, so no HTTP stubs needed.
	err := mountAndRunSheets(t, SheetExport, []string{
		"+export",
		"--spreadsheet-token", "shtTOKEN",
		"--file-extension", "xlsx",
		"--output-path", "report.xlsx",
		"--as", "user",
	}, f, nil)
	if err == nil {
		t.Fatal("expected overwrite protection error, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSheetExportAllowsOverwriteWithFlag(t *testing.T) {
	f, _, _, reg := cmdutil.TestFactory(t, sheetsTestConfig())

	// Register stubs for the export task creation API.
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/drive/v1/export_tasks",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"ticket": "tkt_123"},
		},
	})
	// Register stub for the poll API (returns completed immediately).
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "/open-apis/drive/v1/export_tasks/tkt_123",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"result": map[string]interface{}{
					"file_token": "box_export_123",
					"file_name":  "report.xlsx",
					"file_size":  100,
				},
			},
		},
	})
	// Register stub for the download API.
	reg.Register(&httpmock.Stub{
		Method:  "GET",
		URL:     "/open-apis/drive/v1/export_tasks/file/box_export_123/download",
		RawBody: []byte("new-content"),
		Headers: http.Header{
			"Content-Type": []string{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		},
	})

	tmpDir := t.TempDir()
	cmdutil.TestChdir(t, tmpDir)

	if err := os.WriteFile("report.xlsx", []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	err := mountAndRunSheets(t, SheetExport, []string{
		"+export",
		"--spreadsheet-token", "shtTOKEN",
		"--file-extension", "xlsx",
		"--output-path", "report.xlsx",
		"--overwrite",
		"--as", "user",
	}, f, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile("report.xlsx")
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(data) != "new-content" {
		t.Fatalf("file content = %q, want %q", string(data), "new-content")
	}
}

func TestSheetExportNoOutputPathReturnsFileToken(t *testing.T) {
	// When --output-path is not provided, Execute should return the
	// file_token JSON without downloading.
	f, stdout, _, reg := cmdutil.TestFactory(t, sheetsTestConfig())

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/drive/v1/export_tasks",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"ticket": "tkt_456"},
		},
	})
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "/open-apis/drive/v1/export_tasks/tkt_456",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"result": map[string]interface{}{
					"file_token": "box_export_456",
				},
			},
		},
	})

	err := mountAndRunSheets(t, SheetExport, []string{
		"+export",
		"--spreadsheet-token", "shtTOKEN",
		"--file-extension", "xlsx",
		"--as", "user",
	}, f, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "box_export_456") {
		t.Fatalf("stdout should contain file_token, got: %s", stdout.String())
	}
}
