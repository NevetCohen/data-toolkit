package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"data-toolkit/internal/adapter"
	"data-toolkit/internal/config"
	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

func TestPreflightValidatesBeforeMappingAndRequiresAllMappings(t *testing.T) {
	fake := &planningAdapter{format: contract.FormatCSV}
	service := testService(t, fake)
	invalid := testWorkflow(t)
	invalid.SchemaVersion = "v9"
	if _, err := service.Preflight(context.Background(), contract.RunRequest{Workflow: invalid}); err == nil {
		t.Fatal("expected invalid workflow")
	}
	if fake.mapCalls != 0 {
		t.Fatalf("mapping started before contract validation: %d calls", fake.mapCalls)
	}

	workflow := testWorkflow(t)
	workflow.Sources = append(workflow.Sources, contract.Source{ID: "second", Format: contract.FormatCSV, Location: filepath.Join(t.TempDir(), "second.csv")})
	plan, err := service.Preflight(context.Background(), contract.RunRequest{Workflow: workflow})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Mappings) != 2 || fake.mapCalls != 2 {
		t.Fatalf("mappings = %d, calls = %d", len(plan.Mappings), fake.mapCalls)
	}
}

func TestPreflightRejectsUndeclaredOperationParameterBeforeMapping(t *testing.T) {
	fake := &planningAdapter{format: contract.FormatCSV}
	service := testService(t, fake)
	workflow := testWorkflow(t)
	workflow.Operations = []contract.Operation{{
		ID: "rename", Kind: contract.OperationRename, Inputs: []string{"source"},
		Rename: []contract.RenameField{{Column: contract.ColumnRef{Header: "name"}, To: "full_name"}},
		Sort:   []contract.SortField{{Column: contract.ColumnRef{Header: "name"}}},
	}}
	workflow.Outputs[0].Input = "rename"

	_, err := service.Preflight(context.Background(), contract.RunRequest{Workflow: workflow})
	if err == nil || !strings.Contains(err.Error(), "$.operations[0].sort") {
		t.Fatalf("error = %v, want undeclared sort parameter path", err)
	}
	if fake.mapCalls != 0 {
		t.Fatalf("mapping started before operation registry validation: %d calls", fake.mapCalls)
	}
}

func TestPreflightBuildsExplicitOrderedPlan(t *testing.T) {
	fake := &planningAdapter{format: contract.FormatCSV}
	service := testService(t, fake)
	workflow := testWorkflow(t)
	workflow.Operations = []contract.Operation{{
		ID: "rename", Kind: contract.OperationRename, Inputs: []string{"source"},
		Rename: []contract.RenameField{{Column: contract.ColumnRef{Header: "name"}, To: "full_name"}},
	}}
	workflow.Outputs[0].Input = "rename"

	plan, err := service.Preflight(context.Background(), contract.RunRequest{Workflow: workflow})
	if err != nil {
		t.Fatal(err)
	}
	wantKinds := []StepKind{StepRead, StepOperation, StepProjection, StepWrite, StepValidate, StepPublish}
	if len(plan.Steps) != len(wantKinds) {
		t.Fatalf("step count = %d, want %d: %#v", len(plan.Steps), len(wantKinds), plan.Steps)
	}
	for index, want := range wantKinds {
		if plan.Steps[index].Order != index || plan.Steps[index].Kind != want ||
			plan.Steps[index].MaximumMemoryBytes != plan.Settings.MaximumMemoryBytes {
			t.Fatalf("step[%d] = %#v, want kind %q with budget", index, plan.Steps[index], want)
		}
	}
}

func TestOutputCollisionBlocksAfterMappingBeforePlanExecution(t *testing.T) {
	fake := &planningAdapter{format: contract.FormatCSV}
	service := testService(t, fake)
	workflow := testWorkflow(t)
	if err := os.WriteFile(workflow.Outputs[0].Location, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Preflight(context.Background(), contract.RunRequest{Workflow: workflow}); err == nil {
		t.Fatal("expected blocking output collision")
	}
	if fake.mapCalls != 1 {
		t.Fatalf("mapping calls = %d, want successful mapping before collision check", fake.mapCalls)
	}
}

func testService(t *testing.T, selected adapter.Adapter) *Service {
	t.Helper()
	registry := adapter.NewRegistry()
	if err := registry.Register(selected); err != nil {
		t.Fatal(err)
	}
	return &Service{
		Adapters:   registry,
		Defaults:   config.BuiltInDefaults(t.TempDir()),
		UserConfig: config.Config{SchemaVersion: config.CurrentSchemaVersion},
		Now:        func() time.Time { return time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC) },
	}
}

func testWorkflow(t *testing.T) contract.Workflow {
	t.Helper()
	root := t.TempDir()
	return contract.Workflow{
		SchemaVersion: contract.CurrentSchemaVersion,
		ID:            "workflow",
		Sources: []contract.Source{{
			ID: "source", Format: contract.FormatCSV, Location: filepath.Join(root, "source.csv"),
		}},
		Outputs: []contract.Output{{
			ID: "output", Input: "source", Format: contract.FormatCSV,
			Location:   filepath.Join(root, "output.csv"),
			Projection: []contract.ProjectionField{{Column: contract.ColumnRef{Header: "name"}, OutputAs: "name"}},
		}},
		Interface: contract.InterfaceMetadata{Name: "cli"},
		Exception: contract.ExceptionPolicy{Action: contract.ExceptionBlock},
	}
}

type planningAdapter struct {
	format   contract.DataFormat
	mapCalls int
}

func (selected *planningAdapter) Format() contract.DataFormat { return selected.format }
func (selected *planningAdapter) Capabilities() adapter.Capabilities {
	return adapter.Capabilities{
		adapter.CapabilityInspect: true, adapter.CapabilityMapBasic: true,
		adapter.CapabilityRead: true, adapter.CapabilityWrite: true, adapter.CapabilityValidate: true,
	}
}
func (selected *planningAdapter) Inspect(context.Context, adapter.SourceRequest) (adapter.Inspection, error) {
	return adapter.Inspection{}, nil
}
func (selected *planningAdapter) Map(_ context.Context, request adapter.SourceRequest, _ adapter.MappingMode) (adapter.MappingReport, error) {
	selected.mapCalls++
	return adapter.MappingReport{SourceID: table.SourceID(request.Source.ID)}, nil
}
func (selected *planningAdapter) OpenRows(context.Context, adapter.SourceRequest, table.Table) (table.RowStream, error) {
	return table.NewSliceRowStream(nil), nil
}
func (selected *planningAdapter) NewWriter(context.Context, adapter.OutputRequest, table.Table) (adapter.RowWriter, error) {
	return nil, nil
}
func (selected *planningAdapter) ValidateOutput(context.Context, adapter.OutputRequest) error {
	return nil
}
