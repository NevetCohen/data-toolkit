package orchestrator

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"data-toolkit/internal/adapter"
	"data-toolkit/internal/config"
	"data-toolkit/internal/contract"
	"data-toolkit/internal/testsupport"
)

func TestMVP2RepresentativeFixture(t *testing.T) {
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	fixtureRoot := filepath.Join(repositoryRoot, "testdata", "fixtures", "mvp2-json-to-csv")
	inputPath := filepath.Join(fixtureRoot, "input.json")
	before, err := testsupport.SnapshotFile(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer testsupport.AssertFileUnchanged(t, before)

	workflowFile, err := os.Open(filepath.Join(fixtureRoot, "workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow, err := contract.DecodeWorkflow(workflowFile, contract.DocumentYAML)
	workflowFile.Close()
	if err != nil {
		t.Fatal(err)
	}
	workflow.Sources[0].Location = inputPath
	generated := testsupport.NewGeneratedDir(t, repositoryRoot, "mvp2")
	workflow.Outputs[0].Location = filepath.Join(generated, "mvp2.csv")

	registry := adapter.NewRegistry()
	if err := registry.Register(adapter.NewJSONAdapter()); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(adapter.NewCSVAdapter()); err != nil {
		t.Fatal(err)
	}
	defaults := config.BuiltInDefaults(t.TempDir())
	defaults.FailedRunRetention = 0
	service := Service{
		Adapters: registry, Defaults: defaults,
		UserConfig: config.Config{SchemaVersion: config.CurrentSchemaVersion},
		RunID:      func() string { return "mvp2-fixture" },
		Backend:    &LocalPipelineBackend{Adapters: registry},
	}
	final, err := service.Run(context.Background(), contract.RunRequest{Workflow: workflow})
	if err != nil {
		t.Fatal(err)
	}
	if len(final.Outputs) != 1 || final.Outputs[0].Rows != 3 {
		t.Fatalf("final output = %#v", final)
	}
	got := readCSVRecords(t, final.Outputs[0].Location)
	want := readCSVRecords(t, filepath.Join(fixtureRoot, "expected.csv"))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MVP 2 output differs:\nwant %#v\ngot  %#v", want, got)
	}
}

func readCSVRecords(t *testing.T, path string) [][]string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	return records
}
