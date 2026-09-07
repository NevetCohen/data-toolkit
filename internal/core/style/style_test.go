package style

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestStyleDescriptorHasExactV1Fields(t *testing.T) {
	typeOfDescriptor := reflect.TypeFor[StyleDescriptor]()
	want := []struct {
		name string
		json string
	}{
		{"Name", "name"},
		{"RightToLeft", "right_to_left"},
	}
	if typeOfDescriptor.NumField() != len(want) {
		t.Fatalf("StyleDescriptor field count = %d, want %d", typeOfDescriptor.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfDescriptor.Field(index)
		if field.Name != expected.name || field.Tag.Get("json") != expected.json {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, expected.json)
		}
	}
}

func TestStyleDescriptorJSONUsesExactFields(t *testing.T) {
	data, err := json.Marshal(StyleDescriptor{Name: "Hebrew", RightToLeft: true})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"name":"Hebrew","right_to_left":true}`; got != want {
		t.Fatalf("descriptor JSON = %s, want %s", got, want)
	}
}
