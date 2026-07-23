package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync/atomic"

	"data-toolkit/internal/adapter"
	"data-toolkit/internal/contract"
	logicaloperation "data-toolkit/internal/operation"
	"data-toolkit/internal/table"
)

var pipelineRunSequence atomic.Uint64

// LocalPipelineBackend implements the streaming JSON/CSV vertical slice. It
// always reads local sources through validated immutable snapshots.
type LocalPipelineBackend struct {
	Adapters *adapter.Registry
	Retry    adapter.RetryPolicy
}

func (backend *LocalPipelineBackend) Execute(ctx context.Context, plan Plan) (ExecutionOutcome, error) {
	if backend.Adapters == nil {
		return ExecutionOutcome{}, errors.New("local pipeline requires adapter registry")
	}
	runID := "pipeline-" + strconv.FormatUint(pipelineRunSequence.Add(1), 10)
	workspace, err := adapter.NewRunWorkspace(
		plan.Settings.TemporaryWorkspace, runID, plan.Settings.FailedRunRetention,
	)
	if err != nil {
		return ExecutionOutcome{}, err
	}
	outcome := ExecutionOutcome{}
	fail := func(cause error) (ExecutionOutcome, error) {
		_, finalizeError := workspace.Finalize(context.Background(), false)
		if finalizeError != nil {
			cause = errors.Join(cause, finalizeError)
		}
		return outcome, cause
	}

	snapshots := make(map[string]adapter.SourceSnapshot, len(plan.Workflow.Sources))
	tables := make(map[string]table.Table, len(plan.Workflow.Sources))
	for sourceIndex, source := range plan.Workflow.Sources {
		if source.Format == contract.FormatGoogleSheets {
			return fail(errors.New("local pipeline does not process Google sources"))
		}
		snapshot, err := adapter.CreateSourceSnapshot(
			ctx, source.Location, workspace.Path(), backend.Retry,
		)
		if err != nil {
			return fail(fmt.Errorf("snapshot source %q: %w", source.ID, err))
		}
		snapshots[source.ID] = snapshot
		selected, err := backend.Adapters.Require(source.Format, adapter.CapabilityMapBasic, adapter.CapabilityRead)
		if err != nil {
			return fail(err)
		}
		request := adapter.SourceRequest{
			Source: source, SnapshotPath: snapshot.SnapshotPath,
			Mappings: sourceMappings(plan.Workflow.Mappings, source.ID),
		}
		report, err := selected.Map(ctx, request, adapter.MappingBasic)
		if err != nil {
			return fail(fmt.Errorf("map snapshotted source %q: %w", source.ID, err))
		}
		if len(report.Tables) != 1 {
			return fail(fmt.Errorf("source %q mapped %d tables, local slice requires one", source.ID, len(report.Tables)))
		}
		_ = sourceIndex
		tables[source.ID] = report.Tables[0]
	}

	for _, output := range plan.Workflow.Outputs {
		source, sourceTable, err := pipelineInput(plan.Workflow, tables, output.Input)
		if err != nil {
			return fail(fmt.Errorf("resolve output %q input: %w", output.ID, err))
		}
		sourceAdapter, err := backend.Adapters.Require(source.Format, adapter.CapabilityRead)
		if err != nil {
			return fail(err)
		}
		sourceRequest := adapter.SourceRequest{
			Source: source, SnapshotPath: snapshots[source.ID].SnapshotPath,
			Mappings: sourceMappings(plan.Workflow.Mappings, source.ID),
		}
		stream, err := sourceAdapter.OpenRows(ctx, sourceRequest, sourceTable)
		if err != nil {
			return fail(err)
		}
		transformedTable, transformedStream, err := applyLinearOperations(plan.Workflow, source.ID, sourceTable, stream, output.Input)
		if err != nil {
			stream.Close()
			return fail(err)
		}

		stage, err := adapter.NewStagedOutput(plan.OutputLocations[output.ID])
		if err != nil {
			transformedStream.Close()
			return fail(err)
		}
		outcome.Outputs = append(outcome.Outputs, PendingOutput{ID: output.ID, Stage: stage})
		outputAdapter, err := backend.Adapters.Require(output.Format, adapter.CapabilityWrite, adapter.CapabilityValidate)
		if err != nil {
			transformedStream.Close()
			return fail(err)
		}
		settings := plan.OutputSettings[output.ID]
		outputRequest := adapter.OutputRequest{
			Output: output, StagedPath: stage.Path(),
			Render: table.RenderOptions{
				NullDisplay: settings.NullDisplay, DecimalPrecision: settings.DecimalPrecision,
				DateFormat: settings.DateFormat, TimeFormat: settings.TimeFormat,
			},
		}
		writer, err := outputAdapter.NewWriter(ctx, outputRequest, transformedTable)
		if err != nil {
			transformedStream.Close()
			return fail(err)
		}
		var rowCount int64
		for {
			row, err := transformedStream.Next(ctx)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				writer.Close()
				transformedStream.Close()
				return fail(err)
			}
			if err := writer.WriteRow(ctx, row); err != nil {
				writer.Close()
				transformedStream.Close()
				return fail(err)
			}
			rowCount++
		}
		if err := transformedStream.Close(); err != nil {
			writer.Close()
			return fail(err)
		}
		if err := writer.Close(); err != nil {
			return fail(err)
		}
		if err := stage.Validate(ctx, func(ctx context.Context, _ string) error {
			return outputAdapter.ValidateOutput(ctx, outputRequest)
		}); err != nil {
			outcome.Validations = append(outcome.Validations, contract.ValidationResult{
				ID: "output:" + output.ID, Passed: false, Message: err.Error(),
			})
			return fail(err)
		}
		outcome.Outputs[len(outcome.Outputs)-1].Rows = rowCount
		outcome.Validations = append(outcome.Validations, contract.ValidationResult{
			ID: "output:" + output.ID, Passed: true,
		})
	}

	for _, query := range plan.Workflow.Queries {
		if query.Kind != contract.QueryCount {
			return fail(fmt.Errorf("local pipeline query %q kind %q is not implemented", query.ID, query.Kind))
		}
		source, sourceTable, err := pipelineInput(plan.Workflow, tables, query.Input)
		if err != nil {
			return fail(err)
		}
		selected, err := backend.Adapters.Require(source.Format, adapter.CapabilityRead)
		if err != nil {
			return fail(err)
		}
		stream, err := selected.OpenRows(ctx, adapter.SourceRequest{
			Source: source, SnapshotPath: snapshots[source.ID].SnapshotPath,
			Mappings: sourceMappings(plan.Workflow.Mappings, source.ID),
		}, sourceTable)
		if err != nil {
			return fail(err)
		}
		_, transformed, err := applyLinearOperations(plan.Workflow, source.ID, sourceTable, stream, query.Input)
		if err != nil {
			stream.Close()
			return fail(err)
		}
		var count int64
		for {
			_, err := transformed.Next(ctx)
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				transformed.Close()
				return fail(err)
			}
			count++
		}
		transformed.Close()
		outcome.QueryAnswers = append(outcome.QueryAnswers, contract.QueryAnswer{
			ID: query.ID, Shape: contract.QueryShapeScalar, Value: strconv.FormatInt(count, 10),
		})
	}

	if _, err := workspace.Finalize(context.Background(), true); err != nil {
		return outcome, err
	}
	return outcome, nil
}

func pipelineInput(workflow contract.Workflow, tables map[string]table.Table, input string) (contract.Source, table.Table, error) {
	sourceID := input
	for _, operation := range workflow.Operations {
		if operation.ID == input {
			if len(operation.Inputs) != 1 {
				return contract.Source{}, table.Table{}, fmt.Errorf("operation %q must have exactly one input in local pipeline", input)
			}
			sourceID = operation.Inputs[0]
			break
		}
	}
	for _, source := range workflow.Sources {
		if source.ID == sourceID {
			return source, tables[source.ID], nil
		}
	}
	return contract.Source{}, table.Table{}, fmt.Errorf("input %q does not resolve to a source", input)
}

func applyLinearOperations(workflow contract.Workflow, sourceID string, sourceTable table.Table, source table.RowStream, until string) (table.Table, table.RowStream, error) {
	currentTable := sourceTable
	currentStream := source
	if until == sourceID {
		return currentTable, currentStream, nil
	}
	for _, operation := range workflow.Operations {
		if len(operation.Inputs) != 1 || (operation.Inputs[0] != sourceID && operation.Inputs[0] != previousOperationID(workflow.Operations, operation.ID)) {
			continue
		}
		switch operation.Kind {
		case contract.OperationConcatenate:
			if operation.Text == nil {
				return table.Table{}, nil, fmt.Errorf("operation %q has no text options", operation.ID)
			}
			nextTable, nextStream, err := logicaloperation.ApplyConcatenate(currentTable, currentStream, *operation.Text)
			if err != nil {
				return table.Table{}, nil, fmt.Errorf("operation %q: %w", operation.ID, err)
			}
			currentTable, currentStream = nextTable, nextStream
		case contract.OperationRename:
			nextTable, nextStream, err := logicaloperation.ApplyRename(currentTable, currentStream, operation.Rename)
			if err != nil {
				return table.Table{}, nil, fmt.Errorf("operation %q: %w", operation.ID, err)
			}
			currentTable, currentStream = nextTable, nextStream
		default:
			return table.Table{}, nil, fmt.Errorf("operation %q kind %q is not implemented by local pipeline", operation.ID, operation.Kind)
		}
		if operation.ID == until {
			return currentTable, currentStream, nil
		}
	}
	return table.Table{}, nil, fmt.Errorf("operation input %q was not reached", until)
}

func previousOperationID(operations []contract.Operation, id string) string {
	for index, operation := range operations {
		if operation.ID == id && index > 0 {
			return operations[index-1].ID
		}
	}
	return ""
}
