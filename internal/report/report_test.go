package report

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/locator"
)

func TestFindingHasExactFields(t *testing.T) {
	typeOfFinding := reflect.TypeFor[Finding]()
	want := []struct {
		name string
		json string
	}{
		{"Code", "code"},
		{"Severity", "severity"},
		{"Message", "message"},
		{"Locator", "locator,omitempty"},
		{"RequiresDecision", "requires_decision"},
	}
	if typeOfFinding.NumField() != len(want) {
		t.Fatalf("Finding field count = %d, want %d", typeOfFinding.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfFinding.Field(index)
		if field.Name != expected.name || field.Tag.Get("json") != expected.json {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, expected.json)
		}
	}
}

func TestValidationResultUsesOnlyFixedKinds(t *testing.T) {
	for _, kind := range []ValidationKind{
		ValidationKindFormat,
		ValidationKindSchema,
		ValidationKindExactColumns,
		ValidationKindAllRowsMatch,
	} {
		if err := (ValidationResult{Kind: kind, Passed: true, Message: "passed"}).Validate(); err != nil {
			t.Errorf("ValidationResult kind %q rejected: %v", kind, err)
		}
	}
	if err := (ValidationResult{Kind: "unknown", Message: "invalid"}).Validate(); err == nil {
		t.Fatal("ValidationResult accepted unknown kind")
	}
}

func TestExceptionSummaryHasExactFields(t *testing.T) {
	typeOfSummary := reflect.TypeFor[ExceptionSummary]()
	want := []struct {
		name string
		json string
	}{
		{"Action", "action"},
		{"Count", "count"},
		{"FirstLocator", "first_locator,omitempty"},
		{"ReportPath", "report_path,omitempty"},
	}
	if typeOfSummary.NumField() != len(want) {
		t.Fatalf("ExceptionSummary field count = %d, want %d", typeOfSummary.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfSummary.Field(index)
		if field.Name != expected.name || field.Tag.Get("json") != expected.json {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, expected.json)
		}
	}
}

func TestExceptionSummaryJSONOmitsOptionalFields(t *testing.T) {
	summary := ExceptionSummary{Action: core.ExceptionActionReport, Count: 2}
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(encoded), `{"action":"report","count":2}`; got != want {
		t.Fatalf("exception summary JSON = %s, want %s", got, want)
	}
}

func TestExceptionSummaryJSONIncludesLocatorAndReportPath(t *testing.T) {
	row := uint64(4)
	position := "B5"
	reportPath := `C:\\reports\\exceptions.json`
	summary := ExceptionSummary{
		Action:       core.ExceptionActionReport,
		Count:        2,
		FirstLocator: &locator.SourceLocator{SourceID: core.SourceID("source-1"), RowOrdinal: &row, PhysicalPosition: &position},
		ReportPath:   &reportPath,
	}
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"action":"report","count":2,"first_locator":{"source_id":"source-1","row_ordinal":4,"physical_position":"B5"},"report_path":"C:\\\\reports\\\\exceptions.json"}`
	if got := string(encoded); got != want {
		t.Fatalf("exception summary JSON = %s, want %s", got, want)
	}
}

func TestExceptionSummaryValidatesAction(t *testing.T) {
	if err := (ExceptionSummary{Action: core.ExceptionActionReport}).Validate(); err != nil {
		t.Fatalf("valid exception action rejected: %v", err)
	}
	if err := (ExceptionSummary{Action: "unknown"}).Validate(); err == nil {
		t.Fatal("unknown exception action accepted")
	}
}

func TestPublishedOutputHasExactFields(t *testing.T) {
	typeOfOutput := reflect.TypeFor[PublishedOutput]()
	want := []struct {
		name string
		json string
	}{
		{"Path", "path"},
		{"FormatID", "format_id"},
		{"SHA256", "sha256"},
		{"SizeBytes", "size_bytes"},
	}
	if typeOfOutput.NumField() != len(want) {
		t.Fatalf("PublishedOutput field count = %d, want %d", typeOfOutput.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfOutput.Field(index)
		if field.Name != expected.name || field.Tag.Get("json") != expected.json {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, expected.json)
		}
	}
}

func TestPublishedOutputValidatesAndMarshals(t *testing.T) {
	output := PublishedOutput{
		Path:      `C:\\out\\result.csv`,
		FormatID:  core.CSVFormat,
		SHA256:    core.SHA256Digest(strings.Repeat("a", 64)),
		SizeBytes: 42,
	}
	if err := output.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"path":"C:\\\\out\\\\result.csv","format_id":"csv","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","size_bytes":42}`
	if string(encoded) != want {
		t.Fatalf("PublishedOutput JSON = %s, want %s", encoded, want)
	}
}

func TestPublishedOutputRejectsInvalidFields(t *testing.T) {
	valid := PublishedOutput{Path: "out.csv", FormatID: core.CSVFormat, SHA256: core.SHA256Digest(strings.Repeat("a", 64))}
	cases := []PublishedOutput{
		{FormatID: valid.FormatID, SHA256: valid.SHA256},
		{Path: valid.Path, FormatID: "txt", SHA256: valid.SHA256},
		{Path: valid.Path, FormatID: valid.FormatID, SHA256: "short"},
		{Path: valid.Path, FormatID: valid.FormatID, SHA256: valid.SHA256, SizeBytes: -1},
	}
	for index, candidate := range cases {
		if err := candidate.Validate(); err == nil {
			t.Errorf("case %d unexpectedly validated", index)
		}
	}
}

func TestValidationResultOutputLocatorOnlyOnFailure(t *testing.T) {
	row := uint64(2)
	result := ValidationResult{
		Kind:    ValidationKindAllRowsMatch,
		Passed:  false,
		Code:    "output_invalid",
		Message: "row does not match",
		Locator: &locator.OutputLocator{OutputPath: `C:\out\result.csv`, RowOrdinal: &row},
	}
	if err := result.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"kind":"all_rows_match","passed":false,"code":"output_invalid","message":"row does not match","locator":{"output_path":"C:\\out\\result.csv","row_ordinal":2}}`
	if string(encoded) != want {
		t.Fatalf("validation result JSON = %s, want %s", encoded, want)
	}
	if err := (ValidationResult{Kind: ValidationKindFormat, Passed: true, Message: "passed", Locator: result.Locator}).Validate(); err == nil {
		t.Fatal("successful validation accepted an output locator")
	}
}

func TestFindingJSONOmitsAbsentLocator(t *testing.T) {
	finding := Finding{
		Code:             "duplicate_header",
		Severity:         core.FindingSeverityWarning,
		Message:          "duplicate source header",
		RequiresDecision: true,
	}

	encoded, err := json.Marshal(finding)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(encoded), `{"code":"duplicate_header","severity":"warning","message":"duplicate source header","requires_decision":true}`; got != want {
		t.Fatalf("finding JSON = %s, want %s", got, want)
	}
}

func TestFindingJSONIncludesLocator(t *testing.T) {
	row := uint64(3)
	position := "B4"
	finding := Finding{
		Code:     "invalid_value",
		Severity: core.FindingSeverityError,
		Message:  "value is invalid",
		Locator: &locator.SourceLocator{
			SourceID:         core.SourceID("source-1"),
			RowOrdinal:       &row,
			PhysicalPosition: &position,
		},
	}

	encoded, err := json.Marshal(finding)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(encoded), `{"code":"invalid_value","severity":"error","message":"value is invalid","locator":{"source_id":"source-1","row_ordinal":3,"physical_position":"B4"},"requires_decision":false}`; got != want {
		t.Fatalf("finding JSON = %s, want %s", got, want)
	}
}

func TestErrorDetailHasExactFields(t *testing.T) {
	typeOfDetail := reflect.TypeFor[ErrorDetail]()
	want := []struct {
		name string
		json string
	}{
		{"Key", "key"},
		{"Value", "value"},
	}
	if typeOfDetail.NumField() != len(want) {
		t.Fatalf("ErrorDetail field count = %d, want %d", typeOfDetail.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfDetail.Field(index)
		if field.Name != expected.name || field.Tag.Get("json") != expected.json {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, expected.json)
		}
	}
}

func TestNewErrorDetailsSortsAndRedactsSensitiveValues(t *testing.T) {
	details := NewErrorDetails(map[string]string{
		"operation_reason": "invalid_parameters",
		"api_token":        "do-not-leak",
		"credentials":      "also-do-not-leak",
	})
	want := []ErrorDetail{
		{Key: "api_token", Value: "[redacted]"},
		{Key: "credentials", Value: "[redacted]"},
		{Key: "operation_reason", Value: "invalid_parameters"},
	}
	if !reflect.DeepEqual(details, want) {
		t.Fatalf("NewErrorDetails() = %#v, want %#v", details, want)
	}
}
