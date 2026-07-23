package contract_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/datatype"
	"data-toolkit/internal/core/table"
	fileformat "data-toolkit/internal/fileengine/format"
	logicaloperation "data-toolkit/internal/logical/operation"
)

type upperType struct {
	version string
	id      datatype.ID
}

func (value upperType) Descriptor() datatype.Descriptor {
	identity := value.id
	if identity == "" {
		identity = "upper"
	}
	return datatype.Descriptor{ID: identity, Version: value.version, Name: "Uppercase text"}
}

func (value upperType) Parse(input string) (datatype.Value, error) {
	return datatype.NewValue("upper", []byte(strings.ToUpper(input)))
}

func (value upperType) Validate(input datatype.Value) error {
	if input.TypeID() != "upper" || string(input.Encoded()) != strings.ToUpper(string(input.Encoded())) {
		return errors.New("value is not canonical uppercase text")
	}
	return nil
}

func (upperType) Compare(left, right datatype.Value) (int, error) {
	return strings.Compare(string(left.Encoded()), string(right.Encoded())), nil
}

func (upperType) Render(value datatype.Value, _ datatype.RenderOptions) (string, error) {
	return string(value.Encoded()), nil
}

type identityOperation struct {
	version  string
	executed *bool
}

func (value identityOperation) Descriptor() logicaloperation.Descriptor {
	return logicaloperation.Descriptor{
		Kind:          "test.identity",
		Version:       value.version,
		Name:          "Test identity",
		Deterministic: true,
		MinimumInputs: 1,
		MaximumInputs: 1,
		ParameterSchema: json.RawMessage(`{
			"type":"object",
			"additionalProperties":false,
			"required":["label"],
			"properties":{"label":{"type":"string","minLength":1}}
		}`),
	}
}

func (identityOperation) Validate(_ logicaloperation.Spec, inputs []table.Schema) error {
	if len(inputs) != 1 {
		return errors.New("one input is required")
	}
	return nil
}

func (value identityOperation) Execute(_ context.Context, request logicaloperation.Request) (logicaloperation.Result, error) {
	if value.executed != nil {
		*value.executed = true
	}
	return logicaloperation.Result{Dataset: request.Inputs[0]}, nil
}

type mappingFormat struct {
	called bool
}

func (provider *mappingFormat) Descriptor() fileformat.Descriptor {
	return fileformat.Descriptor{
		ID:           "test-map",
		Version:      fileformat.CurrentVersion,
		Name:         "Test mapper",
		Extensions:   []string{".test"},
		Capabilities: []fileformat.Capability{fileformat.CapabilityMap},
	}
}

func (provider *mappingFormat) Map(_ context.Context, request fileformat.MapRequest) (fileformat.Mapping, error) {
	provider.called = true
	return fileformat.Mapping{
		SourceID: request.Source.ID,
		Tables: []table.Schema{{
			ID:       "mapped",
			SourceID: request.Source.ID,
		}},
	}, nil
}

type descriptorOnlyFormat struct{}

func (descriptorOnlyFormat) Descriptor() fileformat.Descriptor {
	return fileformat.Descriptor{
		ID:           "invalid",
		Version:      fileformat.CurrentVersion,
		Name:         "Invalid provider",
		Capabilities: []fileformat.Capability{fileformat.CapabilityMap},
	}
}

func TestExternalDataTypeRequiresRegistrationOnly(t *testing.T) {
	registry := datatype.NewRegistry(datatype.CurrentVersion)
	handler := upperType{version: datatype.CurrentVersion}
	if err := registry.RegisterConstructor(func() (datatype.Handler, error) { return handler, nil }); err != nil {
		t.Fatalf("register custom data type: %v", err)
	}
	value, err := registry.Parse("upper", "Abc")
	if err != nil {
		t.Fatalf("parse custom data type: %v", err)
	}
	if got := string(value.Encoded()); got != "ABC" {
		t.Fatalf("canonical value = %q, want ABC", got)
	}
	encoded := value.Encoded()
	encoded[0] = 'X'
	if got := string(value.Encoded()); got != "ABC" {
		t.Fatalf("value was mutated through accessor: %q", got)
	}
	if err := registry.Register(handler); err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("duplicate registration error = %v", err)
	}
	if err := registry.Register(upperType{version: datatype.CurrentVersion, id: "UPPER"}); err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("normalized duplicate registration error = %v", err)
	}
	if err := datatype.NewRegistry(datatype.CurrentVersion).Register(upperType{version: "v2"}); err == nil || !strings.Contains(err.Error(), "incompatible") {
		t.Fatalf("unsupported version error = %v", err)
	}
}

func TestExternalLogicalOperationValidatesBeforeExecution(t *testing.T) {
	registry := logicaloperation.NewRegistry(logicaloperation.CurrentVersion)
	executed := false
	executor := identityOperation{version: logicaloperation.CurrentVersion, executed: &executed}
	if err := registry.RegisterConstructor(func() (logicaloperation.Executor, error) { return executor, nil }); err != nil {
		t.Fatalf("register operation: %v", err)
	}
	input := table.Dataset{
		Schema: table.Schema{ID: "input", SourceID: "source"},
		Rows:   table.NewSliceStream(nil),
	}
	spec := logicaloperation.Spec{
		ID:         "copy",
		Kind:       "test.identity",
		Inputs:     []string{"input"},
		Parameters: json.RawMessage(`{"label":"copy"}`),
	}
	result, err := registry.Execute(context.Background(), logicaloperation.Request{Spec: spec, Inputs: []table.Dataset{input}})
	if err != nil {
		t.Fatalf("execute registered operation: %v", err)
	}
	if result.Dataset.Schema.ID != input.Schema.ID {
		t.Fatalf("result schema = %q, want %q", result.Dataset.Schema.ID, input.Schema.ID)
	}

	spec.Parameters = json.RawMessage(`{"unexpected":true}`)
	executed = false
	if _, err := registry.Execute(context.Background(), logicaloperation.Request{Spec: spec, Inputs: []table.Dataset{input}}); err == nil || !strings.Contains(err.Error(), "parameters") {
		t.Fatalf("invalid parameter error = %v", err)
	}
	if executed {
		t.Fatal("executor was invoked for invalid parameters")
	}
	if err := registry.Register(executor); err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("duplicate operation error = %v", err)
	}
	if err := logicaloperation.NewRegistry(logicaloperation.CurrentVersion).Register(identityOperation{version: "v2"}); err == nil || !strings.Contains(err.Error(), "incompatible") {
		t.Fatalf("unsupported operation version error = %v", err)
	}
}

func TestExternalFileFormatRequiresMatchingCapabilities(t *testing.T) {
	registry := fileformat.NewRegistry(fileformat.CurrentVersion)
	provider := &mappingFormat{}
	if err := registry.RegisterConstructor(func() (fileformat.Provider, error) { return provider, nil }); err != nil {
		t.Fatalf("register format: %v", err)
	}
	mapping, err := registry.Map(context.Background(), fileformat.MapRequest{
		FormatID: "test-map",
		Source:   fileformat.Source{ID: core.SourceID("source"), Location: "opaque.test"},
	})
	if err != nil {
		t.Fatalf("map with registered format: %v", err)
	}
	if !provider.called || mapping.SourceID != "source" {
		t.Fatalf("mapper was not invoked correctly: called=%t source=%q", provider.called, mapping.SourceID)
	}
	if _, err := registry.Require("test-map", fileformat.CapabilityRead); err == nil || !strings.Contains(err.Error(), "does not support") {
		t.Fatalf("missing capability error = %v", err)
	}
	if err := registry.Register(provider); err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("duplicate format error = %v", err)
	}
	if err := fileformat.NewRegistry(fileformat.CurrentVersion).Register(descriptorOnlyFormat{}); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("capability mismatch error = %v", err)
	}
}
