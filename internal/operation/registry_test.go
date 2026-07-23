package operation

import (
	"errors"
	"testing"

	"data-toolkit/internal/contract"
)

func TestDefaultRegistryDeclaresClosedVersionedDeterministicOperations(t *testing.T) {
	registry := NewDefaultRegistry()
	declarations := registry.Declarations(contract.CurrentSchemaVersion)
	if len(declarations) != 12 {
		t.Fatalf("declarations = %d, want 12", len(declarations))
	}
	for _, declaration := range declarations {
		if declaration.SchemaVersion != contract.CurrentSchemaVersion || !declaration.Deterministic || declaration.InputShape == "" || declaration.OutputShape == "" || declaration.MinInputs < 1 {
			t.Fatalf("incomplete declaration: %#v", declaration)
		}
	}
	if _, err := registry.Require(contract.CurrentSchemaVersion, contract.OperationFilter); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Require(contract.CurrentSchemaVersion, contract.OperationKind("mystery")); err == nil {
		t.Fatal("expected unregistered operation error")
	}
	if _, err := registry.Require("v9", contract.OperationFilter); err == nil {
		t.Fatal("expected unregistered version error")
	}
}

func TestRegistryRejectsUndeclaredOperationParameter(t *testing.T) {
	workflow := contract.Workflow{SchemaVersion: contract.CurrentSchemaVersion, Operations: []contract.Operation{{
		ID: "rename", Kind: contract.OperationRename, Inputs: []string{"source"},
		Rename: []contract.RenameField{{Column: contract.ColumnRef{Header: "name"}, To: "full_name"}},
		Sort:   []contract.SortField{{Column: contract.ColumnRef{Header: "name"}}},
	}}}
	err := NewDefaultRegistry().ValidateWorkflow(workflow)
	var findings contract.ValidationErrors
	if !errors.As(err, &findings) {
		t.Fatalf("error type = %T, want ValidationErrors: %v", err, err)
	}
	if len(findings) != 1 || findings[0].Path != "$.operations[0].sort" {
		t.Fatalf("findings = %#v", findings)
	}
}

func TestRegisterRejectsNonDeterministicAndDuplicateDeclarations(t *testing.T) {
	registry := NewRegistry()
	declaration := Declaration{SchemaVersion: "v1", Kind: contract.OperationFilter, Deterministic: false, MinInputs: 1, InputShape: ShapeRows, OutputShape: ShapeRows}
	if err := registry.Register(declaration); err == nil {
		t.Fatal("expected non-deterministic declaration rejection")
	}
	declaration.Deterministic = true
	if err := registry.Register(declaration); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(declaration); err == nil {
		t.Fatal("expected duplicate declaration rejection")
	}
}
