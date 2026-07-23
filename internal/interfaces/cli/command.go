// Package cli provides the V1 command-line scaffold over the application API.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"data-toolkit/internal/application"
	"data-toolkit/internal/config"

	"github.com/spf13/cobra"
)

func New(api application.API, output io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:           "data-toolkit",
		Short:         "Deterministic Data Toolkit V1 foundation",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(output)
	root.AddCommand(capabilitiesCommand(api, output), configCommand(api, output))
	return root
}

func capabilitiesCommand(api application.API, output io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "capabilities",
		Short: "List compiled V1 extension capabilities",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			capabilities, err := api.Capabilities(command.Context())
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(output)
			encoder.SetIndent("", "  ")
			return encoder.Encode(capabilities)
		},
	}
}

func configCommand(api application.API, output io.Writer) *cobra.Command {
	command := &cobra.Command{Use: "config", Short: "Work with V1 configuration"}
	command.AddCommand(&cobra.Command{
		Use:   "validate <path>",
		Short: "Validate a complete V1 YAML or JSON configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			payload, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read configuration: %w", err)
			}
			format := config.FormatYAML
			if strings.EqualFold(filepath.Ext(args[0]), ".json") {
				format = config.FormatJSON
			}
			result, err := api.ValidateConfig(command.Context(), payload, format)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(output, "valid=%t schema_version=%s\n", result.Valid, result.Config.SchemaVersion)
			return err
		},
	})
	return command
}

func Execute(ctx context.Context, api application.API, args []string, output io.Writer) error {
	command := New(api, output)
	command.SetArgs(args)
	return command.ExecuteContext(ctx)
}
