package table

import (
	"context"
	"io"
	"testing"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/cell"
	"data-toolkit/internal/core/column"
	"data-toolkit/internal/core/datatype"
)

func TestCanonicalMetadataAndRowStream(t *testing.T) {
	value, err := datatype.NewValue(datatype.StringID, []byte(`"Alice"`))
	if err != nil {
		t.Fatal(err)
	}
	schema := Schema{
		ID:       "people",
		SourceID: "source",
		SheetID:  "sheet",
		Columns: []column.Descriptor{{
			ID:            "name",
			Header:        "Name",
			Subheader:     "Canonical",
			DataType:      datatype.StringID,
			TypeAuthority: column.TypeDeclared,
			Range:         &column.Range{StartRow: 2, EndRow: 10, StartColumn: "A", EndColumn: "A"},
			Binding:       column.SourceBinding{SourceID: "source", SheetID: "sheet", Kind: "opaque", Locator: "A"},
		}},
	}
	if err := schema.Validate(); err != nil {
		t.Fatal(err)
	}
	row := Row{
		ID:       "row-1",
		SourceID: "source",
		SheetID:  "sheet",
		Ordinal:  1,
		Cells: []cell.Cell{{
			ColumnID: "name",
			Value:    value,
			Provenance: cell.Provenance{
				SourceID: "source",
				SheetID:  "sheet",
				RowID:    "row-1",
				ColumnID: core.ColumnID("name"),
			},
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
