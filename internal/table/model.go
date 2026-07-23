package table

type SourceID string
type SheetID string
type TableID string
type ColumnID string
type RowID string

type Sheet struct {
	ID       SheetID  `json:"id"`
	SourceID SourceID `json:"source_id"`
	Name     string   `json:"name"`
	Index    int      `json:"index"`
}

// Table is metadata for a canonical row stream; it does not force rows into
// memory.
type Table struct {
	ID      TableID  `json:"id"`
	Sheet   Sheet    `json:"sheet"`
	Columns []Column `json:"columns"`
}

type TypeAuthority string

const (
	TypeInferred TypeAuthority = "inferred"
	TypeDeclared TypeAuthority = "declared"
)

type Column struct {
	ID               ColumnID      `json:"id"`
	PhysicalPosition string        `json:"physical_position"`
	OriginalHeader   string        `json:"original_header"`
	GlobalAlias      string        `json:"global_alias,omitempty"`
	Type             ValueKind     `json:"type"`
	TypeAuthority    TypeAuthority `json:"type_authority"`
	Header           string        `json:"header,omitempty"`
	Subheader        string        `json:"subheader,omitempty"`
}

type Row struct {
	ID       RowID    `json:"id"`
	SourceID SourceID `json:"source_id"`
	SheetID  SheetID  `json:"sheet_id"`
	Ordinal  uint64   `json:"ordinal"`
	Cells    []Cell   `json:"cells"`
}

type Cell struct {
	ColumnID   ColumnID   `json:"column_id"`
	Value      Value      `json:"value"`
	Provenance Provenance `json:"provenance"`
}

type Provenance struct {
	SourceID       SourceID `json:"source_id"`
	SheetID        SheetID  `json:"sheet_id"`
	RowID          RowID    `json:"row_id"`
	ColumnID       ColumnID `json:"column_id"`
	PhysicalRow    int      `json:"physical_row,omitempty"`
	PhysicalColumn string   `json:"physical_column,omitempty"`
	JSONPointer    string   `json:"json_pointer,omitempty"`
}
