// Package application composes the local Data Toolkit runtime and exposes the
// interface-neutral application service used by terminal and future UI layers.
package application

import (
	"context"
	"errors"
	"os"
	"time"

	"data-toolkit/internal/adapter"
	"data-toolkit/internal/config"
	"data-toolkit/internal/contract"
	"data-toolkit/internal/operation"
	"data-toolkit/internal/orchestrator"
)

type LocalOptions struct {
	TemporaryRoot string
	UserConfig    config.Config
	Observer      orchestrator.Observer
	Now           func() time.Time
	RunID         func() string
}

type LocalService struct {
	orchestrator *orchestrator.Service
	operations   *operation.Registry
}

// NewLocalService is the composition root for the local executable. It
// registers only adapters and operations that are currently implemented.
func NewLocalService(options LocalOptions) (*LocalService, error) {
	temporaryRoot := options.TemporaryRoot
	if temporaryRoot == "" {
		temporaryRoot = os.TempDir()
	}
	userConfig := options.UserConfig
	if userConfig.SchemaVersion == "" {
		userConfig.SchemaVersion = config.CurrentSchemaVersion
	}
	if err := config.ValidateConfig(userConfig); err != nil {
		return nil, err
	}

	adapters := adapter.NewRegistry()
	for _, selected := range []adapter.Adapter{
		adapter.NewJSONAdapter(),
		adapter.NewCSVAdapter(),
		adapter.NewExcelAdapter(),
	} {
		if err := adapters.Register(selected); err != nil {
			return nil, err
		}
	}
	operations := operation.NewDefaultRegistry()
	defaults := config.BuiltInDefaults(temporaryRoot)
	service := &orchestrator.Service{
		Adapters: adapters, Operations: operations,
		Defaults: defaults, UserConfig: userConfig,
		Backend: &orchestrator.LocalPipelineBackend{
			Adapters: adapters,
			Retry: adapter.RetryPolicy{
				RetryCount: defaults.LockRetryCount,
				Interval:   defaults.LockRetryInterval,
				Timeout:    defaults.LockTimeout,
			},
		},
		Observer: options.Observer,
		Now:      options.Now,
		RunID:    options.RunID,
	}
	return &LocalService{orchestrator: service, operations: operations}, nil
}

// ValidateWorkflow performs contract and operation-registry validation only.
// It intentionally does not inspect sources or output targets.
func (service *LocalService) ValidateWorkflow(workflow contract.Workflow) error {
	if service == nil || service.operations == nil {
		return errors.New("local application service is not initialized")
	}
	contract.NormalizeWorkflow(&workflow)
	if err := contract.ValidateWorkflow(workflow); err != nil {
		return err
	}
	return service.operations.ValidateWorkflow(workflow)
}

func (service *LocalService) Run(ctx context.Context, request contract.RunRequest) (contract.FinalResult, error) {
	if service == nil || service.orchestrator == nil {
		return contract.FinalResult{}, errors.New("local application service is not initialized")
	}
	return service.orchestrator.Run(ctx, request)
}
