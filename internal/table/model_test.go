package table

import "testing"

func TestTablePreservesPhysicalPositionHeaderStructureAndIdentity(t *testing.T) {
	table := Table{
		ID:    "members",
		Sheet: Sheet{ID: "members-sheet-2", SourceID: "workbook", Name: "פעילים 🧭", Index: 1},
		Columns: []Column{
			{
				ID: "phone", PhysicalPosition: "C", OriginalHeader: "טלפון",
				GlobalAlias: "Phones", Type: KindString, TypeAuthority: TypeDeclared,
				Header: "פרטי קשר", Subheader: "טלפון",
			},
			{
				ID: "name", PhysicalPosition: "D", OriginalHeader: "שם",
				Type: KindString, TypeAuthority: TypeInferred,
				Header: "פרטים", Subheader: "שם מלא",
			},
		},
	}
	if err := ValidateTable(table); err != nil {
		t.Fatal(err)
	}
	if table.Columns[0].PhysicalPosition != "C" {
		t.Fatalf("first physical position = %q", table.Columns[0].PhysicalPosition)
	}
	if table.Columns[0].Header != "פרטי קשר" || table.Columns[0].Subheader != "טלפון" {
		t.Fatalf("header structure was not retained: %#v", table.Columns[0])
	}

	row := Row{
		ID: "row-7", SourceID: "workbook", SheetID: "members-sheet-2", Ordinal: 6,
		Cells: []Cell{
			{ColumnID: "phone", Value: StringValue("052-6105412"), Provenance: Provenance{SourceID: "workbook", SheetID: "members-sheet-2", RowID: "row-7", ColumnID: "phone", PhysicalRow: 9, PhysicalColumn: "C"}},
			{ColumnID: "name", Value: StringValue("נועם 🧭"), Provenance: Provenance{SourceID: "workbook", SheetID: "members-sheet-2", RowID: "row-7", ColumnID: "name", PhysicalRow: 9, PhysicalColumn: "D"}},
		},
	}
	if err := ValidateRow(table, row); err != nil {
		t.Fatal(err)
	}
}

func TestGlobalAliasCanBeReusedAcrossSources(t *testing.T) {
	first := Table{
		ID: "first", Sheet: Sheet{ID: "first-sheet", SourceID: "first-source", Name: "Sheet1"},
		Columns: []Column{{ID: "a", PhysicalPosition: "A", OriginalHeader: "טלפון", GlobalAlias: "Phones", Type: KindString, TypeAuthority: TypeInferred}},
	}
	second := Table{
		ID: "second", Sheet: Sheet{ID: "second-sheet", SourceID: "second-source", Name: "Sheet1"},
		Columns: []Column{{ID: "b", PhysicalPosition: "B", OriginalHeader: "מספר נייד", GlobalAlias: "Phones", Type: KindString, TypeAuthority: TypeInferred}},
	}
	if err := ValidateTable(first); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTable(second); err != nil {
		t.Fatal(err)
	}
}
