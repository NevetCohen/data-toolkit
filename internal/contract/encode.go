package contract

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// EncodeWorkflow writes one validated normalized document deterministically.
func EncodeWorkflow(writer io.Writer, workflow Workflow, format DocumentFormat) error {
	if err := ValidateWorkflow(workflow); err != nil {
		return err
	}

	switch format {
	case DocumentJSON:
		encoder := json.NewEncoder(writer)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(workflow); err != nil {
			return fmt.Errorf("encode JSON workflow: %w", err)
		}
	case DocumentYAML:
		encoder := yaml.NewEncoder(writer)
		encoder.SetIndent(2)
		defer encoder.Close()
		if err := encoder.Encode(workflow); err != nil {
			return fmt.Errorf("encode YAML workflow: %w", err)
		}
	default:
		return fmt.Errorf("unsupported document format %q", format)
	}
	return nil
}
