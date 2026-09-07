package table

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/cell"
	"data-toolkit/internal/core/column"
)

func TestRowHasExactV1Fields(t *testing.T) {
	typeOfRow := reflect.TypeFor[Row]()
	want := []struct {
		name   string
		typeOf reflect.Type
		tag    string
	}{
		{name: "ID", typeOf: reflect.TypeFor[core.RowID](), tag: `json:"id"`},
		{name: "SourceID", typeOf: reflect.TypeFor[core.SourceID](), tag: `json:"source_id"`},
		{name: "SheetID", typeOf: reflect.TypeFor[*string](), tag: `json:"sheet_id,omitempty"`},
		{name: "Ordinal", typeOf: reflect.TypeFor[uint64](), tag: `json:"ordinal"`},
		{name: "Cells", typeOf: reflect.TypeFor[[]cell.Cell](), tag: `json:"-"`},
	}
	if typeOfRow.NumField() != len(want) {
		t.Fatalf("Row field count = %d, want %d", typeOfRow.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfRow.Field(index)
		if field.Name != expected.name || field.Type != expected.typeOf || string(field.Tag) != expected.tag {
			t.Errorf("field %d = %s %s %q, want %s %s %q", index, field.Name, field.Type, field.Tag, expected.name, expected.typeOf, expected.tag)
		}
	}
}

func TestRowJSONExposesOnlyPublicFields(t *testing.T) {
	schema := rowTestSchema(t)
	row := rowForSchema(t, schema)

	encoded, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(encoded), `{"id":"row-1","source_id":"source","sheet_id":"Sheet1","ordinal":0}`; got != want {
		t.Errorf("Row JSON = %s, want %s", got, want)
	}

	row.SheetID = nil
	encoded, err = json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(encoded), `{"id":"row-1","source_id":"source","ordinal":0}`; got != want {
		t.Errorf("Row JSON without sheet ID = %s, want %s", got, want)
	}
}

func TestRowValidateRejectsEachSchemaMismatch(t *testing.T) {
	schema := rowTestSchema(t)
	valid := rowForSchema(t, schema)
	if err := valid.Validate(schema); err != nil {
		t.Fatalf("valid row: %v", err)
	}
	integer, err := core.NewValue("integer", []byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		edit func(*Row)
		want string
	}{
		{name: "SourceID", edit: func(row *Row) { row.SourceID = "other" }, want: "row.source_id"},
		{name: "SheetID absent", edit: func(row *Row) { row.SheetID = nil }, want: "row.sheet_id"},
		{name: "SheetID value", edit: func(row *Row) { row.SheetID = stringPointer("other") }, want: "row.sheet_id"},
		{name: "cell count", edit: func(row *Row) { row.Cells = row.Cells[:1] }, want: "row has 1 cells; table has 2 columns"},
		{name: "cell order", edit: func(row *Row) { row.Cells[0], row.Cells[1] = row.Cells[1], row.Cells[0] }, want: "does not match table column"},
		{name: "Cell.ColumnID", edit: func(row *Row) { row.Cells[0].ColumnID = "other" }, want: "does not match table column"},
		{name: "Value.DataTypeID", edit: func(row *Row) { row.Cells[0].Value = integer }, want: "does not match schema column type"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			row := cloneRow(valid)
			test.edit(&row)
			if err := row.Validate(schema); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func rowTestSchema(t *testing.T) Schema {
	t.Helper()
	sheetID := "Sheet1"
	return Schema{
		ID:       "people",
		SourceID: "source",
		SheetID:  &sheetID,
		Columns: []column.ColumnDescriptor{
			{ID: "name", PhysicalPosition: "A", SourceHeader: "Name", DataType: "string", TypeAuthority: core.TypeAuthorityDeclared},
			{ID: "age", PhysicalPosition: "B", SourceHeader: "Age", DataType: "integer", TypeAuthority: core.TypeAuthorityInferred},
		},
	}
}

func rowForSchema(t *testing.T, schema Schema) Row {
	t.Helper()
	name, err := core.NewValue("string", []byte(`"Alice"`))
	if err != nil {
		t.Fatal(err)
	}
	age, err := core.NewValue("integer", []byte("42"))
	if err != nil {
		t.Fatal(err)
	}
	return Row{
		ID:       "row-1",
		SourceID: schema.SourceID,
		SheetID:  stringPointer(*schema.SheetID),
		Ordinal:  0,
		Cells: []cell.Cell{
			{ColumnID: "name", Value: name},
			{ColumnID: "age", Value: age},
		},
	}
}

func cloneRow(row Row) Row {
	clone := row
	clone.Cells = append([]cell.Cell(nil), row.Cells...)
	if row.SheetID != nil {
		sheetID := *row.SheetID
		clone.SheetID = &sheetID
	}
	return clone
}
