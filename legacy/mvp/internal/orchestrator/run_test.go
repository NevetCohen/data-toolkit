package orchestrator

import (
	"context"
	"errors"
	"os"
	"testing"

	"data-toolkit/internal/adapter"
	"data-toolkit/internal/contract"
)

func TestStartedIsEmittedOnlyAfterBlockingPreflight(t *testing.T) {
	selected := &planningAdapter{format: contract.FormatCSV}
	service := testService(t, selected)
	observerCalls := 0
	backend := &outcomeBackend{}
	service.Observer = ObserverFunc(func(context.Context, contract.StartedResult) error {
		observerCalls++
		return nil
	})
	service.Backend = backend
	service.RunID = func() string { return "run-1" }

	blocked := testWorkflow(t)
	if err := os.WriteFile(blocked.Outputs[0].Location, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Run(context.Background(), contract.RunRequest{Workflow: blocked}); err == nil {
		t.Fatal("expected blocking preflight error")
	}
	if observerCalls != 0 || backend.calls != 0 {
		t.Fatalf("started=%d backend=%d before successful preflight", observerCalls, backend.calls)
	}
}

func TestRunReturnsFileAndQueryOnlyAfterValidation(t *testing.T) {
	selected := &planningAdapter{format: contract.FormatCSV}
	service := testService(t, selected)
	service.RunID = func() string { return "run-success" }
	startedCalls := 0
	service.Observer = ObserverFunc(func(_ context.Context, started contract.StartedResult) error {
		startedCalls++
		if started.Status != contract.RunStatusStarted {
			t.Fatalf("started status = %q", started.Status)
		}
		return nil
	})
	service.Backend = &outcomeBackend{t: t, build: func(t *testing.T, plan Plan) ExecutionOutcome {
		stage, err := adapter.NewStagedOutput(plan.OutputLocations["output"])
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(stage.Path(), []byte("name\nנועה 🧭\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := stage.Validate(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
		return ExecutionOutcome{
			Outputs:      []PendingOutput{{ID: "output", Stage: stage, Rows: 1}},
			QueryAnswers: []contract.QueryAnswer{{ID: "count", Shape: contract.QueryShapeScalar, Value: "1"}},
			Validations:  []contract.ValidationResult{{ID: "rows", Passed: true}},
			Warnings:     []contract.Diagnostic{{Code: "fixture", Message: "test warning"}},
			Exceptions: []contract.ExceptionRecord{
				{Reference: "later", SourceOrdinal: 1, RowOrdinal: 0, ColumnOrdinal: 0, Code: "mixed"},
				{Reference: "earlier", SourceOrdinal: 0, RowOrdinal: 2, ColumnOrdinal: 0, Code: "mixed"},
			},
		}
	}}
	workflow := testWorkflow(t)
	workflow.Queries = []contract.Query{{ID: "count", Input: "source", Kind: contract.QueryCount, Shape: contract.QueryShapeScalar}}

	final, err := service.Run(context.Background(), contract.RunRequest{Workflow: workflow})
	if err != nil {
		t.Fatal(err)
	}
	if startedCalls != 1 || final.Status != contract.RunStatusSucceeded ||
		len(final.Outputs) != 1 || len(final.QueryAnswers) != 1 ||
		len(final.Validations) != 1 || len(final.Warnings) != 1 ||
		len(final.Exceptions) != 2 {
		t.Fatalf("incomplete final result: %#v", final)
	}
	if final.Exceptions[0].Reference != "earlier" {
		t.Fatalf("exceptions are not in source order: %#v", final.Exceptions)
	}
	if _, err := os.Stat(final.Outputs[0].Location); err != nil {
		t.Fatal(err)
	}
}

func TestValidationFailurePreventsPublication(t *testing.T) {
	selected := &planningAdapter{format: contract.FormatCSV}
	service := testService(t, selected)
	service.Defaults.FailedRunRetention = 0
	service.RunID = func() string { return "run-failed-validation" }
	var stagedPath string
	service.Backend = &outcomeBackend{t: t, build: func(t *testing.T, plan Plan) ExecutionOutcome {
		stage, err := adapter.NewStagedOutput(plan.OutputLocations["output"])
		if err != nil {
			t.Fatal(err)
		}
		stagedPath = stage.Path()
		if err := os.WriteFile(stage.Path(), []byte("bad"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := stage.Validate(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
		return ExecutionOutcome{
			Outputs:     []PendingOutput{{ID: "output", Stage: stage}},
			Validations: []contract.ValidationResult{{ID: "rows", Passed: false, Message: "expected 2, got 1"}},
		}
	}}
	workflow := testWorkflow(t)
	finalPath := workflow.Outputs[0].Location
	final, err := service.Run(context.Background(), contract.RunRequest{Workflow: workflow})
	if err == nil {
		t.Fatal("expected validation gate error")
	}
	if final.Status != contract.RunStatusFailed || len(final.Outputs) != 0 {
		t.Fatalf("failed result reported successful output: %#v", final)
	}
	if _, statErr := os.Stat(finalPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("final output was published: %v", statErr)
	}
	if _, statErr := os.Stat(stagedPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("failed stage was not removed under zero retention: %v", statErr)
	}
}

func TestHardExecutionFailureReturnsDetailedFinalError(t *testing.T) {
	selected := &planningAdapter{format: contract.FormatCSV}
	service := testService(t, selected)
	service.RunID = func() string { return "run-hard-failure" }
	service.Backend = &outcomeBackend{failure: errors.New("logical engine stopped")}
	workflow := testWorkflow(t)
	workflow.Outputs = nil
	workflow.Queries = []contract.Query{{ID: "count", Input: "source", Kind: contract.QueryCount, Shape: contract.QueryShapeScalar}}

	final, err := service.Run(context.Background(), contract.RunRequest{Workflow: workflow})
	if err == nil {
		t.Fatal("expected execution error")
	}
	if final.Status != contract.RunStatusFailed || len(final.Errors) != 1 ||
		final.Errors[0].Code != "execution_failed" ||
		final.Errors[0].Message == "" {
		t.Fatalf("hard failure result = %#v", final)
	}
}

type outcomeBackend struct {
	calls   int
	build   func(*testing.T, Plan) ExecutionOutcome
	t       *testing.T
	failure error
}

func (backend *outcomeBackend) Execute(_ context.Context, plan Plan) (ExecutionOutcome, error) {
	backend.calls++
	if backend.failure != nil {
		return ExecutionOutcome{}, backend.failure
	}
	if backend.build == nil {
		return ExecutionOutcome{}, nil
	}
	return backend.build(backend.t, plan), nil
}
