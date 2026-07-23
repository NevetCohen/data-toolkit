package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"data-toolkit/internal/adapter"
	"data-toolkit/internal/config"
	"data-toolkit/internal/contract"
	"data-toolkit/internal/operation"
)

type Service struct {
	Adapters   *adapter.Registry
	Operations *operation.Registry
	Defaults   config.Resolved
	UserConfig config.Config
	Now        func() time.Time
	RunID      func() string
	Backend    ExecutionBackend
	Observer   Observer
	Redactor   config.Redactor
}

func (service *Service) Preflight(ctx context.Context, request contract.RunRequest) (Plan, error) {
	workflow := request.Workflow
	contract.NormalizeWorkflow(&workflow)
	if err := contract.ValidateWorkflow(workflow); err != nil {
		return Plan{}, fmt.Errorf("validate workflow contract: %w", err)
	}
	operations := service.Operations
	if operations == nil {
		operations = operation.NewDefaultRegistry()
	}
	if err := operations.ValidateWorkflow(workflow); err != nil {
		return Plan{}, fmt.Errorf("validate operation registry: %w", err)
	}
	settings, err := config.Resolve(service.Defaults, service.UserConfig, workflow.Overrides, nil)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve configuration: %w", err)
	}
	if service.Adapters == nil {
		return Plan{}, errors.New("adapter registry is required")
	}

	mappings := make([]adapter.MappingReport, 0, len(workflow.Sources))
	for _, source := range workflow.Sources {
		selected, err := service.Adapters.Require(
			source.Format,
			adapter.CapabilityInspect,
			adapter.CapabilityMapBasic,
			adapter.CapabilityRead,
		)
		if err != nil {
			return Plan{}, fmt.Errorf("preflight source %q: %w", source.ID, err)
		}
		report, err := selected.Map(ctx, adapter.SourceRequest{
			Source: source, Mappings: sourceMappings(workflow.Mappings, source.ID),
		}, adapter.MappingBasic)
		if err != nil {
			return Plan{}, fmt.Errorf("map source %q: %w", source.ID, err)
		}
		mappings = append(mappings, report)
	}
	if len(mappings) != len(workflow.Sources) {
		return Plan{}, errors.New("all source mappings are required before planning")
	}

	outputSettings := make(map[string]config.Resolved, len(workflow.Outputs))
	locations := make(map[string]string, len(workflow.Outputs))
	sourceIdentities, err := workflowSourceIdentities(workflow)
	if err != nil {
		return Plan{}, err
	}
	outputIdentities := make([]adapter.ResourceIdentity, 0, len(workflow.Outputs))
	now := time.Now().UTC()
	if service.Now != nil {
		now = service.Now().UTC()
	}
	for _, output := range workflow.Outputs {
		selected, err := service.Adapters.Require(output.Format, adapter.CapabilityWrite, adapter.CapabilityValidate)
		if err != nil {
			return Plan{}, fmt.Errorf("preflight output %q: %w", output.ID, err)
		}
		_ = selected
		overrides := output.Overrides
		if output.Style != "" && overrides.Style == "" {
			overrides.Style = output.Style
		}
		resolved, err := config.Resolve(service.Defaults, service.UserConfig, workflow.Overrides, &overrides)
		if err != nil {
			return Plan{}, fmt.Errorf("resolve output %q configuration: %w", output.ID, err)
		}
		resolved.OutputFormat = output.Format
		location, err := resolveOutputLocation(workflow.ID, output, resolved, now)
		if err != nil {
			return Plan{}, fmt.Errorf("resolve output %q: %w", output.ID, err)
		}
		if output.Format == contract.FormatGoogleSheets {
			identity, err := adapter.GoogleIdentity(location)
			if err != nil {
				return Plan{}, err
			}
			outputIdentities = append(outputIdentities, identity)
		} else {
			identity, err := adapter.LocalIdentity(location)
			if err != nil {
				return Plan{}, err
			}
			outputIdentities = append(outputIdentities, identity)
			if resolved.Collision == contract.CollisionBlock {
				if _, err := os.Stat(location); err == nil {
					return Plan{}, fmt.Errorf("output %q target %q exists and collision policy is block", output.ID, location)
				} else if !errors.Is(err, os.ErrNotExist) {
					return Plan{}, fmt.Errorf("inspect output %q target: %w", output.ID, err)
				}
			}
		}
		outputSettings[output.ID] = resolved
		locations[output.ID] = location
	}
	if err := adapter.ValidateDistinctOutputIdentities(sourceIdentities, outputIdentities); err != nil {
		return Plan{}, fmt.Errorf("validate immutable source boundary: %w", err)
	}

	return buildPlan(workflow, mappings, settings, outputSettings, locations), nil
}

func sourceMappings(mappings []contract.FieldMapping, sourceID string) []contract.FieldMapping {
	selected := make([]contract.FieldMapping, 0)
	for _, mapping := range mappings {
		if mapping.SourceID == sourceID {
			selected = append(selected, mapping)
		}
	}
	return selected
}

func workflowSourceIdentities(workflow contract.Workflow) ([]adapter.ResourceIdentity, error) {
	identities := make([]adapter.ResourceIdentity, 0, len(workflow.Sources))
	for index, source := range workflow.Sources {
		var identity adapter.ResourceIdentity
		var err error
		if source.Format == contract.FormatGoogleSheets {
			identity, err = adapter.GoogleIdentity(source.Location)
		} else {
			identity, err = adapter.LocalIdentity(source.Location)
		}
		if err != nil {
			return nil, fmt.Errorf("resolve sources[%d] identity: %w", index, err)
		}
		identities = append(identities, identity)
	}
	return identities, nil
}

func resolveOutputLocation(workflowID string, output contract.Output, settings config.Resolved, now time.Time) (string, error) {
	if output.Location != "" {
		return output.Location, nil
	}
	if output.LocationPolicy == nil {
		return "", errors.New("location or location policy is required")
	}
	directory := settings.OutputDirectory
	template := settings.FilenameTemplate
	if output.LocationPolicy.Directory != "" {
		directory = output.LocationPolicy.Directory
	}
	if output.LocationPolicy.FilenameTemplate != "" {
		template = output.LocationPolicy.FilenameTemplate
	}
	replacer := strings.NewReplacer(
		"{{workflow_id}}", workflowID,
		"{{output_id}}", output.ID,
		"{{timestamp}}", now.Format("20060102-150405"),
	)
	filename := replacer.Replace(template)
	if filename == "" || filename == "." || filename == ".." {
		return "", errors.New("filename template resolved to an invalid filename")
	}
	if filepath.Base(filename) != filename {
		return "", errors.New("filename template must not resolve outside the output directory")
	}
	if filepath.Ext(filename) == "" {
		filename += outputExtension(output.Format)
	}
	return filepath.Join(directory, filename), nil
}

func outputExtension(format contract.DataFormat) string {
	switch format {
	case contract.FormatJSON:
		return ".json"
	case contract.FormatCSV:
		return ".csv"
	case contract.FormatExcel:
		return ".xlsx"
	case contract.FormatTXT:
		return ".txt"
	default:
		return ""
	}
}
