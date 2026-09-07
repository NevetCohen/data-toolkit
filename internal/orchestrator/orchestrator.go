// Package orchestrator validates workflows, creates deterministic plans, and
// dispatches registered extensions without implementing data or file logic.
package orchestrator

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"data-toolkit/internal/core/table"
	fileformat "data-toolkit/internal/fileengine/format"
	logicaloperation "data-toolkit/internal/logical/operation"
	"data-toolkit/internal/report"
)

const CurrentVersion = "v1"

type Workflow struct {
	SchemaVersion string                  `json:"schema_version" yaml:"schema_version"`
	ID            string                  `json:"id" yaml:"id"`
	Inputs        []string                `json:"inputs" yaml:"inputs"`
	Operations    []logicaloperation.Spec `json:"operations" yaml:"operations"`
}

type PlanStep struct {
	Order  int      `json:"order"`
	ID     string   `json:"id"`
	Kind   string   `json:"kind"`
	Inputs []string `json:"inputs"`
}

type Plan struct {
	WorkflowID string     `json:"workflow_id"`
	Steps      []PlanStep `json:"steps"`
}

type Observer interface {
	OnEvent(report.Event)
}

type ObserverFunc func(report.Event)

func (function ObserverFunc) OnEvent(event report.Event) {
	function(event)
}

type RunRequest struct {
	Workflow Workflow
	Inputs   map[string]table.Dataset
}

type RunResult struct {
	Outputs map[string]table.Dataset `json:"-"`
	Events  []report.Event           `json:"events"`
}

type Service struct {
	Operations *logicaloperation.Registry
	Formats    *fileformat.Registry
	Observer   Observer
}

func (service Service) ValidateWorkflow(workflow Workflow) error {
	if service.Operations == nil {
		return fmt.Errorf("validate workflow: logical-operation registry is required")
	}
	if workflow.SchemaVersion != CurrentVersion {
		return fmt.Errorf("workflow.schema_version %q is unsupported", workflow.SchemaVersion)
	}
	if strings.TrimSpace(workflow.ID) == "" {
		return fmt.Errorf("workflow.id is required")
	}

	available := make(map[string]struct{}, len(workflow.Inputs)+len(workflow.Operations))
	for index, input := range workflow.Inputs {
		key := normalize(input)
		if key == "" {
			return fmt.Errorf("workflow.inputs[%d] is empty", index)
		}
		if _, exists := available[key]; exists {
			return fmt.Errorf("workflow input identity %q is duplicated", input)
		}
		available[key] = struct{}{}
	}

	for index, spec := range workflow.Operations {
		id := normalize(spec.ID)
		if id == "" {
			return fmt.Errorf("workflow.operations[%d].id is required", index)
		}
		if normalize(spec.Kind) == "" {
			return fmt.Errorf("workflow operation %q kind is required", spec.ID)
		}
		if _, exists := available[id]; exists {
			return fmt.Errorf("workflow identity %q is duplicated", spec.ID)
		}
		if _, err := service.Operations.Require(spec.Kind); err != nil {
			return fmt.Errorf("workflow operation %q: %w", spec.ID, err)
		}
		for inputIndex, input := range spec.Inputs {
			if _, exists := available[normalize(input)]; !exists {
				return fmt.Errorf("workflow operation %q input[%d] %q is not available from a workflow input or earlier operation", spec.ID, inputIndex, input)
			}
		}
		available[id] = struct{}{}
	}
	return nil
}

func (service Service) Plan(workflow Workflow) (Plan, error) {
	if err := service.ValidateWorkflow(workflow); err != nil {
		return Plan{}, err
	}
	steps := make([]PlanStep, len(workflow.Operations))
	for index, spec := range workflow.Operations {
		steps[index] = PlanStep{
			Order:  index + 1,
			ID:     spec.ID,
			Kind:   spec.Kind,
			Inputs: append([]string(nil), spec.Inputs...),
		}
	}
	return Plan{WorkflowID: workflow.ID, Steps: steps}, nil
}

func (service Service) Map(ctx context.Context, request fileformat.MapRequest) (fileformat.Mapping, error) {
	if service.Formats == nil {
		return fileformat.Mapping{}, fmt.Errorf("map source: file-format registry is required")
	}
	mapping, err := service.Formats.Map(ctx, request)
	if err != nil {
		return fileformat.Mapping{}, fmt.Errorf("map source: %w", err)
	}
	for index, schema := range mapping.Tables {
		if err := schema.Validate(); err != nil {
			return fileformat.Mapping{}, fmt.Errorf("map source table[%d]: %w", index, err)
		}
	}
	return mapping, nil
}

func (service Service) Run(ctx context.Context, request RunRequest) (RunResult, error) {
	events := make([]report.Event, 0, 4+2*len(request.Workflow.Operations))
	emit := func(kind, component, message string) {
		event := report.Event{
			Sequence:  uint64(len(events) + 1),
			Kind:      kind,
			Component: component,
			Message:   message,
		}
		events = append(events, event)
		if service.Observer != nil {
			service.Observer.OnEvent(event)
		}
	}

	emit("workflow.validation.started", request.Workflow.ID, "")
	plan, err := service.Plan(request.Workflow)
	if err != nil {
		emit("workflow.failed", request.Workflow.ID, err.Error())
		return RunResult{Events: events}, err
	}
	emit("workflow.planned", request.Workflow.ID, fmt.Sprintf("%d step(s)", len(plan.Steps)))

	resolved := make(map[string]table.Dataset, len(request.Inputs)+len(plan.Steps))
	for _, identity := range request.Workflow.Inputs {
		dataset, exists := request.Inputs[identity]
		if !exists {
			dataset, exists = lookupDataset(request.Inputs, identity)
		}
		if !exists {
			err := fmt.Errorf("workflow input %q was not provided", identity)
			emit("workflow.failed", request.Workflow.ID, err.Error())
			return RunResult{Events: events}, err
		}
		if err := dataset.Validate(); err != nil {
			err = fmt.Errorf("workflow input %q: %w", identity, err)
			emit("workflow.failed", request.Workflow.ID, err.Error())
			return RunResult{Events: events}, err
		}
		resolved[normalize(identity)] = dataset
	}

	outputs := make(map[string]table.Dataset, len(plan.Steps))
	for _, spec := range request.Workflow.Operations {
		inputs := make([]table.Dataset, len(spec.Inputs))
		for index, identity := range spec.Inputs {
			inputs[index] = resolved[normalize(identity)]
		}
		emit("operation.started", spec.ID, spec.Kind)
		result, err := service.Operations.Execute(ctx, logicaloperation.Request{Spec: spec, Inputs: inputs})
		if err != nil {
			emit("workflow.failed", request.Workflow.ID, err.Error())
			return RunResult{Outputs: outputs, Events: events}, err
		}
		resolved[normalize(spec.ID)] = result.Dataset
		outputs[spec.ID] = result.Dataset
		emit("operation.completed", spec.ID, spec.Kind)
	}
	emit("workflow.completed", request.Workflow.ID, "")
	return RunResult{Outputs: outputs, Events: events}, nil
}

func lookupDataset(values map[string]table.Dataset, identity string) (table.Dataset, bool) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if normalize(key) == normalize(identity) {
			return values[key], true
		}
	}
	return table.Dataset{}, false
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
