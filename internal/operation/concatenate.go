package operation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

// ApplyConcatenate appends one derived string column without changing the
// source stream's existing values or metadata.
func ApplyConcatenate(sourceTable table.Table, source table.RowStream, options contract.TextOptions) (table.Table, table.RowStream, error) {
	if source == nil {
		return table.Table{}, nil, errors.New("concatenate source stream is required")
	}
	targetID := options.Target.Header
	if targetID == "" {
		targetID = options.Target.Alias
	}
	if strings.TrimSpace(targetID) == "" {
		return table.Table{}, nil, errors.New("concatenate target header or alias is required")
	}
	for _, column := range sourceTable.Columns {
		if column.ID == table.ColumnID(targetID) {
			return table.Table{}, nil, fmt.Errorf("concatenate target %q collides with existing column", targetID)
		}
	}
	indices := make([]int, len(options.Sources))
	for index, reference := range options.Sources {
		resolved, err := resolveTableColumn(sourceTable.Columns, reference)
		if err != nil {
			return table.Table{}, nil, fmt.Errorf("concatenate source[%d]: %w", index, err)
		}
		indices[index] = resolved
	}
	target := table.Column{
		ID: table.ColumnID(targetID), PhysicalPosition: "derived:" + targetID,
		OriginalHeader: targetID, Header: targetID, Type: table.KindString, TypeAuthority: table.TypeDeclared,
	}
	next := sourceTable
	next.Columns = append(append([]table.Column(nil), sourceTable.Columns...), target)
	if err := table.ValidateTable(next); err != nil {
		return table.Table{}, nil, fmt.Errorf("validate concatenated table: %w", err)
	}
	return next, &concatenateStream{
		source: source, targetColumn: target, sourceIndices: indices,
		separator: options.Separator, trim: options.Transform == "trim",
	}, nil
}

type concatenateStream struct {
	source        table.RowStream
	targetColumn  table.Column
	sourceIndices []int
	separator     string
	trim          bool
}

func (stream *concatenateStream) Next(ctx context.Context) (table.Row, error) {
	row, err := stream.source.Next(ctx)
	if err != nil {
		return table.Row{}, err
	}
	parts := make([]string, 0, len(stream.sourceIndices))
	for _, index := range stream.sourceIndices {
		value, ok := row.Cells[index].Value.StringContent()
		if !ok {
			if row.Cells[index].Value.IsNull() {
				continue
			}
			return table.Row{}, errors.New("concatenate source is not a string")
		}
		if stream.trim {
			value = strings.TrimSpace(value)
		}
		if value != "" {
			parts = append(parts, value)
		}
	}
	provenance := table.Provenance{
		SourceID: row.SourceID, SheetID: row.SheetID, RowID: row.ID, ColumnID: stream.targetColumn.ID,
	}
	row.Cells = append(row.Cells, table.Cell{
		ColumnID:   stream.targetColumn.ID,
		Value:      table.StringValue(strings.Join(parts, stream.separator)),
		Provenance: provenance,
	})
	return row, nil
}

func (stream *concatenateStream) Close() error {
	return stream.source.Close()
}
