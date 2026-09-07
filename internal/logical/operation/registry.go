package operation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"data-toolkit/internal/core/table"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

type registration struct {
	executor Executor
	schema   *jsonschema.Schema
}

type Registry struct {
	version string
	mu      sync.RWMutex
	entries map[string]registration
}

func NewRegistry(version string) *Registry {
	return &Registry{version: version, entries: make(map[string]registration)}
}

func (registry *Registry) RegisterConstructor(constructor Constructor) error {
	if constructor == nil {
		return fmt.Errorf("register operation: constructor is nil")
	}
	executor, err := constructor()
	if err != nil {
		return fmt.Errorf("construct operation: %w", err)
	}
	return registry.Register(executor)
}

func (registry *Registry) Register(executor Executor) error {
	if registry == nil {
		return fmt.Errorf("register operation: registry is nil")
	}
	if executor == nil {
		return fmt.Errorf("register operation: executor is nil")
	}
	descriptor := executor.Descriptor()
	if err := validateDescriptor(descriptor, registry.version); err != nil {
		return fmt.Errorf("register operation: %w", err)
	}
	compiled, err := compileParameterSchema(descriptor)
	if err != nil {
		return fmt.Errorf("register operation %q: %w", descriptor.Kind, err)
	}
	key := normalize(descriptor.Kind)
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.entries[key]; exists {
		return fmt.Errorf("register operation: identity %q is already registered", descriptor.Kind)
	}
	registry.entries[key] = registration{executor: executor, schema: compiled}
	return nil
}

func (registry *Registry) Require(kind string) (Executor, error) {
	entry, err := registry.require(kind)
	if err != nil {
		return nil, err
	}
	return entry.executor, nil
}

func (registry *Registry) Descriptors() []Descriptor {
	if registry == nil {
		return nil
	}
	registry.mu.RLock()
	result := make([]Descriptor, 0, len(registry.entries))
	for _, entry := range registry.entries {
		descriptor := entry.executor.Descriptor()
		descriptor.ParameterSchema = append([]byte(nil), descriptor.ParameterSchema...)
		result = append(result, descriptor)
	}
	registry.mu.RUnlock()
	sort.Slice(result, func(i, j int) bool { return result[i].Kind < result[j].Kind })
	return result
}

func (registry *Registry) Validate(spec Spec, inputs []table.Schema) error {
	entry, err := registry.require(spec.Kind)
	if err != nil {
		return err
	}
	descriptor := entry.executor.Descriptor()
	if strings.TrimSpace(spec.ID) == "" {
		return fmt.Errorf("operation.id is required")
	}
	if len(spec.Inputs) < descriptor.MinimumInputs {
		return fmt.Errorf("operation %q requires at least %d input(s)", spec.Kind, descriptor.MinimumInputs)
	}
	if descriptor.MaximumInputs != 0 && len(spec.Inputs) > descriptor.MaximumInputs {
		return fmt.Errorf("operation %q accepts at most %d input(s)", spec.Kind, descriptor.MaximumInputs)
	}
	if len(inputs) != len(spec.Inputs) {
		return fmt.Errorf("operation %q resolved %d schemas for %d input identities", spec.ID, len(inputs), len(spec.Inputs))
	}
	parameters := spec.Parameters
	if len(parameters) == 0 {
		parameters = []byte(`{}`)
	}
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(parameters))
	if err != nil {
		return fmt.Errorf("operation %q parameters are invalid JSON: %w", spec.ID, err)
	}
	if err := entry.schema.Validate(document); err != nil {
		return fmt.Errorf("operation %q parameters: %w", spec.ID, err)
	}
	if err := entry.executor.Validate(cloneSpec(spec), append([]table.Schema(nil), inputs...)); err != nil {
		return fmt.Errorf("operation %q validation: %w", spec.ID, err)
	}
	return nil
}

func (registry *Registry) Execute(ctx context.Context, request Request) (Result, error) {
	schemas := make([]table.Schema, len(request.Inputs))
	for index, input := range request.Inputs {
		if err := input.Validate(); err != nil {
			return Result{}, fmt.Errorf("operation %q input[%d]: %w", request.Spec.ID, index, err)
		}
		schemas[index] = input.Schema
	}
	if err := registry.Validate(request.Spec, schemas); err != nil {
		return Result{}, err
	}
	entry, err := registry.require(request.Spec.Kind)
	if err != nil {
		return Result{}, err
	}
	result, err := entry.executor.Execute(ctx, Request{Spec: cloneSpec(request.Spec), Inputs: append([]table.Dataset(nil), request.Inputs...)})
	if err != nil {
		return Result{}, fmt.Errorf("execute operation %q: %w", request.Spec.ID, err)
	}
	if err := result.Dataset.Validate(); err != nil {
		return Result{}, fmt.Errorf("operation %q returned invalid dataset: %w", request.Spec.ID, err)
	}
	return result, nil
}

func (registry *Registry) require(kind string) (registration, error) {
	if registry == nil {
		return registration{}, fmt.Errorf("operation registry is nil")
	}
	key := normalize(kind)
	registry.mu.RLock()
	entry, exists := registry.entries[key]
	registry.mu.RUnlock()
	if !exists {
		return registration{}, fmt.Errorf("operation %q is not registered", kind)
	}
	return entry, nil
}

func validateDescriptor(descriptor Descriptor, version string) error {
	if strings.TrimSpace(descriptor.Kind) == "" || strings.TrimSpace(descriptor.Name) == "" {
		return fmt.Errorf("kind and name are required")
	}
	if descriptor.Version != version {
		return fmt.Errorf("version %q is incompatible with registry version %q", descriptor.Version, version)
	}
	if !descriptor.Deterministic {
		return fmt.Errorf("operation %q must declare deterministic execution", descriptor.Kind)
	}
	if descriptor.MinimumInputs < 0 {
		return fmt.Errorf("operation %q minimum inputs must not be negative", descriptor.Kind)
	}
	if descriptor.MaximumInputs != 0 && descriptor.MaximumInputs < descriptor.MinimumInputs {
		return fmt.Errorf("operation %q maximum inputs is less than minimum", descriptor.Kind)
	}
	if len(descriptor.ParameterSchema) == 0 {
		return fmt.Errorf("operation %q parameter schema is required", descriptor.Kind)
	}
	return nil
}

func compileParameterSchema(descriptor Descriptor) (*jsonschema.Schema, error) {
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(descriptor.ParameterSchema))
	if err != nil {
		return nil, fmt.Errorf("decode parameter schema: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	resource := "https://data-toolkit.local/operation/" + normalize(descriptor.Kind) + "/" + descriptor.Version
	if err := compiler.AddResource(resource, document); err != nil {
		return nil, fmt.Errorf("add parameter schema: %w", err)
	}
	schema, err := compiler.Compile(resource)
	if err != nil {
		return nil, fmt.Errorf("compile parameter schema: %w", err)
	}
	return schema, nil
}

func cloneSpec(source Spec) Spec {
	result := source
	result.Inputs = append([]string(nil), source.Inputs...)
	result.Parameters = append([]byte(nil), source.Parameters...)
	if source.Overrides != nil {
		result.Overrides = make(map[string]json.RawMessage, len(source.Overrides))
		for key, value := range source.Overrides {
			result.Overrides[key] = append([]byte(nil), value...)
		}
	}
	return result
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
