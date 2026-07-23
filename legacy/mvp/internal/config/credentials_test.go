package config

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"data-toolkit/internal/contract"
)

func TestCredentialFieldsAreRejectedFromOrdinaryDocuments(t *testing.T) {
	configuration := "{\"schema_version\":\"v1\",\"access_token\":\"secret\"}"
	if _, err := Decode(strings.NewReader(configuration), contract.DocumentJSON); err == nil {
		t.Fatal("configuration accepted a credential field")
	}

	workflowBytes, err := os.ReadFile("../contract/testdata/workflow.json")
	if err != nil {
		t.Fatal(err)
	}
	workflow := strings.Replace(string(workflowBytes), "\"id\": \"city-filter\",", "\"id\": \"city-filter\",\n  \"client_secret\": \"secret\",", 1)
	if _, err := contract.DecodeWorkflow(strings.NewReader(workflow), contract.DocumentJSON); err == nil {
		t.Fatal("workflow accepted a credential field")
	}
}

func TestRedactorRemovesSecretsFromLogsErrorsAndReports(t *testing.T) {
	const secret = "s3cr3t-token"
	redactor := NewRedactor(secret)
	if got := redactor.RedactString("refresh failed for " + secret); strings.Contains(got, secret) {
		t.Fatalf("log text still contains secret: %q", got)
	}
	if got := redactor.RedactError(errors.New("oauth token=" + secret)); strings.Contains(got.Error(), secret) {
		t.Fatalf("error still contains secret: %q", got)
	}

	now := time.Now().UTC()
	result := contract.FinalResult{
		SchemaVersion: contract.CurrentSchemaVersion,
		RunID:         "run",
		WorkflowID:    "workflow",
		Status:        contract.RunStatusFailed,
		StartedAt:     now,
		FinishedAt:    now,
		ResolvedSettings: map[string]any{
			"nested": map[string]any{"accidental": secret},
		},
		Warnings: []contract.Diagnostic{{Code: "oauth", Message: "token " + secret}},
		Exceptions: []contract.ExceptionRecord{{
			Reference: "exception", Code: "oauth", Details: map[string]any{"token": secret},
		}},
		Errors: []contract.Diagnostic{{
			Code: "oauth", Message: "provider returned " + secret, Details: map[string]any{"response": []any{secret}},
		}},
	}
	sanitized := redactor.RedactFinalResult(result)
	encoded, err := json.Marshal(sanitized)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), secret) {
		t.Fatalf("report still contains secret: %s", encoded)
	}
	if !strings.Contains(string(encoded), redactedValue) {
		t.Fatalf("report does not contain redaction marker: %s", encoded)
	}
}
