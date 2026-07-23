package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type DocumentFormat string

const (
	DocumentJSON DocumentFormat = "json"
	DocumentYAML DocumentFormat = "yaml"
)

// DecodeWorkflow decodes exactly one strict YAML or JSON document and returns
// the shared typed representation. Semantic validation is a separate step.
func DecodeWorkflow(reader io.Reader, format DocumentFormat) (Workflow, error) {
	payload, err := io.ReadAll(reader)
	if err != nil {
		return Workflow{}, fmt.Errorf("read %s workflow: %w", format, err)
	}
	if err := validateWorkflowDocument(payload, format); err != nil {
		return Workflow{}, err
	}

	var workflow Workflow

	switch format {
	case DocumentJSON:
		err = decodeJSONDocument(bytes.NewReader(payload), &workflow)
	case DocumentYAML:
		err = decodeYAMLDocument(bytes.NewReader(payload), &workflow)
	default:
		return Workflow{}, fmt.Errorf("unsupported document format %q", format)
	}
	if err != nil {
		return Workflow{}, fmt.Errorf("decode %s workflow: %w", format, err)
	}
	NormalizeWorkflow(&workflow)
	if err := ValidateWorkflow(workflow); err != nil {
		return Workflow{}, err
	}

	return workflow, nil
}

func decodeJSONDocument(reader io.Reader, target any) error {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON documents are not allowed")
		}
		return fmt.Errorf("read trailing JSON content: %w", err)
	}
	return nil
}

func decodeYAMLDocument(reader io.Reader, target any) error {
	decoder := yaml.NewDecoder(reader)
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		return err
	}

	var trailing yaml.Node
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil && len(trailing.Content) == 0 {
			return nil
		}
		if err == nil {
			return errors.New("multiple YAML documents are not allowed")
		}
		return fmt.Errorf("read trailing YAML content: %w", err)
	}
	return nil
}
