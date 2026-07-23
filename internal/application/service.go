// Package application exposes the interface-neutral V1 product API.
package application

import (
	"context"
	"fmt"

	"data-toolkit/internal/config"
	"data-toolkit/internal/core/datatype"
	fileformat "data-toolkit/internal/fileengine/format"
	logicaloperation "data-toolkit/internal/logical/operation"
	"data-toolkit/internal/orchestrator"
)

type Capabilities struct {
	DataTypes         []datatype.Descriptor         `json:"data_types"`
	LogicalOperations []logicaloperation.Descriptor `json:"logical_operations"`
	FileFormats       []fileformat.Descriptor       `json:"file_formats"`
}

type ConfigValidation struct {
	Valid  bool            `json:"valid"`
	Config config.Document `json:"config"`
}

type API interface {
	Capabilities(context.Context) (Capabilities, error)
	ValidateConfig(context.Context, []byte, config.Format) (ConfigValidation, error)
	ValidateWorkflow(context.Context, orchestrator.Workflow) error
	Map(context.Context, fileformat.MapRequest) (fileformat.Mapping, error)
	Plan(context.Context, orchestrator.Workflow) (orchestrator.Plan, error)
	Run(context.Context, orchestrator.RunRequest) (orchestrator.RunResult, error)
}

type Options struct {
	DataTypes  *datatype.Registry
	Operations *logicaloperation.Registry
	Formats    *fileformat.Registry
	Defaults   config.Document
	Observer   orchestrator.Observer
}

type LocalService struct {
	dataTypes    *datatype.Registry
	operations   *logicaloperation.Registry
	formats      *fileformat.Registry
	defaults     config.Document
	orchestrator orchestrator.Service
}

func New(options Options) (*LocalService, error) {
	if options.DataTypes == nil || options.Operations == nil || options.Formats == nil {
		return nil, fmt.Errorf("application service requires data-type, logical-operation, and file-format registries")
	}
	if err := config.Validate(options.Defaults); err != nil {
		return nil, fmt.Errorf("application defaults: %w", err)
	}
	return &LocalService{
		dataTypes:  options.DataTypes,
		operations: options.Operations,
		formats:    options.Formats,
		defaults:   options.Defaults,
		orchestrator: orchestrator.Service{
			Operations: options.Operations,
			Formats:    options.Formats,
			Observer:   options.Observer,
		},
	}, nil
}

func (service *LocalService) Capabilities(ctx context.Context) (Capabilities, error) {
	if err := ctx.Err(); err != nil {
		return Capabilities{}, err
	}
	return Capabilities{
		DataTypes:         service.dataTypes.Descriptors(),
		LogicalOperations: service.operations.Descriptors(),
		FileFormats:       service.formats.Descriptors(),
	}, nil
}

func (service *LocalService) ValidateConfig(ctx context.Context, payload []byte, format config.Format) (ConfigValidation, error) {
	if err := ctx.Err(); err != nil {
		return ConfigValidation{}, err
	}
	document, err := config.Decode(payload, format)
	if err != nil {
		return ConfigValidation{}, err
	}
	return ConfigValidation{Valid: true, Config: document}, nil
}

func (service *LocalService) ValidateWorkflow(ctx context.Context, workflow orchestrator.Workflow) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return service.orchestrator.ValidateWorkflow(workflow)
}

func (service *LocalService) Map(ctx context.Context, request fileformat.MapRequest) (fileformat.Mapping, error) {
	return service.orchestrator.Map(ctx, request)
}

func (service *LocalService) Plan(ctx context.Context, workflow orchestrator.Workflow) (orchestrator.Plan, error) {
	if err := ctx.Err(); err != nil {
		return orchestrator.Plan{}, err
	}
	return service.orchestrator.Plan(workflow)
}

func (service *LocalService) Run(ctx context.Context, request orchestrator.RunRequest) (orchestrator.RunResult, error) {
	return service.orchestrator.Run(ctx, request)
}

func (service *LocalService) Defaults() config.Document {
	resolved, _ := config.Resolve(service.defaults)
	return resolved
}
