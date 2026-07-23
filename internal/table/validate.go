package table

import (
	"fmt"
	"strings"
)

func ValidateTable(table Table) error {
	if table.ID == "" {
		return fmt.Errorf("table.id: value is required")
	}
	if table.Sheet.ID == "" || table.Sheet.SourceID == "" || table.Sheet.Name == "" || table.Sheet.Index < 0 {
		return fmt.Errorf("table.sheet: id, source_id, non-negative index, and name are required")
	}
	if len(table.Columns) == 0 {
		return fmt.Errorf("table.columns: at least one column is required")
	}
	ids := make(map[ColumnID]struct{}, len(table.Columns))
	positions := make(map[string]struct{}, len(table.Columns))
	aliases := make(map[string]struct{}, len(table.Columns))
	for index, column := range table.Columns {
		path := fmt.Sprintf("table.columns[%d]", index)
		if column.ID == "" {
			return fmt.Errorf("%s.id: value is required", path)
		}
		if _, exists := ids[column.ID]; exists {
			return fmt.Errorf("%s.id: duplicate column id %q", path, column.ID)
		}
		ids[column.ID] = struct{}{}
		if column.PhysicalPosition == "" {
			return fmt.Errorf("%s.physical_position: value is required", path)
		}
		if _, exists := positions[column.PhysicalPosition]; exists {
			return fmt.Errorf("%s.physical_position: duplicate position %q", path, column.PhysicalPosition)
		}
		positions[column.PhysicalPosition] = struct{}{}
		if column.OriginalHeader == "" {
			return fmt.Errorf("%s.original_header: value is required", path)
		}
		if column.GlobalAlias != "" {
			key := strings.ToLower(column.GlobalAlias)
			if _, exists := aliases[key]; exists {
				return fmt.Errorf("%s.global_alias: duplicate alias %q within table", path, column.GlobalAlias)
			}
			aliases[key] = struct{}{}
		}
		if !supportedValueKind(column.Type) {
			return fmt.Errorf("%s.type: unsupported canonical kind %q", path, column.Type)
		}
		if column.TypeAuthority != TypeInferred && column.TypeAuthority != TypeDeclared {
			return fmt.Errorf("%s.type_authority: unsupported authority %q", path, column.TypeAuthority)
		}
	}
	return nil
}

func ValidateRow(table Table, row Row) error {
	if row.ID == "" || row.SourceID == "" || row.SheetID == "" {
		return fmt.Errorf("row: id, source_id, and sheet_id are required")
	}
	if row.SourceID != table.Sheet.SourceID || row.SheetID != table.Sheet.ID {
		return fmt.Errorf("row: source or sheet identity does not match table")
	}
	if len(row.Cells) != len(table.Columns) {
		return fmt.Errorf("row.cells: got %d cells, want %d", len(row.Cells), len(table.Columns))
	}
	for index, cell := range row.Cells {
		column := table.Columns[index]
		if cell.ColumnID != column.ID {
			return fmt.Errorf("row.cells[%d].column_id: got %q, want %q", index, cell.ColumnID, column.ID)
		}
		if cell.Provenance.SourceID != row.SourceID ||
			cell.Provenance.SheetID != row.SheetID ||
			cell.Provenance.RowID != row.ID ||
			cell.Provenance.ColumnID != cell.ColumnID {
			return fmt.Errorf("row.cells[%d].provenance: identity does not match row and column", index)
		}
	}
	return nil
}

func supportedValueKind(kind ValueKind) bool {
	switch kind {
	case KindNull, KindString, KindBoolean, KindInteger, KindDecimal, KindDate, KindTime, KindJSONObject, KindJSONArray:
		return true
	default:
		return false
	}
}
