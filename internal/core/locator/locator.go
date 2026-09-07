// Package locator defines precise diagnostic locations for source and output
// failures without adding provenance to canonical cells.
package locator

import "data-toolkit/internal/core"

// SourceLocator identifies an optional precise location in an input source.
// Optional fields are omitted when the corresponding source dimension is not
// applicable or cannot be identified.
type SourceLocator struct {
	SourceID         core.SourceID  `json:"source_id"`
	SheetID          *core.SheetID  `json:"sheet_id,omitempty"`
	RootPath         *string        `json:"root_path,omitempty"`
	RowOrdinal       *uint64        `json:"row_ordinal,omitempty"`
	ColumnID         *core.ColumnID `json:"column_id,omitempty"`
	PhysicalPosition *string        `json:"physical_position,omitempty"`
}

// OutputLocator identifies an optional precise location in reopened output.
// It is used for output validation failures and never persists per-cell
// provenance in the canonical data model.
type OutputLocator struct {
	OutputPath       string         `json:"output_path"`
	RowOrdinal       *uint64        `json:"row_ordinal,omitempty"`
	ColumnID         *core.ColumnID `json:"column_id,omitempty"`
	PhysicalPosition *string        `json:"physical_position,omitempty"`
}
