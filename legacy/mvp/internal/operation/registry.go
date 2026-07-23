package operation

import (
	"fmt"
	"sort"

	"data-toolkit/internal/contract"
)

type Shape string

const (
	ShapeRows       Shape = "rows"
	ShapePartitions Shape = "partitions"
)

type Parameter string

const (
	ParameterCondition   Parameter = "condition"
	ParameterRename      Parameter = "rename"
	ParameterColumns     Parameter = "columns"
	ParameterNormalize   Parameter = "normalize"
	ParameterDeduplicate Parameter = "deduplicate"
	ParameterSort        Parameter = "sort"
	ParameterSplit       Parameter = "split"
	ParameterText        Parameter = "text"
	ParameterRegex       Parameter = "regex"
	ParameterOverrides   Parameter = "overrides"
)

type ParameterRule struct {
	Name     Parameter
	Required bool
}

// Declaration is the versioned, inspectable contract for one deterministic
// logical operation. MaxInputs == 0 means that no upper bound is declared.
type Declaration struct {
	SchemaVersion string
	Kind          contract.OperationKind
	Deterministic bool
	MinInputs     int
	MaxInputs     int
	InputShape    Shape
	OutputShape   Shape
	Parameters    []ParameterRule
}

type Registry struct {
	declarations map[string]map[contract.OperationKind]Declaration
}

func NewRegistry() *Registry {
	return &Registry{declarations: make(map[string]map[contract.OperationKind]Declaration)}
}

func NewDefaultRegistry() *Registry {
	registry := NewRegistry()
	for _, declaration := range defaultDeclarations() {
		if err := registry.Register(declaration); err != nil {
			panic(fmt.Sprintf("register built-in operation: %v", err))
		}
	}
	return registry
}

func (registry *Registry) Register(declaration Declaration) error {
	if registry == nil {
		return fmt.Errorf("register operation: registry is nil")
	}
	if declaration.SchemaVersion == "" {
		return fmt.Errorf("register operation: schema version is required")
	}
	if declaration.Kind == "" {
		return fmt.Errorf("register operation: kind is required")
	}
	if !declaration.Deterministic {
		return fmt.Errorf("register operation %q: deterministic declaration is required", declaration.Kind)
	}
	if declaration.MinInputs < 1 {
		return fmt.Errorf("register operation %q: minimum inputs must be at least one", declaration.Kind)
	}
	if declaration.MaxInputs != 0 && declaration.MaxInputs < declaration.MinInputs {
		return fmt.Errorf("register operation %q: maximum inputs is less than minimum inputs", declaration.Kind)
	}
	if declaration.InputShape == "" || declaration.OutputShape == "" {
		return fmt.Errorf("register operation %q: input and output shapes are required", declaration.Kind)
	}
	seen := make(map[Parameter]struct{}, len(declaration.Parameters))
	for _, rule := range declaration.Parameters {
		if !knownParameter(rule.Name) {
			return fmt.Errorf("register operation %q: unknown parameter %q", declaration.Kind, rule.Name)
		}
		if _, exists := seen[rule.Name]; exists {
			return fmt.Errorf("register operation %q: duplicate parameter %q", declaration.Kind, rule.Name)
		}
		seen[rule.Name] = struct{}{}
	}
	byKind := registry.declarations[declaration.SchemaVersion]
	if byKind == nil {
		byKind = make(map[contract.OperationKind]Declaration)
		registry.declarations[declaration.SchemaVersion] = byKind
	}
	if _, exists := byKind[declaration.Kind]; exists {
		return fmt.Errorf("register operation %q: already registered for schema version %q", declaration.Kind, declaration.SchemaVersion)
	}
	byKind[declaration.Kind] = cloneDeclaration(declaration)
	return nil
}

func (registry *Registry) Require(schemaVersion string, kind contract.OperationKind) (Declaration, error) {
	if registry == nil {
		return Declaration{}, fmt.Errorf("operation registry is nil")
	}
	byKind, exists := registry.declarations[schemaVersion]
	if !exists {
		return Declaration{}, fmt.Errorf("schema version %q is not registered", schemaVersion)
	}
	declaration, exists := byKind[kind]
	if !exists {
		return Declaration{}, fmt.Errorf("operation %q is not registered for schema version %q", kind, schemaVersion)
	}
	return cloneDeclaration(declaration), nil
}

func (registry *Registry) Declarations(schemaVersion string) []Declaration {
	byKind := registry.declarations[schemaVersion]
	result := make([]Declaration, 0, len(byKind))
	for _, declaration := range byKind {
		result = append(result, cloneDeclaration(declaration))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Kind < result[j].Kind })
	return result
}

// ValidateWorkflow rejects operations and populated parameter fields that are
// not declared by the selected workflow schema version.
func (registry *Registry) ValidateWorkflow(workflow contract.Workflow) error {
	findings := make(contract.ValidationErrors, 0)
	if registry == nil || registry.declarations[workflow.SchemaVersion] == nil {
		findings = append(findings, contract.PathError{Path: "$.schema_version", Message: fmt.Sprintf("operation registry does not support schema version %q", workflow.SchemaVersion)})
		return findings
	}
	for index, selected := range workflow.Operations {
		path := fmt.Sprintf("$.operations[%d]", index)
		declaration, err := registry.Require(workflow.SchemaVersion, selected.Kind)
		if err != nil {
			findings = append(findings, contract.PathError{Path: path + ".kind", Message: err.Error()})
			continue
		}
		if len(selected.Inputs) < declaration.MinInputs {
			findings = append(findings, contract.PathError{Path: path + ".inputs", Message: fmt.Sprintf("operation %q requires at least %d input(s)", selected.Kind, declaration.MinInputs)})
		}
		if declaration.MaxInputs != 0 && len(selected.Inputs) > declaration.MaxInputs {
			findings = append(findings, contract.PathError{Path: path + ".inputs", Message: fmt.Sprintf("operation %q accepts at most %d input(s)", selected.Kind, declaration.MaxInputs)})
		}
		present := presentParameters(selected)
		declared := make(map[Parameter]ParameterRule, len(declaration.Parameters))
		for _, rule := range declaration.Parameters {
			declared[rule.Name] = rule
			if rule.Required && !present[rule.Name] {
				findings = append(findings, contract.PathError{Path: path + "." + string(rule.Name), Message: fmt.Sprintf("parameter %q is required for operation %q", rule.Name, selected.Kind)})
			}
		}
		for _, parameter := range allParameters() {
			if !present[parameter] {
				continue
			}
			if _, allowed := declared[parameter]; !allowed {
				findings = append(findings, contract.PathError{Path: path + "." + string(parameter), Message: fmt.Sprintf("parameter %q is not declared for operation %q in schema version %q", parameter, selected.Kind, workflow.SchemaVersion)})
			}
		}
	}
	if len(findings) != 0 {
		return findings
	}
	return nil
}

func defaultDeclarations() []Declaration {
	required := func(name Parameter) ParameterRule { return ParameterRule{Name: name, Required: true} }
	optional := func(name Parameter) ParameterRule { return ParameterRule{Name: name} }
	declare := func(kind contract.OperationKind, maxInputs int, output Shape, parameters ...ParameterRule) Declaration {
		parameters = append(parameters, optional(ParameterOverrides))
		return Declaration{SchemaVersion: contract.CurrentSchemaVersion, Kind: kind, Deterministic: true, MinInputs: 1, MaxInputs: maxInputs, InputShape: ShapeRows, OutputShape: output, Parameters: parameters}
	}
	return []Declaration{
		declare(contract.OperationFilter, 2, ShapeRows, required(ParameterCondition)),
		declare(contract.OperationRename, 1, ShapeRows, required(ParameterRename)),
		declare(contract.OperationNormalize, 1, ShapeRows, required(ParameterColumns), required(ParameterNormalize)),
		declare(contract.OperationTrim, 1, ShapeRows, required(ParameterColumns)),
		declare(contract.OperationDeduplicate, 1, ShapeRows, required(ParameterDeduplicate)),
		declare(contract.OperationDeleteEmptyRows, 1, ShapeRows),
		declare(contract.OperationDeleteRows, 2, ShapeRows, required(ParameterCondition)),
		declare(contract.OperationSort, 1, ShapeRows, required(ParameterSort)),
		declare(contract.OperationSplit, 1, ShapePartitions, required(ParameterSplit)),
		declare(contract.OperationConcatenate, 1, ShapeRows, required(ParameterText)),
		declare(contract.OperationTransformText, 1, ShapeRows, required(ParameterText)),
		declare(contract.OperationRegexTransform, 1, ShapeRows, required(ParameterRegex)),
	}
}

func presentParameters(selected contract.Operation) map[Parameter]bool {
	return map[Parameter]bool{
		ParameterCondition:   selected.Condition != nil,
		ParameterRename:      len(selected.Rename) != 0,
		ParameterColumns:     len(selected.Columns) != 0,
		ParameterNormalize:   selected.Normalize != nil,
		ParameterDeduplicate: selected.Deduplicate != nil,
		ParameterSort:        len(selected.Sort) != 0,
		ParameterSplit:       selected.Split != nil,
		ParameterText:        selected.Text != nil,
		ParameterRegex:       selected.Regex != nil,
		ParameterOverrides:   selected.Overrides.MixedTypeAction != "",
	}
}

func allParameters() []Parameter {
	return []Parameter{ParameterCondition, ParameterRename, ParameterColumns, ParameterNormalize, ParameterDeduplicate, ParameterSort, ParameterSplit, ParameterText, ParameterRegex, ParameterOverrides}
}

func knownParameter(parameter Parameter) bool {
	for _, known := range allParameters() {
		if parameter == known {
			return true
		}
	}
	return false
}

func cloneDeclaration(declaration Declaration) Declaration {
	declaration.Parameters = append([]ParameterRule(nil), declaration.Parameters...)
	return declaration
}
