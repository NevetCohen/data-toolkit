package operation

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestOperationDescriptorHasExactV1Fields(t *testing.T) {
	typeOfDescriptor := reflect.TypeFor[OperationDescriptor]()
	want := []struct {
		name string
		json string
	}{
		{"ID", "id"},
		{"Version", "version"},
		{"Owner", "owner"},
		{"Exposure", "exposure"},
		{"Name", "name"},
		{"Description", "description"},
		{"Deterministic", "deterministic"},
		{"StreamingMode", "streaming_mode"},
		{"InputShape", "input_shape"},
		{"OutputShape", "output_shape"},
		{"ParameterSchema", "parameter_schema"},
		{"RequiredCapabilities", "required_capabilities"},
	}
	if typeOfDescriptor.NumField() != len(want) {
		t.Fatalf("OperationDescriptor field count = %d, want %d", typeOfDescriptor.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfDescriptor.Field(index)
		if field.Name != expected.name || field.Tag.Get("json") != expected.json {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, expected.json)
		}
	}
}

func TestOperationDescriptorJSONUsesExactFields(t *testing.T) {
	data, err := json.Marshal(OperationDescriptor{
		ID: "row.filter", Version: "v1", Owner: "logical_engine", Exposure: "workflow",
		Name: "Filter rows", Description: "Filters rows", Deterministic: true,
		StreamingMode: "streaming", InputShape: "Dataset", OutputShape: "Dataset",
		ParameterSchema: json.RawMessage(`{"type":"object"}`), RequiredCapabilities: []string{"canonical_values"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":"row.filter","version":"v1","owner":"logical_engine","exposure":"workflow","name":"Filter rows","description":"Filters rows","deterministic":true,"streaming_mode":"streaming","input_shape":"Dataset","output_shape":"Dataset","parameter_schema":{"type":"object"},"required_capabilities":["canonical_values"]}`
	if got := string(data); got != want {
		t.Fatalf("descriptor JSON = %s, want %s", got, want)
	}
}

func TestOperationRefHasExactResolvedFields(t *testing.T) {
	typeOfRef := reflect.TypeFor[OperationRef]()
	want := []struct {
		name string
		json string
	}{
		{"OperationID", "operation_id"},
		{"Version", "version"},
		{"Owner", "owner"},
	}
	if typeOfRef.NumField() != len(want) {
		t.Fatalf("OperationRef field count = %d, want %d", typeOfRef.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfRef.Field(index)
		if field.Name != expected.name || field.Tag.Get("json") != expected.json {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, expected.json)
		}
	}
}

func TestOperationRefJSONUsesResolvedIDVersionAndOwner(t *testing.T) {
	data, err := json.Marshal(OperationRef{
		OperationID: "row.filter", Version: "v1", Owner: "logical_engine",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"operation_id":"row.filter","version":"v1","owner":"logical_engine"}`; got != want {
		t.Fatalf("operation ref JSON = %s, want %s", got, want)
	}
}
