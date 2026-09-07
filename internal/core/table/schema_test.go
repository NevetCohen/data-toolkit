package table

import (
	"reflect"
	"strings"
	"testing"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/column"
)

func TestSchemaHasExactV1Fields(t *testing.T) {
	typeOfSchema := reflect.TypeFor[Schema]()
	want := []struct {
		name   string
		typeOf reflect.Type
	}{
		{name: "ID", typeOf: reflect.TypeFor[core.TableID]()},
		{name: "SourceID", typeOf: reflect.TypeFor[core.SourceID]()},
		{name: "SheetID", typeOf: reflect.TypeFor[*string]()},
		{name: "Columns", typeOf: reflect.TypeFor[[]column.ColumnDescriptor]()},
	}
	if typeOfSchema.NumField() != len(want) {
		t.Fatalf("Schema field count = %d, want %d", typeOfSchema.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfSchema.Field(index)
		if field.Name != expected.name || field.Type != expected.typeOf {
			t.Errorf("field %d = %s %s, want %s %s", index, field.Name, field.Type, expected.name, expected.typeOf)
		}
	}
}

func TestSchemaValidate(t *testing.T) {
	sheetID := "Sheet1"
	valid := Schema{
		ID:       "people",
		SourceID: "source",
		SheetID:  &sheetID,
		Columns: []column.ColumnDescriptor{{
			ID:               "name",
			PhysicalPosition: "A",
			SourceHeader:     "Name",
			DataType:         "string",
			TypeAuthority:    core.TypeAuthorityDeclared,
		}, {
			ID:               "email",
			PhysicalPosition: "B",
			SourceHeader:     "Email",
			DataType:         "string",
			TypeAuthority:    core.TypeAuthorityInferred,
		}},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid schema: %v", err)
	}

	tests := []struct {
		name string
		edit func(*Schema)
		want string
	}{
		{name: "missing ID", edit: func(value *Schema) { value.ID = "" }, want: "schema.id"},
		{name: "invalid source ID", edit: func(value *Schema) { value.SourceID = " " }, want: "schema.source_id"},
		{name: "invalid sheet ID", edit: func(value *Schema) { sheet := " "; value.SheetID = &sheet }, want: "schema.sheet_id"},
		{name: "no columns", edit: func(value *Schema) { value.Columns = nil }, want: "schema.columns must be non-empty"},
		{name: "invalid column", edit: func(value *Schema) { value.Columns[0].ID = "" }, want: "schema.columns[0]"},
		{name: "duplicate column ID", edit: func(value *Schema) { value.Columns[1].ID = value.Columns[0].ID }, want: "schema.columns[1]: duplicate column"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := cloneSchema(valid)
			test.edit(&value)
			if err := value.Validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestSchemaValidateAllowsAbsentSheetID(t *testing.T) {
	schema := Schema{
		ID:       "people",
		SourceID: "source",
		Columns: []column.ColumnDescriptor{{
			ID:               "name",
			PhysicalPosition: "A",
			SourceHeader:     "Name",
			DataType:         "string",
			TypeAuthority:    core.TypeAuthorityDeclared,
		}},
	}
	if err := schema.Validate(); err != nil {
		t.Fatalf("schema without sheet ID: %v", err)
	}
}

func cloneSchema(schema Schema) Schema {
	clone := schema
	clone.Columns = append([]column.ColumnDescriptor(nil), schema.Columns...)
	if schema.SheetID != nil {
		sheetID := *schema.SheetID
		clone.SheetID = &sheetID
	}
	return clone
}
