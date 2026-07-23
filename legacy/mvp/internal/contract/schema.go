package contract

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

const workflowSchemaURL = "https://data-toolkit.local/schema/workflow-v1.json"

//go:embed schema/workflow-v1.schema.json
var schemaFiles embed.FS

var (
	compileSchemaOnce sync.Once
	compiledSchema    *jsonschema.Schema
	compiledSchemaErr error
)

// PathError identifies an invalid value using a stable document path.
type PathError struct {
	Path    string
	Message string
}

func (e PathError) Error() string {
	return e.Path + ": " + e.Message
}

// ValidationErrors preserves all independent schema findings in document order.
type ValidationErrors []PathError

func (e ValidationErrors) Error() string {
	parts := make([]string, len(e))
	for index := range e {
		parts[index] = e[index].Error()
	}
	return strings.Join(parts, "; ")
}

func validateWorkflowDocument(payload []byte, format DocumentFormat) error {
	document, err := documentForSchema(payload, format)
	if err != nil {
		return err
	}

	schema, err := workflowSchema()
	if err != nil {
		return fmt.Errorf("compile workflow schema: %w", err)
	}
	if err := schema.Validate(document); err != nil {
		var validationError *jsonschema.ValidationError
		if !errors.As(err, &validationError) {
			return fmt.Errorf("validate workflow schema: %w", err)
		}
		findings := make(ValidationErrors, 0)
		collectSchemaErrors(validationError, &findings)
		return findings
	}
	return nil
}

func workflowSchema() (*jsonschema.Schema, error) {
	compileSchemaOnce.Do(func() {
		contents, err := schemaFiles.ReadFile("schema/workflow-v1.schema.json")
		if err != nil {
			compiledSchemaErr = err
			return
		}
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(contents))
		if err != nil {
			compiledSchemaErr = err
			return
		}
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(workflowSchemaURL, document); err != nil {
			compiledSchemaErr = err
			return
		}
		compiledSchema, compiledSchemaErr = compiler.Compile(workflowSchemaURL)
	})
	return compiledSchema, compiledSchemaErr
}

func documentForSchema(payload []byte, format DocumentFormat) (any, error) {
	switch format {
	case DocumentJSON:
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("decode JSON for schema validation: %w", err)
		}
		return document, nil
	case DocumentYAML:
		decoder := yaml.NewDecoder(bytes.NewReader(payload))
		var document any
		if err := decoder.Decode(&document); err != nil {
			return nil, fmt.Errorf("decode YAML for schema validation: %w", err)
		}
		var trailing yaml.Node
		if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
			if err == nil && len(trailing.Content) > 0 {
				return nil, errors.New("multiple YAML documents are not allowed")
			}
			if err != nil {
				return nil, fmt.Errorf("read trailing YAML content: %w", err)
			}
		}
		jsonValue, err := json.Marshal(document)
		if err != nil {
			return nil, fmt.Errorf("convert YAML for schema validation: %w", err)
		}
		converted, err := jsonschema.UnmarshalJSON(bytes.NewReader(jsonValue))
		if err != nil {
			return nil, fmt.Errorf("normalize YAML for schema validation: %w", err)
		}
		return converted, nil
	default:
		return nil, fmt.Errorf("unsupported document format %q", format)
	}
}

func collectSchemaErrors(validationError *jsonschema.ValidationError, findings *ValidationErrors) {
	if len(validationError.Causes) > 0 {
		for _, cause := range validationError.Causes {
			collectSchemaErrors(cause, findings)
		}
		return
	}

	message := validationError.Error()
	if output := validationError.BasicOutput(); output != nil && output.Error != nil {
		message = output.Error.String()
	}
	*findings = append(*findings, PathError{
		Path:    formatInstancePath(validationError.InstanceLocation),
		Message: message,
	})
}

func formatInstancePath(segments []string) string {
	var path strings.Builder
	path.WriteByte('$')
	for _, segment := range segments {
		if _, err := strconv.Atoi(segment); err == nil {
			path.WriteByte('[')
			path.WriteString(segment)
			path.WriteByte(']')
			continue
		}
		path.WriteByte('.')
		path.WriteString(segment)
	}
	return path.String()
}
