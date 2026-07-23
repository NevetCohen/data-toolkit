package operation

import (
	"context"
	"io"
	"strings"
	"testing"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

func TestApplyRenamePreservesOrderUnicodeAndProvenance(t *testing.T) {
	sourceTable, rows := operationFixture()
	transformedTable, stream, err := ApplyRename(
		sourceTable,
		table.NewSliceRowStream(rows),
		[]contract.RenameField{{Column: contract.ColumnRef{Header: "name"}, To: "שם מלא"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	if got := transformedTable.Columns[0].ID; got != "שם מלא" {
		t.Fatalf("renamed column ID = %q", got)
	}
	if transformedTable.Columns[1].ID != "note" {
		t.Fatalf("unchanged column moved or changed: %#v", transformedTable.Columns)
	}
	for index, wantID := range []table.RowID{"row-1", "row-2"} {
		row, err := stream.Next(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if row.ID != wantID || row.Cells[0].ColumnID != "שם מלא" {
			t.Fatalf("row %d = %#v", index, row)
		}
		if row.Cells[0].Provenance.ColumnID != "שם מלא" ||
			row.Cells[0].Provenance.SourceID != row.SourceID ||
			row.Cells[0].Provenance.RowID != row.ID {
			t.Fatalf("renamed provenance = %#v", row.Cells[0].Provenance)
		}
	}
	if _, err := stream.Next(context.Background()); err != io.EOF {
		t.Fatalf("end of stream error = %v", err)
	}
	if got, _ := rows[0].Cells[0].Value.StringContent(); got != " נועה 🧭 " || rows[0].Cells[0].ColumnID != "name" {
		t.Fatalf("source row was mutated: %#v", rows[0])
	}
}

func TestApplyRenameRejectsInvalidTargets(t *testing.T) {
	sourceTable, rows := operationFixture()
	tests := []struct {
		name    string
		renames []contract.RenameField
		want    string
	}{
		{
			name: "empty target",
			renames: []contract.RenameField{
				{Column: contract.ColumnRef{Header: "name"}, To: "  "},
			},
			want: "target is required",
		},
		{
			name: "duplicate source",
			renames: []contract.RenameField{
				{Column: contract.ColumnRef{Header: "name"}, To: "full_name"},
				{Column: contract.ColumnRef{Position: "A"}, To: "display_name"},
			},
			want: "repeats source",
		},
		{
			name: "duplicate target",
			renames: []contract.RenameField{
				{Column: contract.ColumnRef{Header: "name"}, To: "combined"},
				{Column: contract.ColumnRef{Header: "note"}, To: "combined"},
			},
			want: "repeats target",
		},
		{
			name: "unchanged collision",
			renames: []contract.RenameField{
				{Column: contract.ColumnRef{Header: "name"}, To: "note"},
			},
			want: "collides with unchanged column",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := ApplyRename(sourceTable, table.NewSliceRowStream(rows), test.renames)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestApplyRenameAllowsTargetsVacatedByOtherRenames(t *testing.T) {
	sourceTable, rows := operationFixture()
	transformed, stream, err := ApplyRename(sourceTable, table.NewSliceRowStream(rows), []contract.RenameField{
		{Column: contract.ColumnRef{Header: "name"}, To: "note"},
		{Column: contract.ColumnRef{Header: "note"}, To: "comment"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if transformed.Columns[0].ID != "note" || transformed.Columns[1].ID != "comment" {
		t.Fatalf("columns = %#v", transformed.Columns)
	}
}

func TestApplyConcatenatePreservesUnicodeNullsAndInternalWhitespace(t *testing.T) {
	sourceTable, rows := operationFixture()
	transformedTable, stream, err := ApplyConcatenate(
		sourceTable,
		table.NewSliceRowStream(rows),
		contract.TextOptions{
			Target:    contract.ColumnRef{Header: "combined"},
			Sources:   []contract.ColumnRef{{Header: "name"}, {Header: "note"}},
			Separator: " | ",
			Transform: "trim",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if got := transformedTable.Columns[len(transformedTable.Columns)-1].ID; got != "combined" {
		t.Fatalf("derived column = %q", got)
	}

	first, err := stream.Next(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	combined, ok := first.Cells[2].Value.StringContent()
	if !ok || combined != "נועה 🧭 | internal  space" {
		t.Fatalf("combined value = %q", combined)
	}
	if first.Cells[2].Provenance.SourceID != first.SourceID ||
		first.Cells[2].Provenance.RowID != first.ID ||
		first.Cells[2].Provenance.ColumnID != "combined" {
		t.Fatalf("derived provenance = %#v", first.Cells[2].Provenance)
	}

	second, err := stream.Next(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	combined, ok = second.Cells[2].Value.StringContent()
	if !ok || combined != "יוסי" {
		t.Fatalf("null concatenation value = %q", combined)
	}
}

func operationFixture() (table.Table, []table.Row) {
	sheet := table.Sheet{ID: "sheet-1", SourceID: "source-1", Name: "Sheet1", Index: 0}
	columns := []table.Column{
		{
			ID: "name", PhysicalPosition: "A", OriginalHeader: "name",
			Header: "name", Type: table.KindString, TypeAuthority: table.TypeDeclared,
		},
		{
			ID: "note", PhysicalPosition: "B", OriginalHeader: "note",
			Header: "note", Type: table.KindString, TypeAuthority: table.TypeDeclared,
		},
	}
	sourceTable := table.Table{ID: "table-1", Sheet: sheet, Columns: columns}
	row := func(id table.RowID, ordinal uint64, values ...table.Value) table.Row {
		cells := make([]table.Cell, len(values))
		for index, value := range values {
			cells[index] = table.Cell{
				ColumnID: columns[index].ID,
				Value:    value,
				Provenance: table.Provenance{
					SourceID: sheet.SourceID, SheetID: sheet.ID, RowID: id,
					ColumnID: columns[index].ID, PhysicalRow: int(ordinal) + 1,
					PhysicalColumn: columns[index].PhysicalPosition,
				},
			}
		}
		return table.Row{
			ID: id, SourceID: sheet.SourceID, SheetID: sheet.ID, Ordinal: ordinal, Cells: cells,
		}
	}
	return sourceTable, []table.Row{
		row("row-1", 0, table.StringValue(" נועה 🧭 "), table.StringValue(" internal  space ")),
		row("row-2", 1, table.StringValue("יוסי"), table.NullValue()),
	}
}
