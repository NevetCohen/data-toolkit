package operation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

// ApplyRename builds a deterministic metadata and row-stream transformation.
// Column order and source provenance are preserved; only the logical column
// identity in the transformed table and cells changes.
func ApplyRename(sourceTable table.Table, source table.RowStream, renames []contract.RenameField) (table.Table, table.RowStream, error) {
	if source == nil {
		return table.Table{}, nil, errors.New("rename source stream is required")
	}
	next := sourceTable
	next.Columns = append([]table.Column(nil), sourceTable.Columns...)

	type resolvedRename struct {
		index  int
		oldID  table.ColumnID
		target table.ColumnID
	}
	resolved := make([]resolvedRename, 0, len(renames))
	renamedIndices := make(map[int]struct{}, len(renames))
	targets := make(map[table.ColumnID]struct{}, len(renames))

	for index, rename := range renames {
		if strings.TrimSpace(rename.To) == "" {
			return table.Table{}, nil, fmt.Errorf("rename[%d] target is required", index)
		}
		columnIndex, err := resolveTableColumn(sourceTable.Columns, rename.Column)
		if err != nil {
			return table.Table{}, nil, fmt.Errorf("rename[%d]: %w", index, err)
		}
		if _, duplicate := renamedIndices[columnIndex]; duplicate {
			return table.Table{}, nil, fmt.Errorf("rename[%d] repeats source column %q", index, sourceTable.Columns[columnIndex].ID)
		}
		target := table.ColumnID(rename.To)
		if _, duplicate := targets[target]; duplicate {
			return table.Table{}, nil, fmt.Errorf("rename[%d] repeats target %q", index, rename.To)
		}
		renamedIndices[columnIndex] = struct{}{}
		targets[target] = struct{}{}
		resolved = append(resolved, resolvedRename{
			index: columnIndex, oldID: sourceTable.Columns[columnIndex].ID, target: target,
		})
	}

	for index, column := range sourceTable.Columns {
		if _, renamed := renamedIndices[index]; renamed {
			continue
		}
		if _, collision := targets[column.ID]; collision {
			return table.Table{}, nil, fmt.Errorf("rename target %q collides with unchanged column", column.ID)
		}
	}

	replacements := make(map[table.ColumnID]table.ColumnID, len(resolved))
	for _, rename := range resolved {
		next.Columns[rename.index].ID = rename.target
		next.Columns[rename.index].Header = string(rename.target)
		replacements[rename.oldID] = rename.target
	}
	if err := table.ValidateTable(next); err != nil {
		return table.Table{}, nil, fmt.Errorf("validate renamed table: %w", err)
	}
	return next, &renameStream{source: source, replacements: replacements}, nil
}

type renameStream struct {
	source       table.RowStream
	replacements map[table.ColumnID]table.ColumnID
}

func (stream *renameStream) Next(ctx context.Context) (table.Row, error) {
	row, err := stream.source.Next(ctx)
	if err != nil {
		return table.Row{}, err
	}
	row.Cells = append([]table.Cell(nil), row.Cells...)
	for index := range row.Cells {
		if replacement, exists := stream.replacements[row.Cells[index].ColumnID]; exists {
			row.Cells[index].ColumnID = replacement
			row.Cells[index].Provenance.ColumnID = replacement
		}
	}
	return row, nil
}

func (stream *renameStream) Close() error {
	return stream.source.Close()
}
