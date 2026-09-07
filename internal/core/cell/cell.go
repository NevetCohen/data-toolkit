// Package cell defines canonical cells.
package cell

import (
	"fmt"

	"data-toolkit/internal/core"
)

type Cell struct {
	ColumnID core.ColumnID `json:"column_id"`
	Value    core.Value    `json:"-"`
}

func (value Cell) Validate() error {
	if value.ColumnID == "" {
		return fmt.Errorf("cell.column_id is required")
	}
	if value.Value.TypeID() == "" {
		return fmt.Errorf("cell.value type is required")
	}
	return nil
}
