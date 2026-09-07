package datatype

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDataTypeDescriptorHasExactV1Fields(t *testing.T) {
	typeOfDescriptor := reflect.TypeFor[DataTypeDescriptor]()
	want := []struct {
		name string
		json string
	}{
		{"ID", "id"},
		{"Version", "version"},
		{"Name", "name"},
		{"CanonicalKind", "canonical_kind"},
	}
	if typeOfDescriptor.NumField() != len(want) {
		t.Fatalf("DataTypeDescriptor field count = %d, want %d", typeOfDescriptor.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfDescriptor.Field(index)
		if field.Name != expected.name || field.Tag.Get("json") != expected.json {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, expected.json)
		}
	}
}

func TestCanonicalKindsAreExactlyTheSevenV1Kinds(t *testing.T) {
	valid := []CanonicalKind{
		CanonicalKindNull, CanonicalKindString, CanonicalKindInteger,
		CanonicalKindDecimal, CanonicalKindBoolean, CanonicalKindDate, CanonicalKindTime,
	}
	for _, kind := range valid {
		if err := kind.Validate(); err != nil {
			t.Errorf("valid canonical kind %q: %v", kind, err)
		}
	}
	if err := CanonicalKind("json_object").Validate(); err == nil {
		t.Fatal("unknown canonical kind was accepted")
	}

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	got := registry.Descriptors()
	if len(got) != len(valid) {
		t.Fatalf("built-in descriptor count = %d, want %d", len(got), len(valid))
	}
	for _, descriptor := range got {
		if err := descriptor.CanonicalKind.Validate(); err != nil {
			t.Errorf("built-in %q has invalid canonical kind: %v", descriptor.ID, err)
		}
	}
}

func TestDataTypeDescriptorJSONUsesExactFields(t *testing.T) {
	data, err := json.Marshal(DataTypeDescriptor{
		ID: "string", Version: CurrentVersion, Name: "String", CanonicalKind: CanonicalKindString,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"id":"string","version":"v1","name":"String","canonical_kind":"string"}`; got != want {
		t.Fatalf("descriptor JSON = %s, want %s", got, want)
	}
}
