package config

import (
	"testing"
	"time"

	"data-toolkit/internal/contract"
)

func TestReportSnapshotCapturesEveryPrecedenceLayer(t *testing.T) {
	userDelimiter := ","
	user := Config{SchemaVersion: CurrentSchemaVersion, Output: OutputConfig{TXTDelimiter: &userDelimiter}}
	tests := []struct {
		name     string
		user     Config
		workflow contract.WorkflowOverrides
		output   *contract.OutputOverrides
		want     string
	}{
		{name: "built-in", user: Config{SchemaVersion: CurrentSchemaVersion}, want: "\t"},
		{name: "user", user: user, want: ","},
		{name: "workflow", user: user, workflow: contract.WorkflowOverrides{TXTDelimiter: "|"}, want: "|"},
		{name: "output", user: user, workflow: contract.WorkflowOverrides{TXTDelimiter: "|"}, output: &contract.OutputOverrides{Delimiter: ";"}, want: ";"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolved, err := Resolve(BuiltInDefaults("temporary"), test.user, test.workflow, test.output)
			if err != nil {
				t.Fatal(err)
			}
			output := resolved.ReportSettings()["output"].(map[string]any)
			if got := output["txt_delimiter"]; got != test.want {
				t.Fatalf("reported delimiter = %q, want %q", got, test.want)
			}
		})
	}
}

func TestResolvedSnapshotFitsFinalRunReport(t *testing.T) {
	resolved, err := Resolve(
		BuiltInDefaults("temporary"),
		Config{SchemaVersion: CurrentSchemaVersion},
		contract.WorkflowOverrides{},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	result := contract.FinalResult{
		SchemaVersion:    contract.CurrentSchemaVersion,
		RunID:            "run",
		WorkflowID:       "workflow",
		Status:           contract.RunStatusSucceeded,
		StartedAt:        now,
		FinishedAt:       now,
		ResolvedSettings: resolved.ReportSettings(),
	}
	if err := contract.ValidateFinalResult(result); err != nil {
		t.Fatal(err)
	}
}
