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

func TestRunReportHasExactFields(t *testing.T) {
	typeOfReport := reflect.TypeFor[application.RunReport]()
	want := []struct {
		name string
		json string
	}{
		{"RequestID", "request_id"},
		{"RunID", "run_id"},
		{"Status", "status"},
		{"Output", "output,omitempty"},
		{"Validations", "validations"},
		{"Findings", "findings"},
		{"ExceptionSummary", "exception_summary,omitempty"},
		{"FirstFailure", "first_failure,omitempty"},
		{"SourceSnapshotHash", "source_snapshot_hash,omitempty"},
		{"EffectiveConfigDigest", "effective_config_digest"},
		{"CapabilitiesDigest", "capabilities_digest"},
		{"ReceiptID", "receipt_id"},
		{"StartedAt", "started_at"},
		{"FinishedAt", "finished_at"},
	}
	if typeOfReport.NumField() != len(want) {
		t.Fatalf("RunReport field count = %d, want %d", typeOfReport.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfReport.Field(index)
		if field.Name != expected.name || field.Tag.Get("json") != expected.json {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, expected.json)
		}
	}
}

func TestRunReportEnforcesTerminalPresenceRules(t *testing.T) {
	digest := core.SHA256Digest(strings.Repeat("a", 64))
	startedAt := time.Date(2026, time.September, 7, 8, 0, 0, 0, time.UTC)
	base := application.RunReport{
		RequestID:             "123e4567-e89b-12d3-a456-426614174000",
		RunID:                 "123e4567-e89b-12d3-a456-426614174001",
		Validations:           []report.ValidationResult{},
		Findings:              []report.Finding{},
		EffectiveConfigDigest: digest,
		CapabilitiesDigest:    digest,
		ReceiptID:             core.ReceiptID(digest),
		StartedAt:             startedAt,
		FinishedAt:            startedAt.Add(time.Nanosecond),
	}
	output := report.PublishedOutput{Path: "out.csv", FormatID: core.CSVFormat, SHA256: digest, SizeBytes: 1}
	failure := application.APIError{
		Code: "operation_failed", Message: "operation failed", Component: "logical_engine", Details: []report.ErrorDetail{},
	}

	for name, report := range map[string]application.RunReport{
		"succeeded": {Status: core.TerminalStatusSucceeded, Output: &output},
		"failed":    {Status: core.TerminalStatusFailed, FirstFailure: &failure},
		"blocked":   {Status: core.TerminalStatusBlocked, FirstFailure: &failure},
		"cancelled": {Status: core.TerminalStatusCancelled, FirstFailure: &failure},
	} {
		t.Run(name, func(t *testing.T) {
			report.RequestID = base.RequestID
			report.RunID = base.RunID
			report.Validations = base.Validations
			report.Findings = base.Findings
			report.EffectiveConfigDigest = base.EffectiveConfigDigest
			report.CapabilitiesDigest = base.CapabilitiesDigest
			report.ReceiptID = base.ReceiptID
			report.StartedAt = base.StartedAt
			report.FinishedAt = base.FinishedAt
			if err := report.Validate(); err != nil {
				t.Fatalf("RunReport.Validate() error = %v", err)
			}
		})
	}

	invalid := []application.RunReport{
		{Status: core.TerminalStatusSucceeded},
		{Status: core.TerminalStatusSucceeded, Output: &output, FirstFailure: &failure},
		{Status: core.TerminalStatusFailed},
		{Status: core.TerminalStatusBlocked, Output: &output, FirstFailure: &failure},
	}
	for index, report := range invalid {
		report.RequestID = base.RequestID
		report.RunID = base.RunID
		report.Validations = base.Validations
		report.Findings = base.Findings
		report.EffectiveConfigDigest = base.EffectiveConfigDigest
		report.CapabilitiesDigest = base.CapabilitiesDigest
		report.ReceiptID = base.ReceiptID
		report.StartedAt = base.StartedAt
		report.FinishedAt = base.FinishedAt
		if err := report.Validate(); err == nil {
			t.Errorf("invalid terminal case %d unexpectedly validated", index)
		}
	}
}

func TestRunReportRequiresArrayFieldsAndUTCTimestamps(t *testing.T) {
	digest := core.SHA256Digest(strings.Repeat("a", 64))
	valid := application.RunReport{
		RequestID:             "123e4567-e89b-12d3-a456-426614174000",
		RunID:                 "123e4567-e89b-12d3-a456-426614174001",
		Status:                core.TerminalStatusSucceeded,
		Output:                &report.PublishedOutput{Path: "out.csv", FormatID: core.CSVFormat, SHA256: digest, SizeBytes: 1},
		Validations:           []report.ValidationResult{},
		Findings:              []report.Finding{},
		EffectiveConfigDigest: digest,
		CapabilitiesDigest:    digest,
		ReceiptID:             core.ReceiptID(digest),
		StartedAt:             time.Date(2026, time.September, 7, 8, 0, 0, 0, time.UTC),
		FinishedAt:            time.Date(2026, time.September, 7, 8, 0, 0, 1, time.UTC),
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}

	missingArrays := valid
	missingArrays.Validations = nil
	if err := missingArrays.Validate(); err == nil {
		t.Fatal("RunReport accepted nil validations")
	}
	nonUTC := valid
	nonUTC.StartedAt = valid.StartedAt.In(time.FixedZone("+02", 2*60*60))
	if err := nonUTC.Validate(); err == nil {
		t.Fatal("RunReport accepted non-UTC timestamp")
	}
}

func TestRunReportJSONGoldens(t *testing.T) {
	sha256 := func(character string) core.SHA256Digest {
		return core.SHA256Digest(strings.Repeat(character, 64))
	}
	startedAt := time.Date(2026, time.September, 7, 8, 0, 0, 123456789, time.UTC)
	finishedAt := time.Date(2026, time.September, 7, 8, 0, 1, 987654321, time.UTC)
	snapshotHash := sha256("a")
	receiptID := core.ReceiptID(sha256("d"))

	base := func(status core.TerminalStatus) application.RunReport {
		return application.RunReport{
			RequestID:             "123e4567-e89b-12d3-a456-426614174000",
			RunID:                 "123e4567-e89b-12d3-a456-426614174001",
			Status:                status,
			Validations:           []report.ValidationResult{},
			Findings:              []report.Finding{},
			EffectiveConfigDigest: sha256("b"),
			CapabilitiesDigest:    sha256("c"),
			ReceiptID:             receiptID,
			StartedAt:             startedAt,
			FinishedAt:            finishedAt,
		}
	}

	succeeded := base(core.TerminalStatusSucceeded)
	succeeded.Output = &report.PublishedOutput{
		Path: "C:\\output\\succeeded.csv", FormatID: core.CSVFormat, SHA256: sha256("e"), SizeBytes: 42,
	}
	succeeded.Validations = []report.ValidationResult{{Kind: report.ValidationKindFormat, Passed: true, Message: "format reopened"}}
	succeeded.Findings = []report.Finding{{Code: "trimmed", Severity: core.FindingSeverityInfo, Message: "value normalized"}}
	succeeded.ExceptionSummary = &report.ExceptionSummary{Action: core.ExceptionActionReport, Count: 1}

	failed := base(core.TerminalStatusFailed)
	failed.SourceSnapshotHash = &snapshotHash
	failed.Validations = []report.ValidationResult{{Kind: report.ValidationKindSchema, Passed: false, Code: "output_invalid", Message: "reopened output schema mismatch"}}
	failed.FirstFailure = &application.APIError{
		Code: "output_invalid", Message: "reopened output schema mismatch", Component: "writer_engine", Retryable: false,
		Details:    []report.ErrorDetail{{Key: "validation", Value: "schema"}},
		RunContext: &application.RunFailureContext{RunID: failed.RunID, Status: core.TerminalStatusFailed, ReceiptID: &receiptID},
	}

	blocked := base(core.TerminalStatusBlocked)
	blocked.Findings = []report.Finding{{Code: "select_sheet", Severity: core.FindingSeverityWarning, Message: "source sheet requires selection", RequiresDecision: true}}
	blocked.ExceptionSummary = &report.ExceptionSummary{Action: core.ExceptionActionBlock, Count: 1}
	blocked.FirstFailure = &application.APIError{
		Code: "inspection_ambiguous", Message: "source sheet requires selection", Component: "reader_engine", FieldPath: pointerTo("/source/sheet_id"), Retryable: false,
		Details:    []report.ErrorDetail{{Key: "decision", Value: "select_sheet"}},
		RunContext: &application.RunFailureContext{RunID: blocked.RunID, Status: core.TerminalStatusBlocked, ReceiptID: &receiptID},
	}

	cancelled := base(core.TerminalStatusCancelled)
	cancelled.SourceSnapshotHash = &snapshotHash
	cancelled.FirstFailure = &application.APIError{
		Code: "cancelled", Message: "run cancelled", Component: "orchestrator", Retryable: false, Details: []report.ErrorDetail{},
		RunContext: &application.RunFailureContext{RunID: cancelled.RunID, Status: core.TerminalStatusCancelled, ReceiptID: &receiptID},
	}

	tests := []struct {
		name   string
		report application.RunReport
		want   string
	}{
		{
			name:   "succeeded",
			report: succeeded,
			want:   `{"request_id":"123e4567-e89b-12d3-a456-426614174000","run_id":"123e4567-e89b-12d3-a456-426614174001","status":"succeeded","output":{"path":"C:\\output\\succeeded.csv","format_id":"csv","sha256":"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","size_bytes":42},"validations":[{"kind":"format","passed":true,"message":"format reopened"}],"findings":[{"code":"trimmed","severity":"info","message":"value normalized","requires_decision":false}],"exception_summary":{"action":"report","count":1},"effective_config_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","capabilities_digest":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","receipt_id":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","started_at":"2026-09-07T08:00:00.123456789Z","finished_at":"2026-09-07T08:00:01.987654321Z"}`,
		},
		{
			name:   "failed",
			report: failed,
			want:   `{"request_id":"123e4567-e89b-12d3-a456-426614174000","run_id":"123e4567-e89b-12d3-a456-426614174001","status":"failed","validations":[{"kind":"schema","passed":false,"code":"output_invalid","message":"reopened output schema mismatch"}],"findings":[],"first_failure":{"code":"output_invalid","message":"reopened output schema mismatch","component":"writer_engine","retryable":false,"details":[{"key":"validation","value":"schema"}],"run_context":{"run_id":"123e4567-e89b-12d3-a456-426614174001","status":"failed","receipt_id":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"}},"source_snapshot_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","effective_config_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","capabilities_digest":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","receipt_id":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","started_at":"2026-09-07T08:00:00.123456789Z","finished_at":"2026-09-07T08:00:01.987654321Z"}`,
		},
		{
			name:   "blocked",
			report: blocked,
			want:   `{"request_id":"123e4567-e89b-12d3-a456-426614174000","run_id":"123e4567-e89b-12d3-a456-426614174001","status":"blocked","validations":[],"findings":[{"code":"select_sheet","severity":"warning","message":"source sheet requires selection","requires_decision":true}],"exception_summary":{"action":"block","count":1},"first_failure":{"code":"inspection_ambiguous","message":"source sheet requires selection","component":"reader_engine","field_path":"/source/sheet_id","retryable":false,"details":[{"key":"decision","value":"select_sheet"}],"run_context":{"run_id":"123e4567-e89b-12d3-a456-426614174001","status":"blocked","receipt_id":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"}},"effective_config_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","capabilities_digest":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","receipt_id":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","started_at":"2026-09-07T08:00:00.123456789Z","finished_at":"2026-09-07T08:00:01.987654321Z"}`,
		},
		{
			name:   "cancelled",
			report: cancelled,
			want:   `{"request_id":"123e4567-e89b-12d3-a456-426614174000","run_id":"123e4567-e89b-12d3-a456-426614174001","status":"cancelled","validations":[],"findings":[],"first_failure":{"code":"cancelled","message":"run cancelled","component":"orchestrator","retryable":false,"details":[],"run_context":{"run_id":"123e4567-e89b-12d3-a456-426614174001","status":"cancelled","receipt_id":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"}},"source_snapshot_hash":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","effective_config_digest":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","capabilities_digest":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","receipt_id":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","started_at":"2026-09-07T08:00:00.123456789Z","finished_at":"2026-09-07T08:00:01.987654321Z"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.report.Validate(); err != nil {
				t.Fatalf("RunReport.Validate() error = %v", err)
			}
			encoded, err := json.Marshal(test.report)
			if err != nil {
				t.Fatal(err)
			}
			if got := string(encoded); got != test.want {
				t.Fatalf("RunReport JSON = %s, want %s", got, test.want)
			}
		})
	}
}

func pointerTo(value string) *string {
	return &value
}
