// Package cell defines canonical cells and source provenance.
package cell

import (
	"fmt"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/datatype"
)

type Provenance struct {
	SourceID core.SourceID `json:"source_id"`
	SheetID  core.SheetID  `json:"sheet_id"`
	RowID    core.RowID    `json:"row_id"`
	ColumnID core.ColumnID `json:"column_id"`
	Locator  string        `json:"locator,omitempty"`
}

type Cell struct {
	ColumnID   core.ColumnID `json:"column_id"`
	Value      datatype.Value
	Provenance Provenance `json:"provenance"`
}

func (value Cell) Validate() error {
	if value.ColumnID == "" {
		return fmt.Errorf("cell.column_id is required")
	}
	if value.Value.TypeID() == "" {
		return fmt.Errorf("cell.value type is required")
	}
	if value.Provenance.SourceID == "" {
		return fmt.Errorf("cell.provenance.source_id is required")
	}
	if value.Provenance.SheetID == "" {
		return fmt.Errorf("cell.provenance.sheet_id is required")
	}
	if value.Provenance.RowID == "" {
		return fmt.Errorf("cell.provenance.row_id is required")
	}
	if value.Provenance.ColumnID != value.ColumnID {
		return fmt.Errorf("cell.provenance.column_id does not match cell.column_id")
	}
	return nil
}
