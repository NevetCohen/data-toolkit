package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"data-toolkit/configs"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

const schemaURL = "https://data-toolkit.local/schema/config-v1.json"

type Format string

const (
	FormatYAML Format = "yaml"
	FormatJSON Format = "json"
)

type PathError struct {
	Path    string
	Message string
}

func (value PathError) Error() string {
	return value.Path + ": " + value.Message
}

type ValidationErrors []PathError

func (values ValidationErrors) Error() string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = value.Error()
	}
	return strings.Join(parts, "; ")
}

var (
	compileOnce   sync.Once
	compiled      *jsonschema.Schema
	compiledError error
)

func LoadDefaults() (Document, error) {
	payload, err := configs.Default()
	if err != nil {
		return Document{}, err
	}
	return Decode(payload, FormatYAML)
}

func Decode(payload []byte, format Format) (Document, error) {
	documentForSchema, err := normalizeForSchema(payload, format)
	if err != nil {
		return Document{}, err
	}
	schema, err := configSchema()
	if err != nil {
		return Document{}, fmt.Errorf("compile configuration schema: %w", err)
	}
	if err := schema.Validate(documentForSchema); err != nil {
		return Document{}, schemaErrors(err)
	}

	var document Document
	switch format {
	case FormatJSON:
		decoder := json.NewDecoder(bytes.NewReader(payload))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(&document)
	case FormatYAML:
		decoder := yaml.NewDecoder(bytes.NewReader(payload))
		decoder.KnownFields(true)
		err = decoder.Decode(&document)
	default:
		err = fmt.Errorf("unsupported configuration format %q", format)
	}
	if err != nil {
		return Document{}, fmt.Errorf("decode configuration: %w", err)
	}
	if err := Validate(document); err != nil {
		return Document{}, err
	}
	return document, nil
}

func Validate(document Document) error {
	if document.SchemaVersion != CurrentVersion {
		return PathError{Path: "$.schema_version", Message: fmt.Sprintf("unsupported version %q", document.SchemaVersion)}
	}
	if document.Display.TextEncoding != "UTF-8" {
		return PathError{Path: "$.display.text_encoding", Message: "only UTF-8 is supported"}
	}
	if document.Display.UnicodeNormalization != "NFC" {
		return PathError{Path: "$.display.unicode_normalization", Message: "only NFC is supported"}
	}
	if document.Runtime.MaximumMemoryBytes <= 0 {
		return PathError{Path: "$.runtime.maximum_memory_bytes", Message: "must be greater than zero"}
	}
	if document.Runtime.LockRetryCount < 0 || document.Interfaces.RecentWorkflows < 0 {
		return PathError{Path: "$.runtime", Message: "counts must not be negative"}
	}
	for path, value := range map[string]string{
		"$.runtime.lock_retry_interval":  document.Runtime.LockRetryInterval,
		"$.runtime.lock_timeout":         document.Runtime.LockTimeout,
		"$.runtime.failed_run_retention": document.Runtime.FailedRunRetention,
	} {
		duration, err := time.ParseDuration(value)
		if err != nil || duration < 0 {
			return PathError{Path: path, Message: fmt.Sprintf("invalid non-negative duration %q", value)}
		}
	}
	if err := validateUniqueAliases(document.Aliases); err != nil {
		return err
	}
	if err := validateUniqueSemanticTypes(document.SemanticTypes); err != nil {
		return err
	}
	if err := validateStyles(document.Styles, document.Output.DefaultStyle); err != nil {
		return err
	}
	return nil
}

func configSchema() (*jsonschema.Schema, error) {
	compileOnce.Do(func() {
		payload, err := configs.Schema()
		if err != nil {
			compiledError = err
			return
		}
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(payload))
		if err != nil {
			compiledError = err
			return
		}
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(schemaURL, document); err != nil {
			compiledError = err
			return
		}
		compiled, compiledError = compiler.Compile(schemaURL)
	})
	return compiled, compiledError
}

func normalizeForSchema(payload []byte, format Format) (any, error) {
	switch format {
	case FormatJSON:
		value, err := jsonschema.UnmarshalJSON(bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("decode JSON configuration: %w", err)
		}
		return value, nil
	case FormatYAML:
		var value any
		if err := yaml.Unmarshal(payload, &value); err != nil {
			return nil, fmt.Errorf("decode YAML configuration: %w", err)
		}
		normalized, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("normalize YAML configuration: %w", err)
		}
		value, err = jsonschema.UnmarshalJSON(bytes.NewReader(normalized))
		if err != nil {
			return nil, fmt.Errorf("normalize YAML configuration: %w", err)
		}
		return value, nil
	default:
		return nil, fmt.Errorf("unsupported configuration format %q", format)
	}
}

func schemaErrors(err error) error {
	var validationError *jsonschema.ValidationError
	if !errors.As(err, &validationError) {
		return fmt.Errorf("validate configuration schema: %w", err)
	}
	findings := make(ValidationErrors, 0)
	collectSchemaErrors(validationError, &findings)
	return findings
}

func collectSchemaErrors(value *jsonschema.ValidationError, findings *ValidationErrors) {
	if len(value.Causes) > 0 {
		for _, cause := range value.Causes {
			collectSchemaErrors(cause, findings)
		}
		return
	}
	message := value.Error()
	if output := value.BasicOutput(); output != nil && output.Error != nil {
		message = output.Error.String()
	}
	*findings = append(*findings, PathError{Path: instancePath(value.InstanceLocation), Message: message})
}

func instancePath(segments []string) string {
	var result strings.Builder
	result.WriteByte('$')
	for _, segment := range segments {
		if _, err := strconv.Atoi(segment); err == nil {
			result.WriteByte('[')
			result.WriteString(segment)
			result.WriteByte(']')
		} else {
			result.WriteByte('.')
			result.WriteString(segment)
		}
	}
	return result.String()
}

func validateUniqueAliases(values []Alias) error {
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		key := strings.ToLower(strings.TrimSpace(value.Name))
		if key == "" {
			return PathError{Path: fmt.Sprintf("$.aliases[%d].name", index), Message: "identity is required"}
		}
		if _, exists := seen[key]; exists {
			return PathError{Path: fmt.Sprintf("$.aliases[%d].name", index), Message: fmt.Sprintf("duplicate identity %q", value.Name)}
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateUniqueSemanticTypes(values []SemanticType) error {
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		key := strings.ToLower(strings.TrimSpace(value.Name))
		if key == "" || strings.TrimSpace(value.DataType) == "" {
			return PathError{Path: fmt.Sprintf("$.semantic_types[%d]", index), Message: "name and data_type are required"}
		}
		if _, exists := seen[key]; exists {
			return PathError{Path: fmt.Sprintf("$.semantic_types[%d].name", index), Message: fmt.Sprintf("duplicate identity %q", value.Name)}
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateStyles(values []Style, selected string) error {
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		key := strings.ToLower(strings.TrimSpace(value.Name))
		if key == "" {
			return PathError{Path: fmt.Sprintf("$.styles[%d].name", index), Message: "identity is required"}
		}
		if _, exists := seen[key]; exists {
			return PathError{Path: fmt.Sprintf("$.styles[%d].name", index), Message: fmt.Sprintf("duplicate identity %q", value.Name)}
		}
		seen[key] = struct{}{}
	}
	if _, exists := seen[strings.ToLower(selected)]; !exists {
		return PathError{Path: "$.output.default_style", Message: fmt.Sprintf("unknown style %q", selected)}
	}
	return nil
}
