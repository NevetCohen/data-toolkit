package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"data-toolkit/internal/application"
	"data-toolkit/internal/config"
	"data-toolkit/internal/contract"
	"data-toolkit/internal/orchestrator"

	"github.com/spf13/cobra"
)

const (
	ExitSuccess = 0
	ExitFailure = 1
	ExitUsage   = 2
)

type usageError struct {
	err error
}

func (err *usageError) Error() string {
	return err.err.Error()
}

func (err *usageError) Unwrap() error {
	return err.err
}

type commandRuntime struct {
	stdout io.Writer
	stderr io.Writer
}

func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	root := NewRootCommand(stdout, stderr)
	root.SetArgs(args)
	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		if isUsageError(err) {
			return ExitUsage
		}
		return ExitFailure
	}
	return ExitSuccess
}

func NewRootCommand(stdout, stderr io.Writer) *cobra.Command {
	runtime := commandRuntime{stdout: stdout, stderr: stderr}
	root := &cobra.Command{
		Use:           "data-toolkit",
		Short:         "Run deterministic local data workflows",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return &usageError{err: err}
	})
	root.AddCommand(runtime.newWorkflowCommand())
	return root
}

func (runtime commandRuntime) newWorkflowCommand() *cobra.Command {
	workflow := &cobra.Command{
		Use:   "workflow",
		Short: "Validate or run a workflow document",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return &usageError{err: fmt.Errorf("unknown workflow command %q", args[0])}
			}
			return &usageError{err: fmt.Errorf("%s requires a subcommand", cmd.CommandPath())}
		},
	}
	workflow.AddCommand(runtime.newValidateCommand())
	workflow.AddCommand(runtime.newRunCommand())
	return workflow
}

func (runtime commandRuntime) newValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <workflow>",
		Short: "Validate workflow schema and registered operations without reading sources",
		Args:  exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflow, err := loadWorkflowFile(args[0])
			if err != nil {
				return err
			}
			service, err := application.NewLocalService(application.LocalOptions{})
			if err != nil {
				return fmt.Errorf("create local application service: %w", err)
			}
			if err := service.ValidateWorkflow(workflow); err != nil {
				return fmt.Errorf("validate workflow: %w", err)
			}
			return writeJSON(runtime.stdout, struct {
				SchemaVersion string `json:"schema_version"`
				WorkflowID    string `json:"workflow_id"`
				Status        string `json:"status"`
			}{
				SchemaVersion: contract.CurrentSchemaVersion,
				WorkflowID:    workflow.ID,
				Status:        "valid",
			})
		},
	}
}

func (runtime commandRuntime) newRunCommand() *cobra.Command {
	var configPath string
	run := &cobra.Command{
		Use:   "run <workflow>",
		Short: "Run a workflow and emit its final result as JSON",
		Args:  exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflow, err := loadWorkflowFile(args[0])
			if err != nil {
				return err
			}
			userConfig := config.Config{SchemaVersion: config.CurrentSchemaVersion}
			if configPath != "" {
				userConfig, err = loadConfigFile(configPath)
				if err != nil {
					return err
				}
			}
			observer := orchestrator.ObserverFunc(func(_ context.Context, started contract.StartedResult) error {
				return writeJSON(runtime.stderr, started)
			})
			service, err := application.NewLocalService(application.LocalOptions{
				UserConfig: userConfig,
				Observer:   observer,
			})
			if err != nil {
				return fmt.Errorf("create local application service: %w", err)
			}
			final, runErr := service.Run(cmd.Context(), contract.RunRequest{Workflow: workflow})
			if final.Status != "" {
				if err := writeJSON(runtime.stdout, final); err != nil {
					return err
				}
			}
			if runErr != nil {
				return fmt.Errorf("run workflow: %w", runErr)
			}
			return nil
		},
	}
	run.Flags().StringVar(&configPath, "config", "", "path to a YAML or JSON user configuration")
	return run
}

func loadWorkflowFile(path string) (contract.Workflow, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return contract.Workflow{}, fmt.Errorf("resolve workflow path: %w", err)
	}
	format, err := documentFormat(absolute)
	if err != nil {
		return contract.Workflow{}, err
	}
	file, err := os.Open(absolute)
	if err != nil {
		return contract.Workflow{}, fmt.Errorf("open workflow %q: %w", absolute, err)
	}
	defer file.Close()
	workflow, err := contract.DecodeWorkflow(file, format)
	if err != nil {
		return contract.Workflow{}, fmt.Errorf("decode workflow %q: %w", absolute, err)
	}
	resolveWorkflowPaths(&workflow, filepath.Dir(absolute))
	return workflow, nil
}

func loadConfigFile(path string) (config.Config, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return config.Config{}, fmt.Errorf("resolve configuration path: %w", err)
	}
	format, err := documentFormat(absolute)
	if err != nil {
		return config.Config{}, err
	}
	file, err := os.Open(absolute)
	if err != nil {
		return config.Config{}, fmt.Errorf("open configuration %q: %w", absolute, err)
	}
	defer file.Close()
	configuration, err := config.Decode(file, format)
	if err != nil {
		return config.Config{}, fmt.Errorf("decode configuration %q: %w", absolute, err)
	}
	if configuration.Output.Directory != nil && !filepath.IsAbs(*configuration.Output.Directory) {
		resolved := filepath.Clean(filepath.Join(filepath.Dir(absolute), *configuration.Output.Directory))
		configuration.Output.Directory = &resolved
	}
	return configuration, nil
}

func documentFormat(path string) (contract.DocumentFormat, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return contract.DocumentJSON, nil
	case ".yaml", ".yml":
		return contract.DocumentYAML, nil
	default:
		return "", fmt.Errorf("document %q must use a .json, .yaml, or .yml extension", path)
	}
}

func resolveWorkflowPaths(workflow *contract.Workflow, baseDirectory string) {
	for index := range workflow.Sources {
		source := &workflow.Sources[index]
		if source.Format != contract.FormatGoogleSheets {
			source.Location = resolveRelativePath(baseDirectory, source.Location)
		}
	}
	if workflow.Overrides.OutputDirectory != "" {
		workflow.Overrides.OutputDirectory = resolveRelativePath(baseDirectory, workflow.Overrides.OutputDirectory)
	}
	for index := range workflow.Outputs {
		output := &workflow.Outputs[index]
		if output.Format == contract.FormatGoogleSheets {
			continue
		}
		if output.Location != "" {
			output.Location = resolveRelativePath(baseDirectory, output.Location)
		}
		if output.LocationPolicy != nil && output.LocationPolicy.Directory != "" {
			output.LocationPolicy.Directory = resolveRelativePath(baseDirectory, output.LocationPolicy.Directory)
		}
		if output.Overrides.Directory != "" {
			output.Overrides.Directory = resolveRelativePath(baseDirectory, output.Overrides.Directory)
		}
	}
}

func resolveRelativePath(baseDirectory, value string) string {
	if value == "" || filepath.IsAbs(value) {
		return value
	}
	return filepath.Clean(filepath.Join(baseDirectory, value))
}

func writeJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("write JSON output: %w", err)
	}
	return nil
}

func exactArgs(count int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != count {
			return &usageError{err: fmt.Errorf("%s requires exactly %d argument(s)", cmd.CommandPath(), count)}
		}
		return nil
	}
}

func isUsageError(err error) bool {
	var usage *usageError
	if errors.As(err, &usage) {
		return true
	}
	message := err.Error()
	return strings.HasPrefix(message, "unknown command") ||
		strings.HasPrefix(message, "unknown flag") ||
		strings.Contains(message, "requires a subcommand")
}
