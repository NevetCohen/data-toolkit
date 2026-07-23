// Package orchestrator validates, maps, plans, executes, validates, and
// publishes deterministic workflows.
package orchestrator

import (
	"data-toolkit/internal/adapter"
	"data-toolkit/internal/config"
	"data-toolkit/internal/contract"
)

type Engine string

const (
	EngineAdapter      Engine = "adapter"
	EngineLogical      Engine = "logical"
	EngineOrchestrator Engine = "orchestrator"
)

type StepKind string

const (
	StepRead       StepKind = "read"
	StepRepair     StepKind = "repair"
	StepOperation  StepKind = "operation"
	StepProjection StepKind = "projection"
	StepWrite      StepKind = "write"
	StepQuery      StepKind = "query"
	StepValidate   StepKind = "validate"
	StepPublish    StepKind = "publish"
)

type PlanStep struct {
	Order              int
	ID                 string
	Kind               StepKind
	Engine             Engine
	Inputs             []string
	Outputs            []string
	Repairs            []string
	MaximumMemoryBytes int64
	Validations        []contract.Validation
}

type Plan struct {
	Workflow        contract.Workflow
	Mappings        []adapter.MappingReport
	Settings        config.Resolved
	OutputSettings  map[string]config.Resolved
	OutputLocations map[string]string
	Steps           []PlanStep
}

func buildPlan(workflow contract.Workflow, mappings []adapter.MappingReport, settings config.Resolved, outputSettings map[string]config.Resolved, locations map[string]string) Plan {
	steps := make([]PlanStep, 0, len(workflow.Sources)+len(workflow.Operations)+len(workflow.Outputs)*4+len(workflow.Queries))
	appendStep := func(step PlanStep) {
		step.Order = len(steps)
		step.MaximumMemoryBytes = settings.MaximumMemoryBytes
		steps = append(steps, step)
	}
	for _, source := range workflow.Sources {
		appendStep(PlanStep{
			ID: "read:" + source.ID, Kind: StepRead, Engine: EngineAdapter,
			Inputs: []string{source.ID}, Outputs: []string{source.ID + ":rows"},
		})
	}
	for _, mapping := range mappings {
		repairs := make([]string, 0, len(mapping.Findings))
		for _, finding := range mapping.Findings {
			if finding.Code == "format-variant" && settings.AutoNormalizeFormats {
				repairs = append(repairs, "normalize:"+finding.Path)
			}
			if finding.Code == "whitespace" && settings.TrimDetectedWhitespace {
				repairs = append(repairs, "trim:"+finding.Path)
			}
		}
		if len(repairs) > 0 {
			appendStep(PlanStep{
				ID: "repair:" + string(mapping.SourceID), Kind: StepRepair, Engine: EngineLogical,
				Inputs: []string{string(mapping.SourceID) + ":rows"}, Outputs: []string{string(mapping.SourceID) + ":clean"}, Repairs: repairs,
			})
		}
	}
	for _, operation := range workflow.Operations {
		appendStep(PlanStep{
			ID: "operation:" + operation.ID, Kind: StepOperation, Engine: EngineLogical,
			Inputs: append([]string(nil), operation.Inputs...), Outputs: []string{operation.ID},
		})
	}
	for _, output := range workflow.Outputs {
		appendStep(PlanStep{
			ID: "project:" + output.ID, Kind: StepProjection, Engine: EngineLogical,
			Inputs: []string{output.Input}, Outputs: []string{output.ID + ":projected"},
		})
		appendStep(PlanStep{
			ID: "write:" + output.ID, Kind: StepWrite, Engine: EngineAdapter,
			Inputs: []string{output.ID + ":projected"}, Outputs: []string{output.ID + ":staged"},
		})
		appendStep(PlanStep{
			ID: "validate:" + output.ID, Kind: StepValidate, Engine: EngineAdapter,
			Inputs: []string{output.ID + ":staged"}, Validations: append([]contract.Validation(nil), output.Validations...),
		})
		appendStep(PlanStep{
			ID: "publish:" + output.ID, Kind: StepPublish, Engine: EngineOrchestrator,
			Inputs: []string{output.ID + ":staged"}, Outputs: []string{output.ID},
		})
	}
	for _, query := range workflow.Queries {
		appendStep(PlanStep{
			ID: "query:" + query.ID, Kind: StepQuery, Engine: EngineLogical,
			Inputs: []string{query.Input}, Outputs: []string{query.ID},
		})
	}
	return Plan{
		Workflow: workflow, Mappings: mappings, Settings: settings,
		OutputSettings: outputSettings, OutputLocations: locations, Steps: steps,
	}
}
