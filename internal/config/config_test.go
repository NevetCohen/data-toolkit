package config

import (
	"strings"
	"testing"

	"data-toolkit/configs"
)

func TestEmbeddedDefaultsPassStrictSchema(t *testing.T) {
	document, err := LoadDefaults()
	if err != nil {
		t.Fatalf("load defaults: %v", err)
	}
	if document.SchemaVersion != CurrentVersion || document.Runtime.MaximumMemoryBytes <= 0 {
		t.Fatalf("defaults were not resolved to typed configuration: %#v", document)
	}
}

func TestStrictDecodeRejectsUnknownAndDuplicateSemanticIdentity(t *testing.T) {
	payload, err := configs.Default()
	if err != nil {
		t.Fatal(err)
	}
	unknown := append(append([]byte(nil), payload...), []byte("\nunknown: true\n")...)
	if _, err := Decode(unknown, FormatYAML); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("unknown field error = %v", err)
	}

	duplicate := strings.Replace(string(payload), "aliases: []", "aliases:\n  - name: phone\n    semantic_type: Phones\n  - name: PHONE\n    semantic_type: Phones", 1)
	if _, err := Decode([]byte(duplicate), FormatYAML); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate identity error = %v", err)
	}
}

func TestResolveAppliesPrecedenceWithoutMutation(t *testing.T) {
	base, err := LoadDefaults()
	if err != nil {
		t.Fatal(err)
	}
	userDirectory := "user"
	workflowDirectory := "workflow"
	outputDirectory := "output"
	resolved, err := Resolve(
		base,
		Patch{Output: &OutputPatch{Directory: &userDirectory}},
		Patch{Output: &OutputPatch{Directory: &workflowDirectory}},
		Patch{Output: &OutputPatch{Directory: &outputDirectory}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Output.Directory != "output" {
		t.Fatalf("resolved directory = %q", resolved.Output.Directory)
	}
	if base.Output.Directory == resolved.Output.Directory {
		t.Fatal("base configuration was mutated")
	}
	resolved.Styles[0].Name = "changed"
	if base.Styles[0].Name == "changed" {
		t.Fatal("resolved slices alias the defaults")
	}
}
