package locator

import (
	"encoding/json"
	"reflect"
	"testing"

	"data-toolkit/internal/core"
)

func TestSourceLocatorHasExactFields(t *testing.T) {
	typeOfLocator := reflect.TypeFor[SourceLocator]()
	want := []struct {
		name string
		json string
	}{
		{"SourceID", "source_id"},
		{"SheetID", "sheet_id"},
		{"RootPath", "root_path"},
		{"RowOrdinal", "row_ordinal"},
		{"ColumnID", "column_id"},
		{"PhysicalPosition", "physical_position"},
	}
	if typeOfLocator.NumField() != len(want) {
		t.Fatalf("SourceLocator field count = %d, want %d", typeOfLocator.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfLocator.Field(index)
		wantTag := expected.json
		if index > 0 {
			wantTag += ",omitempty"
		}
		if field.Name != expected.name || field.Tag.Get("json") != wantTag {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, wantTag)
		}
	}
}

func TestOutputLocatorHasExactFields(t *testing.T) {
	typeOfLocator := reflect.TypeFor[OutputLocator]()
	want := []struct {
		name string
		json string
	}{
		{"OutputPath", "output_path"},
		{"RowOrdinal", "row_ordinal"},
		{"ColumnID", "column_id"},
		{"PhysicalPosition", "physical_position"},
	}
	if typeOfLocator.NumField() != len(want) {
		t.Fatalf("OutputLocator field count = %d, want %d", typeOfLocator.NumField(), len(want))
	}
	for index, expected := range want {
		field := typeOfLocator.Field(index)
		wantTag := expected.json
		if index > 0 {
			wantTag += ",omitempty"
		}
		if field.Name != expected.name || field.Tag.Get("json") != wantTag {
			t.Errorf("field %d = %s with json %q, want %s with json %q", index, field.Name, field.Tag.Get("json"), expected.name, wantTag)
		}
	}
}

func TestLocatorsOmitAbsentOptionalFields(t *testing.T) {
	source := SourceLocator{SourceID: core.SourceID("source-1")}
	output := OutputLocator{OutputPath: `C:\out\result.csv`}

	sourceJSON, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(sourceJSON), `{"source_id":"source-1"}`; got != want {
		t.Fatalf("source locator JSON = %s, want %s", got, want)
	}
	outputJSON, err := json.Marshal(output)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(outputJSON), `{"output_path":"C:\\out\\result.csv"}`; got != want {
		t.Fatalf("output locator JSON = %s, want %s", got, want)
	}
}

func TestLocatorsSerializeAllOptionalFields(t *testing.T) {
	sheet := core.SheetID("sheet-1")
	root := "$.rows"
	row := uint64(4)
	column := core.ColumnID("column-1")
	position := "B5"
	source := SourceLocator{
		SourceID: core.SourceID("source-1"), SheetID: &sheet, RootPath: &root,
		RowOrdinal: &row, ColumnID: &column, PhysicalPosition: &position,
	}
	output := OutputLocator{OutputPath: `C:\out\result.csv`, RowOrdinal: &row, ColumnID: &column, PhysicalPosition: &position}

	sourceJSON, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(sourceJSON), `{"source_id":"source-1","sheet_id":"sheet-1","root_path":"$.rows","row_ordinal":4,"column_id":"column-1","physical_position":"B5"}`; got != want {
		t.Fatalf("source locator JSON = %s, want %s", got, want)
	}
	outputJSON, err := json.Marshal(output)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(outputJSON), `{"output_path":"C:\\out\\result.csv","row_ordinal":4,"column_id":"column-1","physical_position":"B5"}`; got != want {
		t.Fatalf("output locator JSON = %s, want %s", got, want)
	}
}
