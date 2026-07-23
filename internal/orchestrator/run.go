package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/report"
)

type Observer interface {
	Started(context.Context, contract.StartedResult) error
}

type ObserverFunc func(context.Context, contract.StartedResult) error

func (function ObserverFunc) Started(ctx context.Context, result contract.StartedResult) error {
	return function(ctx, result)
}

func (service *Service) Run(ctx context.Context, request contract.RunRequest) (contract.FinalResult, error) {
	plan, err := service.Preflight(ctx, request)
	if err != nil {
		return contract.FinalResult{}, err
	}
	startedAt := service.currentTime()
	runID := "run-" + startedAt.Format("20060102T150405.000000000Z")
	if service.RunID != nil {
		runID = service.RunID()
	}
	started := contract.StartedResult{
		SchemaVersion: contract.CurrentSchemaVersion,
		RunID:         runID, WorkflowID: plan.Workflow.ID,
		Status: contract.RunStatusStarted, StartedAt: startedAt,
	}
	if err := contract.ValidateStartedResult(started); err != nil {
		return contract.FinalResult{}, err
	}
	if service.Observer != nil {
		if err := service.Observer.Started(ctx, started); err != nil {
			return contract.FinalResult{}, fmt.Errorf("emit started status: %w", err)
		}
	}
	if service.Backend == nil {
		err := errors.New("execution backend is required")
		return service.failedResult(started, plan, ExecutionOutcome{}, err), err
	}

	outcome, executionError := service.Backend.Execute(ctx, plan)
	outcome.Exceptions = report.OrderExceptions(outcome.Exceptions)
	if executionError != nil {
		service.abortPending(plan, outcome.Outputs)
		redacted := service.Redactor.RedactError(executionError)
		return service.failedResult(started, plan, outcome, redacted), redacted
	}
	for _, validation := range outcome.Validations {
		if !validation.Passed {
			service.abortPending(plan, outcome.Outputs)
			err := fmt.Errorf("required validation %q failed: %s", validation.ID, validation.Message)
			return service.failedResult(started, plan, outcome, err), err
		}
	}

	pending := make(map[string]PendingOutput, len(outcome.Outputs))
	for _, output := range outcome.Outputs {
		if output.Stage == nil {
			err := fmt.Errorf("execution output %q has no staged file", output.ID)
			return service.failedResult(started, plan, outcome, err), err
		}
		pending[output.ID] = output
	}
	if len(pending) != len(plan.Workflow.Outputs) {
		service.abortPending(plan, outcome.Outputs)
		err := fmt.Errorf("execution produced %d staged outputs, want %d", len(pending), len(plan.Workflow.Outputs))
		return service.failedResult(started, plan, outcome, err), err
	}

	produced := make([]contract.ProducedOutput, 0, len(plan.Workflow.Outputs))
	for _, definition := range plan.Workflow.Outputs {
		output, exists := pending[definition.ID]
		if !exists {
			service.abortPending(plan, outcome.Outputs)
			err := fmt.Errorf("execution did not stage output %q", definition.ID)
			return service.failedResult(started, plan, outcome, err), err
		}
		location, err := output.Stage.Publish(plan.OutputSettings[definition.ID].Collision)
		if err != nil {
			service.abortPending(plan, outcome.Outputs)
			redacted := service.Redactor.RedactError(err)
			return service.failedResult(started, plan, outcome, redacted), redacted
		}
		produced = append(produced, contract.ProducedOutput{
			ID: definition.ID, Format: definition.Format, Location: location, Rows: output.Rows,
		})
	}

	final := contract.FinalResult{
		SchemaVersion: contract.CurrentSchemaVersion,
		RunID:         started.RunID, WorkflowID: started.WorkflowID,
		Status: contract.RunStatusSucceeded, StartedAt: started.StartedAt, FinishedAt: service.currentTime(),
		Outputs: produced, QueryAnswers: outcome.QueryAnswers,
		ResolvedSettings: resolvedReportSettings(plan),
		Validations:      outcome.Validations, Warnings: outcome.Warnings, Exceptions: outcome.Exceptions,
	}
	final = service.Redactor.RedactFinalResult(final)
	if err := contract.ValidateFinalResult(final); err != nil {
		return final, err
	}
	return final, nil
}

func (service *Service) failedResult(started contract.StartedResult, plan Plan, outcome ExecutionOutcome, cause error) contract.FinalResult {
	final := contract.FinalResult{
		SchemaVersion: contract.CurrentSchemaVersion,
		RunID:         started.RunID, WorkflowID: started.WorkflowID,
		Status: contract.RunStatusFailed, StartedAt: started.StartedAt, FinishedAt: service.currentTime(),
		QueryAnswers: outcome.QueryAnswers, ResolvedSettings: resolvedReportSettings(plan),
		Validations: outcome.Validations, Warnings: outcome.Warnings, Exceptions: outcome.Exceptions,
		Errors: []contract.Diagnostic{{Code: "execution_failed", Message: cause.Error()}},
	}
	return service.Redactor.RedactFinalResult(final)
}

func (service *Service) abortPending(plan Plan, outputs []PendingOutput) {
	for _, output := range outputs {
		if output.Stage == nil {
			continue
		}
		retain := plan.OutputSettings[output.ID].FailedRunRetention > 0
		_, _ = output.Stage.Abort(retain)
	}
}

func (service *Service) currentTime() time.Time {
	if service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}

func resolvedReportSettings(plan Plan) map[string]any {
	settings := plan.Settings.ReportSettings()
	outputs := make(map[string]any, len(plan.OutputSettings))
	for id, resolved := range plan.OutputSettings {
		outputs[id] = resolved.ReportSettings()
	}
	settings["outputs"] = outputs
	return settings
}
