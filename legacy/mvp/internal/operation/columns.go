package operation

import (
	"errors"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

func resolveTableColumn(columns []table.Column, reference contract.ColumnRef) (int, error) {
	for index, column := range columns {
		if (reference.Header != "" && (column.OriginalHeader == reference.Header || column.Header == reference.Header)) ||
			(reference.Alias != "" && column.GlobalAlias == reference.Alias) ||
			(reference.Position != "" && column.PhysicalPosition == reference.Position) {
			return index, nil
		}
	}
	return 0, errors.New("column reference was not found")
}
