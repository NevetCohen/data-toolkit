package orchestrator

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"data-toolkit/internal/adapter"
	"data-toolkit/internal/config"
	"data-toolkit/internal/contract"
)

func TestJSONProjectionToCSVRunsThroughOrchestrator(t *testing.T) {
	registry := adapter.NewRegistry()
	if err := registry.Register(adapter.NewJSONAdapter()); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(adapter.NewCSVAdapter()); err != nil {
		t.Fatal(err)
	}
	defaults := config.BuiltInDefaults(t.TempDir())
	defaults.FailedRunRetention = 0
	outputPath := filepath.Join(t.TempDir(), "projected.csv")
	workflow := contract.Workflow{
		SchemaVersion: contract.CurrentSchemaVersion,
		ID:            "json-to-csv",
		Sources: []contract.Source{{
			ID: "json", Format: contract.FormatJSON,
			Location: filepath.Join("..", "adapter", "testdata", "json", "rows.json"),
			Options:  contract.SourceOptions{JSONRoot: "/Data"},
		}},
		Mappings: []contract.FieldMapping{
			{SourceID: "json", Column: contract.ColumnRef{Header: "name"}, JSONPath: "/name"},
			{SourceID: "json", Column: contract.ColumnRef{Header: "id"}, JSONPath: "/id"},
			{SourceID: "json", Column: contract.ColumnRef{Header: "city"}, JSONPath: "/nested/city"},
		},
		Outputs: []contract.Output{{
			ID: "csv", Input: "json", Format: contract.FormatCSV, Location: outputPath,
			Projection: []contract.ProjectionField{
				{Column: contract.ColumnRef{Header: "city"}, OutputAs: "ישוב"},
				{Column: contract.ColumnRef{Header: "name"}, OutputAs: "שם"},
				{Column: contract.ColumnRef{Header: "id"}, OutputAs: "entityID"},
			},
			Validations: []contract.Validation{{ID: "rows", Kind: contract.ValidationRowCount, Expected: "3"}},
		}},
		Interface: contract.InterfaceMetadata{Name: "test"},
		Exception: contract.ExceptionPolicy{Action: contract.ExceptionBlock},
	}
	service := Service{
		Adapters: registry, Defaults: defaults,
		UserConfig: config.Config{SchemaVersion: config.CurrentSchemaVersion},
		RunID:      func() string { return "json-to-csv-run" },
		Backend:    &LocalPipelineBackend{Adapters: registry},
	}
	final, err := service.Run(context.Background(), contract.RunRequest{Workflow: workflow})
	if err != nil {
		t.Fatal(err)
	}
	if final.Status != contract.RunStatusSucceeded || len(final.Outputs) != 1 || final.Outputs[0].Rows != 3 {
		t.Fatalf("final result = %#v", final)
	}
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if rows[0][0] != "ישוב" || rows[1][0] != "בת ים" ||
		rows[1][1] != "נועה 🧭" || rows[1][2] != "9007199254740993" ||
		rows[3][0] != "-" {
		t.Fatalf("CSV rows = %#v", rows)
	}
}
