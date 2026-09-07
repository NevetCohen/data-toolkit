package table

import (
	"context"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"testing"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/cell"
	"data-toolkit/internal/core/column"
)

func stringPointer(value string) *string {
	return &value
}

func TestDatasetHasExactV1Fields(t *testing.T) {
	typeOfDataset := reflect.TypeFor[Dataset]()
	want := []struct {
		name   string
		typeOf reflect.Type
		tag    string
	}{
		{name: "Schema", typeOf: reflect.TypeFor[Schema](), tag: `json:"-"`},
		{name: "Rows", typeOf: reflect.TypeFor[RowStream](), tag: `json:"-"`},
	}
	if typeOfDataset.NumField() != len(want) {
		t.Fatalf("Dataset field count = %d, want %d", typeOfDataset.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfDataset.Field(index)
		if field.Name != expected.name || field.Type != expected.typeOf || string(field.Tag) != expected.tag {
			t.Errorf("field %d = %s %s %q, want %s %s %q", index, field.Name, field.Type, field.Tag, expected.name, expected.typeOf, expected.tag)
		}
	}
}

func TestDatasetJSONDoesNotExposeInternalSchemaOrRowStream(t *testing.T) {
	schema := rowTestSchema(t)
	row := rowForSchema(t, schema)
	dataset := Dataset{Schema: schema, Rows: NewSliceStream([]Row{row})}

	encoded, err := json.Marshal(dataset)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(encoded), `{}`; got != want {
		t.Errorf("Dataset JSON = %s, want %s", got, want)
	}
}

func TestDatasetValidateRequiresSchemaAndUsableRowStream(t *testing.T) {
	schema := rowTestSchema(t)
	row := rowForSchema(t, schema)

	var typedNil *SliceStream
	tests := []struct {
		name    string
		dataset Dataset
		want    string
	}{
		{
			name:    "valid",
			dataset: Dataset{Schema: schema, Rows: NewSliceStream([]Row{row})},
		},
		{
			name:    "invalid schema",
			dataset: Dataset{Schema: Schema{}, Rows: NewSliceStream(nil)},
			want:    "dataset.schema",
		},
		{
			name:    "nil stream",
			dataset: Dataset{Schema: schema},
			want:    "dataset.rows must be a non-nil row stream",
		},
		{
			name:    "typed nil stream",
			dataset: Dataset{Schema: schema, Rows: typedNil},
			want:    "dataset.rows must be a non-nil row stream",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.dataset.Validate()
			if test.want == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestCanonicalMetadataAndRowStream(t *testing.T) {
	value, err := core.NewValue("string", []byte(`"Alice"`))
	if err != nil {
		t.Fatal(err)
	}
	schema := Schema{
		ID:       "people",
		SourceID: "source",
		SheetID:  stringPointer("sheet"),
		Columns: []column.Descriptor{{
			ID:               "name",
			PhysicalPosition: "A",
			SourceHeader:     "Name",
			DataType:         "string",
			TypeAuthority:    core.TypeAuthorityDeclared,
		}},
	}
	if err := schema.Validate(); err != nil {
		t.Fatal(err)
	}
	row := Row{
		ID:       "row-1",
		SourceID: "source",
		SheetID:  stringPointer("sheet"),
		Ordinal:  1,
		Cells: []cell.Cell{{
			ColumnID: "name",
			Value:    value,
		}},
	}
	if err := row.Validate(schema); err != nil {
		t.Fatal(err)
	}
	stream := NewSliceStream([]Row{row})
	got, err := stream.Next(context.Background())
	if err != nil || got.ID != row.ID {
		t.Fatalf("next row = %#v, %v", got, err)
	}
	if _, err := stream.Next(context.Background()); err != io.EOF {
		t.Fatalf("end error = %v, want io.EOF", err)
	}
	if len(schema.Columns) != 1 {
		t.Fatal("column metadata unexpectedly materialized row values")
	}
}
