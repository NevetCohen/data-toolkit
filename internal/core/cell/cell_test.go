package cell

import (
	"encoding/json"
	"reflect"
	"testing"

	"data-toolkit/internal/core"
)

func TestCellHasExactV1FieldsAndCanonicalValue(t *testing.T) {
	typeOfCell := reflect.TypeFor[Cell]()
	want := []struct {
		name   string
		typeOf reflect.Type
		tag    string
	}{
		{name: "ColumnID", typeOf: reflect.TypeFor[core.ColumnID](), tag: `json:"column_id"`},
		{name: "Value", typeOf: reflect.TypeFor[core.Value](), tag: `json:"-"`},
	}
	if typeOfCell.NumField() != len(want) {
		t.Fatalf("Cell field count = %d, want %d", typeOfCell.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfCell.Field(index)
		if field.Name != expected.name || field.Type != expected.typeOf || string(field.Tag) != expected.tag {
			t.Errorf("field %d = %s %s %q, want %s %s %q", index, field.Name, field.Type, field.Tag, expected.name, expected.typeOf, expected.tag)
		}
	}

	value, err := core.NewValue("string", []byte(`"canonical"`))
	if err != nil {
		t.Fatal(err)
	}
	if err := (Cell{ColumnID: "value", Value: value}).Validate(); err != nil {
		t.Fatalf("valid canonical cell: %v", err)
	}
}

func TestCellJSONExposesOnlyColumnID(t *testing.T) {
	value, err := core.NewValue("string", []byte(`"secret"`))
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := json.Marshal(Cell{ColumnID: "value", Value: value})
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"column_id":"value"}` {
		t.Fatalf("JSON = %s, want %s", encoded, `{"column_id":"value"}`)
	}
}
