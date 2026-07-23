package contract

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestStartedAndFinalResultContracts(t *testing.T) {
	startedAt := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	started := StartedResult{
		SchemaVersion: CurrentSchemaVersion,
		RunID:         "run-1",
		WorkflowID:    "workflow-1",
		Status:        RunStatusStarted,
		StartedAt:     startedAt,
	}
	if err := ValidateStartedResult(started); err != nil {
		t.Fatal(err)
	}

	final := FinalResult{
		SchemaVersion: CurrentSchemaVersion,
		RunID:         "run-1",
		WorkflowID:    "workflow-1",
		Status:        RunStatusSucceeded,
		StartedAt:     startedAt,
		FinishedAt:    startedAt.Add(time.Second),
		Outputs: []ProducedOutput{{
			ID: "output", Format: FormatCSV, Location: "output.csv", Rows: 2,
		}},
		QueryAnswers: []QueryAnswer{{
			ID: "count", Shape: QueryShapeScalar, Value: "2",
		}},
		ResolvedSettings: map[string]any{"null_display": "-", "collision": "block"},
		Validations:      []ValidationResult{{ID: "rows", Passed: true}},
		Warnings:         []Diagnostic{{Code: "cached-value", Message: "cached value used"}},
		Exceptions:       []ExceptionRecord{{Reference: "exception-1", SourceID: "source", RowID: "2", Code: "mixed-type"}},
	}
	if err := ValidateFinalResult(final); err != nil {
		t.Fatal(err)
	}

	encoded, err := json.Marshal(final)
	if err != nil {
		t.Fatal(err)
	}
	var decoded FinalResult
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(final, decoded) {
		t.Fatalf("round trip changed final result:\nwant %#v\ngot  %#v", final, decoded)
	}
}

func TestFailedResultRequiresDetailedError(t *testing.T) {
	now := time.Now().UTC()
	result := FinalResult{
		SchemaVersion:    CurrentSchemaVersion,
		RunID:            "run-1",
		WorkflowID:       "workflow-1",
		Status:           RunStatusFailed,
		StartedAt:        now,
		FinishedAt:       now,
		ResolvedSettings: map[string]any{},
	}
	if err := ValidateFinalResult(result); err == nil {
		t.Fatal("expected missing detailed error to fail validation")
	}
	result.Errors = []Diagnostic{{Code: "source-read", Path: "$.sources[0]", Message: "source is locked"}}
	if err := ValidateFinalResult(result); err != nil {
		t.Fatal(err)
	}
}
