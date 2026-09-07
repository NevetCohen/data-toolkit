package application_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"data-toolkit/internal/application"
	"data-toolkit/internal/core"
	"data-toolkit/internal/report"
)

func TestRunReportRetainsFindingsAndExceptionsOutsideCompactReceipt(t *testing.T) {
	digest := core.SHA256Digest(strings.Repeat("a", 64))
	startedAt := time.Date(2026, time.September, 7, 10, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(time.Nanosecond)
	output := report.PublishedOutput{
		Path:      "out.csv",
		FormatID:  core.CSVFormat,
		SHA256:    digest,
		SizeBytes: 1,
	}
	exceptionReportPath := "C:\\reports\\contract-exceptions.json"
	runReport := application.RunReport{
		RequestID: "123e4567-e89b-12d3-a456-426614174000",
		RunID:     "123e4567-e89b-12d3-a456-426614174001",
		Status:    core.TerminalStatusSucceeded,
		Output:    &output,
		Validations: []report.ValidationResult{
			{Kind: report.ValidationKindFormat, Passed: true, Message: "format reopened"},
		},
		Findings: []report.Finding{
			{
				Code:     "contract_finding_must_remain_in_report",
				Severity: core.FindingSeverityWarning,
				Message:  "contract finding detail must not enter receipt",
			},
		},
		ExceptionSummary:      &report.ExceptionSummary{Action: core.ExceptionActionReport, Count: 1, ReportPath: &exceptionReportPath},
		EffectiveConfigDigest: digest,
		CapabilitiesDigest:    digest,
		ReceiptID:             core.ReceiptID(digest),
		StartedAt:             startedAt,
		FinishedAt:            finishedAt,
	}
	receipt := application.RunReceipt{
		ReceiptID:                core.ReceiptID(digest),
		RequestID:                runReport.RequestID,
		RunID:                    runReport.RunID,
		NormalizedWorkflowDigest: digest,
		BinaryVersion:            "v1.0.0",
		CapabilitiesDigest:       digest,
		EffectiveConfigDigest:    digest,
		PublishedOutput:          &output,
		Status:                   core.TerminalStatusSucceeded,
		StartedAt:                startedAt,
		FinishedAt:               finishedAt,
	}

	if err := runReport.Validate(); err != nil {
		t.Fatalf("RunReport.Validate() error = %v", err)
	}
	if err := receipt.Validate(); err != nil {
		t.Fatalf("RunReceipt.Validate() error = %v", err)
	}

	reportJSON, err := json.Marshal(runReport)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	receiptJSON, err := json.Marshal(receipt)
	if err != nil {
		t.Fatalf("marshal receipt: %v", err)
	}

	var reportFields map[string]json.RawMessage
	if err := json.Unmarshal(reportJSON, &reportFields); err != nil {
		t.Fatalf("decode report JSON: %v", err)
	}
	for _, field := range []string{"findings", "exception_summary"} {
		if _, ok := reportFields[field]; !ok {
			t.Errorf("report omits %q: %s", field, reportJSON)
		}
	}

	var receiptFields map[string]json.RawMessage
	if err := json.Unmarshal(receiptJSON, &receiptFields); err != nil {
		t.Fatalf("decode receipt JSON: %v", err)
	}
	for _, field := range []string{"findings", "exception_summary"} {
		if _, ok := receiptFields[field]; ok {
			t.Errorf("compact receipt contains %q: %s", field, receiptJSON)
		}
	}
	for _, detail := range []string{"contract_finding_must_remain_in_report", exceptionReportPath} {
		if strings.Contains(string(receiptJSON), detail) {
			t.Errorf("compact receipt contains report-only detail %q: %s", detail, receiptJSON)
		}
	}
}
