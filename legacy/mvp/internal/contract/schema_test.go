package contract

import (
	"errors"
	"strings"
	"testing"
)

func TestSchemaErrorsCarryDocumentPaths(t *testing.T) {
	base := `{
  "schema_version":"v1",
  "id":"invalid",
  "sources":[{"id":"source","format":"csv","location":"input.csv"}],
  "operations":[{"id":"filter","kind":"filter","inputs":["source"],"condition":{"operator":"compare","comparison":{"left":{"source_id":"source","column":{"header":"A"}},"operator":"eq","right":{"literal":{"kind":"string","value":"x"}}}}}],
  "outputs":[{"id":"out","input":"filter","format":"csv","location":"out.csv","projection":[{"column":{"header":"A"},"output_as":"A"}]}],
  "interface":{"name":"cli","approval_before_run":false},
  "exception":{"action":"block"}
}`

	tests := []struct {
		name       string
		document   string
		wantedPath string
	}{
		{name: "version", document: strings.Replace(base, `"v1"`, `"v9"`, 1), wantedPath: "$.schema_version"},
		{name: "operation", document: strings.Replace(base, `"filter","inputs"`, `"mystery","inputs"`, 1), wantedPath: "$.operations[0].kind"},
		{name: "enum", document: strings.Replace(base, `"action":"block"`, `"action":"explode"`, 1), wantedPath: "$.exception.action"},
		{name: "field", document: strings.Replace(base, `"inputs":["source"]`, `"inputs":["source"],"surprise":true`, 1), wantedPath: "$.operations[0]"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := DecodeWorkflow(strings.NewReader(test.document), DocumentJSON)
			if err == nil {
				t.Fatal("expected schema error")
			}
			var findings ValidationErrors
			if !errors.As(err, &findings) {
				t.Fatalf("error type = %T, want ValidationErrors: %v", err, err)
			}
			found := false
			for _, finding := range findings {
				if finding.Path == test.wantedPath {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("paths = %#v, want %q", findings, test.wantedPath)
			}
		})
	}
}

func TestSemanticValidationReportsMissingOperationParameter(t *testing.T) {
	workflow := validWorkflowForTest()
	workflow.Operations[0].Condition = nil

	err := ValidateWorkflow(workflow)
	var findings ValidationErrors
	if !errors.As(err, &findings) {
		t.Fatalf("error type = %T, want ValidationErrors: %v", err, err)
	}
	if findings[0].Path != "$.operations[0].condition" {
		t.Fatalf("path = %q, want operation condition path", findings[0].Path)
	}
}

func validWorkflowForTest() Workflow {
	return Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            "valid",
		Sources:       []Source{{ID: "source", Format: FormatCSV, Location: "input.csv"}},
		Operations: []Operation{{
			ID:     "filter",
			Kind:   OperationFilter,
			Inputs: []string{"source"},
			Condition: &Expression{
				Operator: ExpressionCompare,
				Comparison: &Comparison{
					Left:     Operand{SourceID: "source", Column: &ColumnRef{Header: "A"}},
					Operator: CompareEqual,
					Right:    Operand{Literal: &ScalarLiteral{Kind: LiteralString, Value: "x"}},
				},
			},
		}},
		Outputs: []Output{{
			ID: "out", Input: "filter", Format: FormatCSV, Location: "out.csv",
			Projection: []ProjectionField{{Column: ColumnRef{Header: "A"}, OutputAs: "A"}},
		}},
		Interface: InterfaceMetadata{Name: "cli"},
		Exception: ExceptionPolicy{Action: ExceptionBlock},
	}
}
