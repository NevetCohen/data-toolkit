package config

import (
	"strings"
	"testing"

	"data-toolkit/internal/contract"
)

func TestDuplicateAliasIdentityIsRejected(t *testing.T) {
	document := "{\"schema_version\":\"v1\",\"aliases\":[{\"name\":\"Phones\",\"semantic_type\":\"phone\"},{\"name\":\"phones\",\"semantic_type\":\"phone\"}]}"
	_, err := Decode(strings.NewReader(document), contract.DocumentJSON)
	if err == nil || !strings.Contains(err.Error(), "$.aliases[1].name") {
		t.Fatalf("expected duplicate alias path, got %v", err)
	}
}

func TestNamedStyleMustExistAndBeUnique(t *testing.T) {
	styleName := "missing"
	user := Config{
		SchemaVersion: CurrentSchemaVersion,
		Styles:        []TableStyle{{Name: "plain"}},
		Output:        OutputConfig{DefaultStyle: &styleName},
	}
	if _, err := Resolve(BuiltInDefaults("temporary"), user, contract.WorkflowOverrides{}, nil); err == nil {
		t.Fatal("expected missing default style to be rejected")
	}

	user.Styles = []TableStyle{{Name: "plain"}, {Name: "PLAIN"}}
	if err := ValidateConfig(user); err == nil {
		t.Fatal("expected duplicate named style to be rejected")
	}
}

func TestOutputOverrideCannotSelectUnknownStyle(t *testing.T) {
	output := contract.OutputOverrides{Style: "unknown"}
	if _, err := Resolve(BuiltInDefaults("temporary"), Config{SchemaVersion: CurrentSchemaVersion}, contract.WorkflowOverrides{}, &output); err == nil {
		t.Fatal("expected unknown output style to be rejected")
	}
}
