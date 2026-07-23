package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"data-toolkit/internal/adapter"
	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

type PendingOutput struct {
	ID    string
	Stage *adapter.StagedOutput
	Rows  int64
}

type ExecutionOutcome struct {
	Outputs      []PendingOutput
	QueryAnswers []contract.QueryAnswer
	Validations  []contract.ValidationResult
	Warnings     []contract.Diagnostic
	Exceptions   []contract.ExceptionRecord
}

type ExecutionBackend interface {
	Execute(context.Context, Plan) (ExecutionOutcome, error)
}

type StepResult struct {
	Outputs      []PendingOutput
	QueryAnswers []contract.QueryAnswer
	Validations  []contract.ValidationResult
	Warnings     []contract.Diagnostic
	Exceptions   []contract.ExceptionRecord
}

type StepProcessor interface {
	Process(context.Context, PlanStep, *Runtime) (StepResult, error)
}

type DeterministicExecutor struct {
	Processor          StepProcessor
	TemporaryWorkspace string
}

func (executor *DeterministicExecutor) Execute(ctx context.Context, plan Plan) (ExecutionOutcome, error) {
	if executor.Processor == nil {
		return ExecutionOutcome{}, errors.New("step processor is required")
	}
	runtime := NewRuntime(plan.Settings.MaximumMemoryBytes, executor.TemporaryWorkspace)
	outcome := ExecutionOutcome{}
	for index, step := range plan.Steps {
		if err := ctx.Err(); err != nil {
			return outcome, err
		}
		if step.Order != index {
			return outcome, fmt.Errorf("plan step %q has order %d, want %d", step.ID, step.Order, index)
		}
		if step.MaximumMemoryBytes <= 0 || step.MaximumMemoryBytes > plan.Settings.MaximumMemoryBytes {
			return outcome, fmt.Errorf("plan step %q has invalid memory budget %d", step.ID, step.MaximumMemoryBytes)
		}
		result, err := executor.Processor.Process(ctx, step, runtime)
		outcome.Outputs = append(outcome.Outputs, result.Outputs...)
		outcome.QueryAnswers = append(outcome.QueryAnswers, result.QueryAnswers...)
		outcome.Validations = append(outcome.Validations, result.Validations...)
		outcome.Warnings = append(outcome.Warnings, result.Warnings...)
		outcome.Exceptions = append(outcome.Exceptions, result.Exceptions...)
		if err != nil {
			return outcome, fmt.Errorf("execute step %q: %w", step.ID, err)
		}
	}
	if runtime.UsedMemory() != 0 {
		return outcome, fmt.Errorf("execution leaked %d reserved memory bytes", runtime.UsedMemory())
	}
	return outcome, nil
}

type Runtime struct {
	mu                 sync.Mutex
	maximumMemoryBytes int64
	usedMemoryBytes    int64
	temporaryWorkspace string
}

func NewRuntime(maximumMemoryBytes int64, temporaryWorkspace string) *Runtime {
	return &Runtime{maximumMemoryBytes: maximumMemoryBytes, temporaryWorkspace: temporaryWorkspace}
}

func (runtime *Runtime) ReserveMemory(bytes int64) error {
	if bytes < 0 {
		return errors.New("memory reservation must not be negative")
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if runtime.usedMemoryBytes+bytes > runtime.maximumMemoryBytes {
		return fmt.Errorf("memory budget exceeded: requested %d with %d of %d already reserved", bytes, runtime.usedMemoryBytes, runtime.maximumMemoryBytes)
	}
	runtime.usedMemoryBytes += bytes
	return nil
}

func (runtime *Runtime) ReleaseMemory(bytes int64) error {
	if bytes < 0 {
		return errors.New("memory release must not be negative")
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if bytes > runtime.usedMemoryBytes {
		return fmt.Errorf("memory release %d exceeds reserved %d", bytes, runtime.usedMemoryBytes)
	}
	runtime.usedMemoryBytes -= bytes
	return nil
}

func (runtime *Runtime) UsedMemory() int64 {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	return runtime.usedMemoryBytes
}

func (runtime *Runtime) NewSpool(memoryBytes int64) (*ManagedSpool, error) {
	if memoryBytes <= 0 {
		return nil, errors.New("spool memory must be greater than zero")
	}
	if err := runtime.ReserveMemory(memoryBytes); err != nil {
		return nil, err
	}
	spool, err := table.NewBoundedSpool(runtime.temporaryWorkspace, memoryBytes)
	if err != nil {
		_ = runtime.ReleaseMemory(memoryBytes)
		return nil, err
	}
	return &ManagedSpool{BoundedSpool: spool, runtime: runtime, reserved: memoryBytes}, nil
}

type ManagedSpool struct {
	*table.BoundedSpool
	runtime  *Runtime
	reserved int64
	closed   bool
}

func (spool *ManagedSpool) Close() error {
	if spool.closed {
		return nil
	}
	spool.closed = true
	closeError := spool.BoundedSpool.Close()
	releaseError := spool.runtime.ReleaseMemory(spool.reserved)
	if closeError != nil {
		return closeError
	}
	return releaseError
}
