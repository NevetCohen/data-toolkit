package config

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

	"data-toolkit/internal/contract"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

const configSchemaURL = "https://data-toolkit.local/schema/config-v1.json"

//go:embed schema/config-v1.schema.json
var configSchemaFiles embed.FS

var (
	compileConfigSchemaOnce sync.Once
	compiledConfigSchema    *jsonschema.Schema
	compiledConfigSchemaErr error
)

func Decode(reader io.Reader, format contract.DocumentFormat) (Config, error) {
	payload, err := io.ReadAll(reader)
	if err != nil {
		return Config{}, fmt.Errorf("read configuration: %w", err)
	}
	document, err := configDocumentForSchema(payload, format)
	if err != nil {
		return Config{}, err
	}
	schema, err := configSchema()
	if err != nil {
		return Config{}, fmt.Errorf("compile configuration schema: %w", err)
	}
	if err := schema.Validate(document); err != nil {
		var validationError *jsonschema.ValidationError
		if !errors.As(err, &validationError) {
			return Config{}, fmt.Errorf("validate configuration schema: %w", err)
		}
		findings := make(contract.ValidationErrors, 0)
		collectConfigSchemaErrors(validationError, &findings)
		return Config{}, findings
	}

	var configuration Config
	switch format {
	case contract.DocumentJSON:
		decoder := json.NewDecoder(bytes.NewReader(payload))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(&configuration)
	case contract.DocumentYAML:
		decoder := yaml.NewDecoder(bytes.NewReader(payload))
		decoder.KnownFields(true)
		err = decoder.Decode(&configuration)
	default:
		err = fmt.Errorf("unsupported configuration format %q", format)
	}
	if err != nil {
		return Config{}, fmt.Errorf("decode configuration: %w", err)
	}
	if err := ValidateConfig(configuration); err != nil {
		return Config{}, err
	}
	return configuration, nil
}

func configSchema() (*jsonschema.Schema, error) {
	compileConfigSchemaOnce.Do(func() {
		contents, err := configSchemaFiles.ReadFile("schema/config-v1.schema.json")
		if err != nil {
			compiledConfigSchemaErr = err
			return
		}
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(contents))
		if err != nil {
			compiledConfigSchemaErr = err
			return
		}
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(configSchemaURL, document); err != nil {
			compiledConfigSchemaErr = err
			return
		}
		compiledConfigSchema, compiledConfigSchemaErr = compiler.Compile(configSchemaURL)
	})
	return compiledConfigSchema, compiledConfigSchemaErr
}

func configDocumentForSchema(payload []byte, format contract.DocumentFormat) (any, error) {
	switch format {
	case contract.DocumentJSON:
		value, err := jsonschema.UnmarshalJSON(bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("decode JSON configuration: %w", err)
		}
		return value, nil
	case contract.DocumentYAML:
		var value any
		if err := yaml.Unmarshal(payload, &value); err != nil {
			return nil, fmt.Errorf("decode YAML configuration: %w", err)
		}
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("normalize YAML configuration: %w", err)
		}
		converted, err := jsonschema.UnmarshalJSON(bytes.NewReader(jsonValue))
		if err != nil {
			return nil, fmt.Errorf("normalize YAML configuration: %w", err)
		}
		return converted, nil
	default:
		return nil, fmt.Errorf("unsupported configuration format %q", format)
	}
}

func collectConfigSchemaErrors(validationError *jsonschema.ValidationError, findings *contract.ValidationErrors) {
	if len(validationError.Causes) > 0 {
		for _, cause := range validationError.Causes {
			collectConfigSchemaErrors(cause, findings)
		}
		return
	}
	message := validationError.Error()
	if output := validationError.BasicOutput(); output != nil && output.Error != nil {
		message = output.Error.String()
	}
	*findings = append(*findings, contract.PathError{
		Path:    configInstancePath(validationError.InstanceLocation),
		Message: message,
	})
}

func configInstancePath(segments []string) string {
	var path strings.Builder
	path.WriteByte('$')
	for _, segment := range segments {
		if _, err := strconv.Atoi(segment); err == nil {
			path.WriteByte('[')
			path.WriteString(segment)
			path.WriteByte(']')
		} else {
			path.WriteByte('.')
			path.WriteString(segment)
		}
	}
	return path.String()
}
