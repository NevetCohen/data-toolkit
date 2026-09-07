// Package table defines canonical table metadata, rows, and deterministic
// row-stream transport.
package table

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/cell"
	"data-toolkit/internal/core/column"
)

type Schema struct {
	ID       core.TableID              `json:"id"`
	SourceID core.SourceID             `json:"source_id"`
	SheetID  *string                   `json:"sheet_id,omitempty"`
	Columns  []column.ColumnDescriptor `json:"columns"`
}

func (schema Schema) Validate() error {
	if err := schema.ID.Validate(); err != nil {
		return fmt.Errorf("schema.id: %w", err)
	}
	if err := schema.SourceID.Validate(); err != nil {
		return fmt.Errorf("schema.source_id: %w", err)
	}
	if schema.SheetID != nil {
		if err := core.SheetID(*schema.SheetID).Validate(); err != nil {
			return fmt.Errorf("schema.sheet_id: %w", err)
		}
	}
	if len(schema.Columns) == 0 {
		return errors.New("schema.columns must be non-empty")
	}
	seen := make(map[core.ColumnID]struct{}, len(schema.Columns))
	for index, descriptor := range schema.Columns {
		if err := descriptor.Validate(); err != nil {
			return fmt.Errorf("schema.columns[%d]: %w", index, err)
		}
		if _, exists := seen[descriptor.ID]; exists {
			return fmt.Errorf("schema.columns[%d]: duplicate column %q", index, descriptor.ID)
		}
		seen[descriptor.ID] = struct{}{}
	}
	return nil
}

type Row struct {
	ID       core.RowID    `json:"id"`
	SourceID core.SourceID `json:"source_id"`
	SheetID  *string       `json:"sheet_id,omitempty"`
	Ordinal  uint64        `json:"ordinal"`
	Cells    []cell.Cell   `json:"-"`
}

func (row Row) Validate(schema Schema) error {
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("row schema: %w", err)
	}
	if err := row.ID.Validate(); err != nil {
		return fmt.Errorf("row.id: %w", err)
	}
	if err := row.SourceID.Validate(); err != nil {
		return fmt.Errorf("row.source_id: %w", err)
	}
	if row.SheetID != nil {
		if err := core.SheetID(*row.SheetID).Validate(); err != nil {
			return fmt.Errorf("row.sheet_id: %w", err)
		}
	}
	if row.SourceID != schema.SourceID {
		return fmt.Errorf("row.source_id %q does not match schema.source_id %q", row.SourceID, schema.SourceID)
	}
	if !sheetIDsMatch(row.SheetID, schema.SheetID) {
		return fmt.Errorf("row.sheet_id does not match schema.sheet_id")
	}
	if len(row.Cells) != len(schema.Columns) {
		return fmt.Errorf("row has %d cells; table has %d columns", len(row.Cells), len(schema.Columns))
	}
	for index, value := range row.Cells {
		if err := value.Validate(); err != nil {
			return fmt.Errorf("row.cells[%d]: %w", index, err)
		}
		if value.ColumnID != schema.Columns[index].ID {
			return fmt.Errorf("row.cells[%d] column %q does not match table column %q", index, value.ColumnID, schema.Columns[index].ID)
		}
		if value.Value.TypeID() != schema.Columns[index].DataType {
			return fmt.Errorf("row.cells[%d] type %q does not match schema column type %q", index, value.Value.TypeID(), schema.Columns[index].DataType)
		}
	}
	return nil
}

func sheetIDsMatch(rowSheetID, schemaSheetID *string) bool {
	if rowSheetID == nil || schemaSheetID == nil {
		return rowSheetID == schemaSheetID
	}
	return *rowSheetID == *schemaSheetID
}

type RowStream interface {
	Next(context.Context) (Row, error)
	Close() error
}

type Dataset struct {
	Schema Schema    `json:"-"`
	Rows   RowStream `json:"-"`
}

// Validate confirms that a dataset has a valid schema before its stream is used.
func (dataset Dataset) Validate() error {
	if err := dataset.Schema.Validate(); err != nil {
		return fmt.Errorf("dataset.schema: %w", err)
	}
	if dataset.Rows == nil || isNilRowStream(dataset.Rows) {
		return errors.New("dataset.rows must be a non-nil row stream")
	}
	return nil
}

func isNilRowStream(stream RowStream) bool {
	value := reflect.ValueOf(stream)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

type SliceStream struct {
	rows   []Row
	index  int
	closed bool
}

func NewSliceStream(rows []Row) *SliceStream {
	return &SliceStream{rows: append([]Row(nil), rows...)}
}

func (stream *SliceStream) Next(ctx context.Context) (Row, error) {
	if stream.closed {
		return Row{}, errors.New("row stream is closed")
	}
	if err := ctx.Err(); err != nil {
		return Row{}, err
	}
	if stream.index >= len(stream.rows) {
		return Row{}, io.EOF
	}
	row := stream.rows[stream.index]
	stream.index++
	return row, nil
}

func (stream *SliceStream) Close() error {
	stream.closed = true
	return nil
}
