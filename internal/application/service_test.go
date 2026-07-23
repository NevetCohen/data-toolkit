package application_test

import (
	"context"
	"encoding/json"
	"testing"

	"data-toolkit/internal/application"
	"data-toolkit/internal/config"
	"data-toolkit/internal/core/datatype"
	"data-toolkit/internal/core/table"
	fileformat "data-toolkit/internal/fileengine/format"
	logicaloperation "data-toolkit/internal/logical/operation"
	"data-toolkit/internal/orchestrator"
	"data-toolkit/internal/report"
)

type passThrough struct{}

func (passThrough) Descriptor() logicaloperation.Descriptor {
	return logicaloperation.Descriptor{
		Kind:            "test.pass",
		Version:         logicaloperation.CurrentVersion,
		Name:            "Pass through",
		Deterministic:   true,
		MinimumInputs:   1,
		MaximumInputs:   1,
		ParameterSchema: json.RawMessage(`{"type":"object","additionalProperties":false}`),
	}
}

func (passThrough) Validate(logicaloperation.Spec, []table.Schema) error {
	return nil
}

func (passThrough) Execute(_ context.Context, request logicaloperation.Request) (logicaloperation.Result, error) {
	return logicaloperation.Result{Dataset: request.Inputs[0]}, nil
}

func TestApplicationAPIPlansAndRunsRegisteredOperations(t *testing.T) {
	dataTypes, err := datatype.NewBuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	operations := logicaloperation.NewRegistry(logicaloperation.CurrentVersion)
	if err := operations.Register(passThrough{}); err != nil {
		t.Fatal(err)
	}
	defaults, err := config.LoadDefaults()
	if err != nil {
		t.Fatal(err)
	}
	var observed []report.Event
	service, err := application.New(application.Options{
		DataTypes:  dataTypes,
		Operations: operations,
		Formats:    fileformat.NewRegistry(fileformat.CurrentVersion),
		Defaults:   defaults,
		Observer: orchestrator.ObserverFunc(func(event report.Event) {
			observed = append(observed, event)
		}),
	})
	if err != nil {
		t.Fatal(err)
	}

	capabilities, err := service.Capabilities(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(capabilities.DataTypes) != 9 || len(capabilities.LogicalOperations) != 1 || len(capabilities.FileFormats) != 0 {
		t.Fatalf("unexpected capabilities: %#v", capabilities)
	}

	workflow := orchestrator.Workflow{
		SchemaVersion: orchestrator.CurrentVersion,
		ID:            "two-step",
		Inputs:        []string{"source"},
		Operations: []logicaloperation.Spec{
			{ID: "first", Kind: "test.pass", Inputs: []string{"source"}, Parameters: json.RawMessage(`{}`)},
			{ID: "second", Kind: "test.pass", Inputs: []string{"first"}, Parameters: json.RawMessage(`{}`)},
		},
	}
	plan, err := service.Plan(context.Background(), workflow)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Steps) != 2 || plan.Steps[1].Order != 2 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	result, err := service.Run(context.Background(), orchestrator.RunRequest{
		Workflow: workflow,
		Inputs: map[string]table.Dataset{
			"source": {Schema: table.Schema{ID: "source-table", SourceID: "source"}, Rows: table.NewSliceStream(nil)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Outputs) != 2 || result.Outputs["second"].Schema.ID != "source-table" {
		t.Fatalf("unexpected outputs: %#v", result.Outputs)
	}
	if len(result.Events) != 7 || len(observed) != len(result.Events) {
		t.Fatalf("event counts result=%d observed=%d", len(result.Events), len(observed))
	}
	for index, event := range result.Events {
		if event.Sequence != uint64(index+1) {
			t.Fatalf("event sequence[%d] = %d", index, event.Sequence)
		}
	}
	if result.Events[len(result.Events)-1].Kind != "workflow.completed" {
		t.Fatalf("terminal event = %q", result.Events[len(result.Events)-1].Kind)
	}
}

func TestWorkflowRejectsForwardAndUnregisteredDependencies(t *testing.T) {
	dataTypes, _ := datatype.NewBuiltinRegistry()
	operations := logicaloperation.NewRegistry(logicaloperation.CurrentVersion)
	_ = operations.Register(passThrough{})
	defaults, _ := config.LoadDefaults()
	service, _ := application.New(application.Options{
		DataTypes: dataTypes, Operations: operations,
		Formats: fileformat.NewRegistry(fileformat.CurrentVersion), Defaults: defaults,
	})

	for name, workflow := range map[string]orchestrator.Workflow{
		"forward": {
			SchemaVersion: "v1", ID: "forward", Inputs: []string{"source"},
			Operations: []logicaloperation.Spec{
				{ID: "first", Kind: "test.pass", Inputs: []string{"later"}},
				{ID: "later", Kind: "test.pass", Inputs: []string{"source"}},
			},
		},
		"unregistered": {
			SchemaVersion: "v1", ID: "unknown", Inputs: []string{"source"},
			Operations: []logicaloperation.Spec{{ID: "step", Kind: "unknown", Inputs: []string{"source"}}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := service.ValidateWorkflow(context.Background(), workflow); err == nil {
				t.Fatal("invalid workflow was accepted")
			}
		})
	}
}
