package application_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"data-toolkit/internal/application"
	"data-toolkit/internal/core"
	"data-toolkit/internal/core/locator"
	"data-toolkit/internal/report"
)

const testRunID = core.RunID("123e4567-e89b-12d3-a456-426614174000")

func TestAPIErrorHasExactlyEightFields(t *testing.T) {
	typeOfError := reflect.TypeFor[application.APIError]()
	want := []struct {
		name string
		json string
	}{
		{"Code", "code"}, {"Message", "message"}, {"Component", "component"},
		{"FieldPath", "field_path,omitempty"}, {"Locator", "locator,omitempty"},
		{"Retryable", "retryable"}, {"Details", "details"}, {"RunContext", "run_context,omitempty"},
	}
	if typeOfError.NumField() != len(want) {
		t.Fatalf("APIError field count = %d, want %d", typeOfError.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfError.Field(index)
		if field.Name != expected.name || field.Tag.Get("json") != expected.json {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, expected.json)
		}
	}
}

func TestAPIErrorSerializesOneLocatorVariant(t *testing.T) {
	row := uint64(7)
	errorValue := application.APIError{
		Code: "source_not_found", Message: "source was not found", Component: "reader_engine",
		Locator:   application.SourceErrorLocator(locator.SourceLocator{SourceID: core.SourceID("source-1"), RowOrdinal: &row}),
		Retryable: false, Details: []report.ErrorDetail{{Key: "source_id", Value: "source-1"}},
	}
	encoded, err := json.Marshal(errorValue)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"code":"source_not_found","message":"source was not found","component":"reader_engine","locator":{"source_id":"source-1","row_ordinal":7},"retryable":false,"details":[{"key":"source_id","value":"source-1"}]}`
	if string(encoded) != want {
		t.Fatalf("APIError JSON = %s, want %s", encoded, want)
	}

	output := application.APIError{
		Code: "output_invalid", Message: "output is invalid", Component: "writer_engine",
		Locator: application.OutputErrorLocator(locator.OutputLocator{OutputPath: `C:\out.csv`}),
		Details: []report.ErrorDetail{},
	}
	if err := output.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAPIErrorRejectsAbsentOrAmbiguousLocator(t *testing.T) {
	base := application.APIError{Code: "internal_error", Message: "failed", Component: "infrastructure", Details: []report.ErrorDetail{}}
	if err := (application.ErrorLocator{}).Validate(); err == nil {
		t.Fatal("empty locator accepted")
	}
	source := locator.SourceLocator{SourceID: "source-1"}
	output := locator.OutputLocator{OutputPath: "out.csv"}
	if err := (application.ErrorLocator{Source: &source, Output: &output}).Validate(); err == nil {
		t.Fatal("ambiguous locator accepted")
	}
	base.Locator = &application.ErrorLocator{Source: &source, Output: &output}
	if err := base.Validate(); err == nil {
		t.Fatal("APIError accepted ambiguous locator")
	}
}

func TestRunFailureContextReceiptIDIsConditional(t *testing.T) {
	withoutReceipt := application.APIError{
		Code: "source_locked", Message: "source is locked", Component: "reader_engine",
		Retryable: true, Details: []report.ErrorDetail{},
		RunContext: &application.RunFailureContext{RunID: testRunID, Status: core.TerminalStatusFailed},
	}
	encoded, err := json.Marshal(withoutReceipt)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "receipt_id") {
		t.Fatalf("pre-append error contains receipt_id: %s", encoded)
	}

	receipt := core.ReceiptID(strings.Repeat("a", 64))
	withReceipt := withoutReceipt
	withReceipt.RunContext = &application.RunFailureContext{RunID: testRunID, Status: core.TerminalStatusFailed, ReceiptID: &receipt}
	encoded, err = json.Marshal(withReceipt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"receipt_id":"`+string(receipt)+`"`) {
		t.Fatalf("post-append error omitted receipt_id: %s", encoded)
	}
}

func TestAPIErrorRejectsUnknownJSONField(t *testing.T) {
	var value application.APIError
	if err := json.Unmarshal([]byte(`{"code":"internal_error","message":"failed","component":"infrastructure","retryable":false,"details":[],"extra":true}`), &value); err == nil {
		t.Fatal("unknown APIError field accepted")
	}
}
