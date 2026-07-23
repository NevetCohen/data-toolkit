package config

import (
	"strings"
	"testing"

	"data-toolkit/internal/contract"
)

func TestDecodeEquivalentYAMLAndJSONConfiguration(t *testing.T) {
	jsonConfig := "{\"schema_version\":\"v1\",\"output\":{\"txt_delimiter\":\",\",\"collision\":\"alternate_name\"}}"
	yamlConfig := "schema_version: v1\noutput:\n  txt_delimiter: ','\n  collision: alternate_name\n"

	fromJSON, err := Decode(strings.NewReader(jsonConfig), contract.DocumentJSON)
	if err != nil {
		t.Fatal(err)
	}
	fromYAML, err := Decode(strings.NewReader(yamlConfig), contract.DocumentYAML)
	if err != nil {
		t.Fatal(err)
	}
	if *fromJSON.Output.TXTDelimiter != *fromYAML.Output.TXTDelimiter ||
		*fromJSON.Output.Collision != *fromYAML.Output.Collision {
		t.Fatalf("equivalent configurations differ: %#v %#v", fromJSON, fromYAML)
	}
}

func TestResolveAppliesEveryPrecedenceLayer(t *testing.T) {
	userDelimiter := ","
	userDirectory := "user"
	user := Config{
		SchemaVersion: CurrentSchemaVersion,
		Output: OutputConfig{
			Directory:    &userDirectory,
			TXTDelimiter: &userDelimiter,
		},
	}
	workflow := contract.WorkflowOverrides{
		OutputDirectory: "workflow",
		TXTDelimiter:    "|",
	}
	output := contract.OutputOverrides{
		Directory: "output",
		Delimiter: ";",
	}

	resolved, err := Resolve(BuiltInDefaults("temporary"), user, workflow, &output)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.OutputDirectory != "output" {
		t.Fatalf("directory = %q, want per-output value", resolved.OutputDirectory)
	}
	if resolved.TXTDelimiter != ";" {
		t.Fatalf("delimiter = %q, want per-output value", resolved.TXTDelimiter)
	}

	withoutOutput, err := Resolve(BuiltInDefaults("temporary"), user, workflow, nil)
	if err != nil {
		t.Fatal(err)
	}
	if withoutOutput.OutputDirectory != "workflow" || withoutOutput.TXTDelimiter != "|" {
		t.Fatalf("workflow layer not applied: %#v", withoutOutput)
	}
}

func TestInvalidCollisionReportsExactPathAndAllowedValues(t *testing.T) {
	document := "{\"schema_version\":\"v1\",\"output\":{\"collision\":\"delete\"}}"
	_, err := Decode(strings.NewReader(document), contract.DocumentJSON)
	if err == nil {
		t.Fatal("expected invalid collision action")
	}
	if !strings.Contains(err.Error(), "$.output.collision") ||
		!strings.Contains(err.Error(), "overwrite") ||
		!strings.Contains(err.Error(), "alternate_name") ||
		!strings.Contains(err.Error(), "block") {
		t.Fatalf("error lacks path or allowed values: %v", err)
	}
}
