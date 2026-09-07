package application_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"data-toolkit/internal/application"
	"data-toolkit/internal/core"
	"data-toolkit/internal/report"
)

func TestRunReceiptHasExactFields(t *testing.T) {
	typeOfReceipt := reflect.TypeFor[application.RunReceipt]()
	want := []struct {
		name string
		json string
	}{
		{"ReceiptID", "receipt_id"},
		{"RequestID", "request_id"},
		{"RunID", "run_id"},
		{"NormalizedWorkflowDigest", "normalized_workflow_digest"},
		{"BinaryVersion", "binary_version"},
		{"CapabilitiesDigest", "capabilities_digest"},
		{"EffectiveConfigDigest", "effective_config_digest"},
		{"SourceSnapshotHash", "source_snapshot_hash,omitempty"},
		{"PublishedOutput", "published_output,omitempty"},
		{"Status", "status"},
		{"TerminalError", "terminal_error,omitempty"},
		{"StartedAt", "started_at"},
		{"FinishedAt", "finished_at"},
	}
	if typeOfReceipt.NumField() != len(want) {
		t.Fatalf("RunReceipt field count = %d, want %d", typeOfReceipt.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfReceipt.Field(index)
		if field.Name != expected.name || field.Tag.Get("json") != expected.json {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, expected.json)
		}
	}
}

func TestRunReceiptEnforcesConditionalTerminalFields(t *testing.T) {
	digest := core.SHA256Digest(strings.Repeat("a", 64))
	startedAt := time.Date(2026, time.September, 7, 9, 0, 0, 0, time.UTC)
	base := application.RunReceipt{
		ReceiptID:                core.ReceiptID(digest),
		RequestID:                "123e4567-e89b-12d3-a456-426614174000",
		RunID:                    "123e4567-e89b-12d3-a456-426614174001",
		NormalizedWorkflowDigest: digest,
		BinaryVersion:            "v1.0.0",
		CapabilitiesDigest:       digest,
		EffectiveConfigDigest:    digest,
		StartedAt:                startedAt,
		FinishedAt:               startedAt.Add(time.Nanosecond),
	}
	output := report.PublishedOutput{Path: "out.csv", FormatID: core.CSVFormat, SHA256: digest, SizeBytes: 1}
	terminalError := application.APIError{
		Code: "operation_failed", Message: "operation failed", Component: "logical_engine", Details: []report.ErrorDetail{},
	}
	snapshotHash := digest

	succeeded := base
	succeeded.Status = core.TerminalStatusSucceeded
	succeeded.PublishedOutput = &output
	if err := succeeded.Validate(); err != nil {
		t.Fatalf("succeeded RunReceipt.Validate() error = %v", err)
	}

	failed := base
	failed.Status = core.TerminalStatusFailed
	failed.SourceSnapshotHash = &snapshotHash
	failed.TerminalError = &terminalError
	if err := failed.Validate(); err != nil {
		t.Fatalf("failed RunReceipt.Validate() error = %v", err)
	}

	invalid := []application.RunReceipt{
		base,
		func() application.RunReceipt { value := succeeded; value.TerminalError = &terminalError; return value }(),
		func() application.RunReceipt { value := failed; value.PublishedOutput = &output; return value }(),
		func() application.RunReceipt { value := failed; value.TerminalError = nil; return value }(),
	}
	invalid[0].Status = core.TerminalStatusSucceeded
	for index, receipt := range invalid {
		if err := receipt.Validate(); err == nil {
			t.Errorf("invalid receipt %d unexpectedly validated", index)
		}
	}
}

func TestRunReceiptJSONOmitsConditionalFields(t *testing.T) {
	digest := core.SHA256Digest(strings.Repeat("a", 64))
	startedAt := time.Date(2026, time.September, 7, 9, 0, 0, 123456789, time.UTC)
	receipt := application.RunReceipt{
		ReceiptID:                core.ReceiptID(digest),
		RequestID:                "123e4567-e89b-12d3-a456-426614174000",
		RunID:                    "123e4567-e89b-12d3-a456-426614174001",
		NormalizedWorkflowDigest: digest,
		BinaryVersion:            "v1.0.0",
		CapabilitiesDigest:       digest,
		EffectiveConfigDigest:    digest,
		Status:                   core.TerminalStatusSucceeded,
		PublishedOutput:          &report.PublishedOutput{Path: "out.csv", FormatID: core.CSVFormat, SHA256: digest, SizeBytes: 1},
		StartedAt:                startedAt,
		FinishedAt:               startedAt.Add(time.Nanosecond),
	}
	if err := receipt.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"source_snapshot_hash", "terminal_error"} {
		if strings.Contains(string(encoded), `"`+field+`"`) {
			t.Fatalf("succeeded receipt contains %s: %s", field, encoded)
		}
	}
}
