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
	"data-toolkit/internal/testsupport"
)

const approvedMVP2Source = `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\all_final\ext_members_from_12072026_to_14072026_no_full_phone_2026-07-14-12-35-21.json`

func TestMVP2ApprovedSource(t *testing.T) {
	sourceBefore, err := testsupport.SnapshotFile(approvedMVP2Source)
	if err != nil {
		t.Fatal(err)
	}
	defer testsupport.AssertFileUnchanged(t, sourceBefore)
	if sourceBefore.Size != 13_559_851 {
		t.Fatalf("approved source size changed: %d", sourceBefore.Size)
	}

	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	fixtureRoot := filepath.Join(repositoryRoot, "testdata", "fixtures", "mvp2-json-to-csv")
	workflowFile, err := os.Open(filepath.Join(fixtureRoot, "workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow, err := contract.DecodeWorkflow(workflowFile, contract.DocumentYAML)
	workflowFile.Close()
	if err != nil {
		t.Fatal(err)
	}
	workflow.Sources[0].Location = approvedMVP2Source
	workflow.Outputs[0].Validations[0].Expected = "4665"
	generated := testsupport.NewGeneratedDir(t, repositoryRoot, "mvp2-approved")
	workflow.Outputs[0].Location = filepath.Join(generated, "members.csv")

	registry := adapter.NewRegistry()
	if err := registry.Register(adapter.NewJSONAdapter()); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(adapter.NewCSVAdapter()); err != nil {
		t.Fatal(err)
	}
	defaults := config.BuiltInDefaults(t.TempDir())
	defaults.MaximumMemoryBytes = 128 * 1024 * 1024
	defaults.FailedRunRetention = 0
	service := Service{
		Adapters: registry, Defaults: defaults,
		UserConfig: config.Config{SchemaVersion: config.CurrentSchemaVersion},
		RunID:      func() string { return "mvp2-approved-source" },
		Backend:    &LocalPipelineBackend{Adapters: registry},
	}
	final, err := service.Run(context.Background(), contract.RunRequest{Workflow: workflow})
	if err != nil {
		t.Fatal(err)
	}
	if len(final.Outputs) != 1 || final.Outputs[0].Rows != 4665 {
		t.Fatalf("MVP 2 result = %#v", final)
	}
	file, err := os.Open(final.Outputs[0].Location)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}
	wantHeader := []string{"שם פרטי", "שם מלא", "entityID", "TZ", "ישוב", "מייל"}
	for index := range wantHeader {
		if header[index] != wantHeader[index] {
			t.Fatalf("header[%d] = %q, want %q", index, header[index], wantHeader[index])
		}
	}
	first, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}
	if first[0] != "" || first[1] != "" || first[2] != "375162" ||
		first[3] != "55499644" || first[4] != "יהוד-מונוסון " {
		t.Fatalf("first projected row differs: %#v", first)
	}
}
