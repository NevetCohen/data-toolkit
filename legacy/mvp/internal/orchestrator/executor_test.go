package orchestrator

import (
	"context"
	"errors"
	"io"
	"testing"

	"data-toolkit/internal/config"
	"data-toolkit/internal/table"
)

func TestDeterministicExecutorUsesOrderedStepsAndBoundedSpool(t *testing.T) {
	processor := &recordingProcessor{t: t}
	settings := config.BuiltInDefaults(t.TempDir())
	settings.MaximumMemoryBytes = 64
	plan := Plan{
		Settings: settings,
		Steps: []PlanStep{
			{Order: 0, ID: "first", MaximumMemoryBytes: 64},
			{Order: 1, ID: "second", MaximumMemoryBytes: 64},
		},
	}
	executor := &DeterministicExecutor{Processor: processor, TemporaryWorkspace: t.TempDir()}
	if _, err := executor.Execute(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	if len(processor.steps) != 2 || processor.steps[0] != "first" || processor.steps[1] != "second" {
		t.Fatalf("execution order = %#v", processor.steps)
	}
}

func TestDeterministicExecutorHonorsCancellationBetweenSteps(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	processor := StepProcessorFunc(func(_ context.Context, step PlanStep, _ *Runtime) (StepResult, error) {
		if step.Order == 0 {
			cancel()
		}
		return StepResult{}, nil
	})
	settings := config.BuiltInDefaults(t.TempDir())
	plan := Plan{Settings: settings, Steps: []PlanStep{
		{Order: 0, ID: "first", MaximumMemoryBytes: settings.MaximumMemoryBytes},
		{Order: 1, ID: "second", MaximumMemoryBytes: settings.MaximumMemoryBytes},
	}}
	executor := &DeterministicExecutor{Processor: processor, TemporaryWorkspace: t.TempDir()}
	if _, err := executor.Execute(ctx, plan); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

type StepProcessorFunc func(context.Context, PlanStep, *Runtime) (StepResult, error)

func (function StepProcessorFunc) Process(ctx context.Context, step PlanStep, runtime *Runtime) (StepResult, error) {
	return function(ctx, step, runtime)
}

type recordingProcessor struct {
	t     *testing.T
	steps []string
}

func (processor *recordingProcessor) Process(ctx context.Context, step PlanStep, runtime *Runtime) (StepResult, error) {
	processor.steps = append(processor.steps, step.ID)
	if step.Order != 0 {
		return StepResult{}, nil
	}
	if err := runtime.ReserveMemory(65); err == nil {
		processor.t.Fatal("memory limiter accepted a reservation above the budget")
	}
	spool, err := runtime.NewSpool(32)
	if err != nil {
		return StepResult{}, err
	}
	defer spool.Close()
	for index, text := range []string{"ראשון", "שני 🧭"} {
		row := table.Row{ID: table.RowID(text), SourceID: "source", SheetID: "sheet", Ordinal: uint64(index)}
		if err := spool.Append(row); err != nil {
			return StepResult{}, err
		}
	}
	stream, err := spool.Stream()
	if err != nil {
		return StepResult{}, err
	}
	defer stream.Close()
	for _, want := range []table.RowID{"ראשון", "שני 🧭"} {
		row, err := stream.Next(ctx)
		if err != nil {
			return StepResult{}, err
		}
		if row.ID != want {
			processor.t.Fatalf("spooled row = %q, want %q", row.ID, want)
		}
	}
	if _, err := stream.Next(ctx); !errors.Is(err, io.EOF) {
		processor.t.Fatalf("spool end = %v", err)
	}
	return StepResult{}, nil
}
