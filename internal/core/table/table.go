// Package table defines canonical table metadata, rows, and deterministic
// row-stream transport.
package table

import (
	"context"
	"errors"
	"fmt"
	"io"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/cell"
	"data-toolkit/internal/core/column"
)

type Schema struct {
	ID       core.TableID        `json:"id"`
	SourceID core.SourceID       `json:"source_id"`
	SheetID  core.SheetID        `json:"sheet_id"`
	Columns  []column.Descriptor `json:"columns"`
}

func (schema Schema) Validate() error {
	if schema.ID == "" || schema.SourceID == "" {
		return fmt.Errorf("table id and source_id are required")
	}
	seen := make(map[core.ColumnID]struct{}, len(schema.Columns))
	for index, descriptor := range schema.Columns {
		if err := descriptor.Validate(); err != nil {
			return fmt.Errorf("table.columns[%d]: %w", index, err)
		}
		if _, exists := seen[descriptor.ID]; exists {
			return fmt.Errorf("table.columns[%d]: duplicate column %q", index, descriptor.ID)
		}
		seen[descriptor.ID] = struct{}{}
	}
	return nil
}

type Row struct {
	ID       core.RowID    `json:"id"`
	SourceID core.SourceID `json:"source_id"`
	SheetID  core.SheetID  `json:"sheet_id"`
	Ordinal  uint64        `json:"ordinal"`
	Cells    []cell.Cell   `json:"cells"`
}

func (row Row) Validate(schema Schema) error {
	if row.ID == "" || row.SourceID == "" {
		return fmt.Errorf("row id and source_id are required")
	}
	if row.SourceID != schema.SourceID || row.SheetID != schema.SheetID {
		return fmt.Errorf("row source or sheet identity does not match table")
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
	}
	return nil
}

type RowStream interface {
	Next(context.Context) (Row, error)
	Close() error
}

type Dataset struct {
	Schema Schema
	Rows   RowStream
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
