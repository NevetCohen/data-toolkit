package contract

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestEquivalentYAMLAndJSONDecodeToSameWorkflow(t *testing.T) {
	jsonFile, err := os.Open("testdata/workflow.json")
	if err != nil {
		t.Fatal(err)
	}
	defer jsonFile.Close()

	yamlFile, err := os.Open("testdata/workflow.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer yamlFile.Close()

	fromJSON, err := DecodeWorkflow(jsonFile, DocumentJSON)
	if err != nil {
		t.Fatal(err)
	}
	fromYAML, err := DecodeWorkflow(yamlFile, DocumentYAML)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(fromJSON, fromYAML) {
		t.Fatalf("decoded workflows differ:\nJSON: %#v\nYAML: %#v", fromJSON, fromYAML)
	}
}

func TestDecodeWorkflowRejectsMultipleDocuments(t *testing.T) {
	tests := []struct {
		name   string
		format DocumentFormat
		value  string
	}{
		{name: "json", format: DocumentJSON, value: `{}` + "\n" + `{}`},
		{name: "yaml", format: DocumentYAML, value: "{}\n---\n{}\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := DecodeWorkflow(strings.NewReader(test.value), test.format); err == nil {
				t.Fatal("expected multiple-document error")
			}
		})
	}
}
