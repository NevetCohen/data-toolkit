package adapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
	"github.com/xuri/excelize/v2"
)

// ExcelAdapter converts one selected workbook sheet into canonical rows.
// Formula warnings, output writing, and output validation are later tasks.
type ExcelAdapter struct{}

func NewExcelAdapter() *ExcelAdapter { return &ExcelAdapter{} }

func (*ExcelAdapter) Format() contract.DataFormat { return contract.FormatExcel }

func (*ExcelAdapter) Capabilities() Capabilities {
	return Capabilities{
		CapabilityInspect: true, CapabilityMapBasic: true,
		CapabilityRead: true, CapabilitySheets: true,
	}
}

func (*ExcelAdapter) Inspect(ctx context.Context, request SourceRequest) (Inspection, error) {
	if err := ctx.Err(); err != nil {
		return Inspection{}, err
	}
	book, err := openExcelSource(request)
	if err != nil {
		return Inspection{}, err
	}
	defer book.Close()
	names := book.GetSheetList()
	sheets := make([]table.Sheet, len(names))
	for i, name := range names {
		sheets[i] = excelSheet(request.Source.ID, name, i)
	}
	return Inspection{SourceID: table.SourceID(request.Source.ID), Sheets: sheets}, nil
}

func (*ExcelAdapter) Map(ctx context.Context, request SourceRequest, mode MappingMode) (MappingReport, error) {
	if mode != MappingBasic {
		return MappingReport{}, fmt.Errorf("Excel mapping mode %q is not supported", mode)
	}
	if err := ctx.Err(); err != nil {
		return MappingReport{}, err
	}
	book, err := openExcelSource(request)
	if err != nil {
		return MappingReport{}, err
	}
	defer book.Close()
	selected, err := selectExcelSheet(request.Source, book.GetSheetList())
	if err != nil {
		return MappingReport{}, err
	}
	rows, err := book.Rows(selected.Name)
	if err != nil {
		return MappingReport{}, fmt.Errorf("open Excel sheet %q rows: %w", selected.Name, err)
	}
	defer rows.Close()
	headers, headerRow, err := readExcelHeader(ctx, rows)
	if err != nil {
		return MappingReport{}, fmt.Errorf("read Excel sheet %q header: %w", selected.Name, err)
	}
	canonical, sourceColumns, err := buildExcelTable(request, selected, headers)
	if err != nil {
		return MappingReport{}, err
	}
	report := MappingReport{SourceID: table.SourceID(request.Source.ID), Tables: []table.Table{canonical}, Columns: make([]ColumnMapping, len(canonical.Columns))}
	for i := range report.Columns {
		report.Columns[i] = newColumnMapping(canonical.Columns[i])
	}
	physicalRow := headerRow
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return MappingReport{}, err
		}
		physicalRow++
		values, err := rows.Columns(excelize.Options{RawCellValue: true})
		if err != nil {
			return MappingReport{}, fmt.Errorf("read Excel sheet %q row %d: %w", selected.Name, physicalRow, err)
		}
		for i, sourceColumn := range sourceColumns {
			value, err := readExcelValue(book, selected.Name, physicalRow, sourceColumn, values)
			if err != nil {
				return MappingReport{}, err
			}
			recordMappedValue(&report.Columns[i], value, canonicalObservedFormat(value))
			if !value.IsNull() && canonical.Columns[i].Type == table.KindNull {
				canonical.Columns[i].Type = value.Kind()
			}
		}
		if excelTypesResolved(canonical.Columns) {
			break
		}
	}
	if err := rows.Error(); err != nil {
		return MappingReport{}, fmt.Errorf("iterate Excel sheet %q: %w", selected.Name, err)
	}
	for i := range canonical.Columns {
		if canonical.Columns[i].Type == table.KindNull {
			canonical.Columns[i].Type = table.KindString
		}
	}
	report.Tables[0] = canonical
	finalizeMappingReport(&report, canonical)
	return report, nil
}

func (*ExcelAdapter) OpenRows(ctx context.Context, request SourceRequest, canonical table.Table) (table.RowStream, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	book, err := openExcelSource(request)
	if err != nil {
		return nil, err
	}
	selected, err := selectExcelSheet(request.Source, book.GetSheetList())
	if err != nil {
		book.Close()
		return nil, err
	}
	if canonical.Sheet != selected {
		book.Close()
		return nil, fmt.Errorf("Excel selected sheet %q does not match canonical table sheet %q", selected.Name, canonical.Sheet.Name)
	}
	rows, err := book.Rows(selected.Name)
	if err != nil {
		book.Close()
		return nil, fmt.Errorf("open Excel sheet %q rows: %w", selected.Name, err)
	}
	headers, headerRow, err := readExcelHeader(ctx, rows)
	if err != nil {
		rows.Close()
		book.Close()
		return nil, fmt.Errorf("read Excel sheet %q header: %w", selected.Name, err)
	}
	sourceColumns, err := validateExcelTable(canonical, headers)
	if err != nil {
		rows.Close()
		book.Close()
		return nil, err
	}
	return &excelRowStream{book: book, rows: rows, sourceID: table.SourceID(request.Source.ID), sheet: selected, columns: append([]table.Column(nil), canonical.Columns...), sourceColumns: sourceColumns, physicalRow: headerRow}, nil
}

func (*ExcelAdapter) NewWriter(context.Context, OutputRequest, table.Table) (RowWriter, error) {
	return nil, errors.New("Excel writing is not implemented")
}
func (*ExcelAdapter) ValidateOutput(context.Context, OutputRequest) error {
	return errors.New("Excel output validation is not implemented")
}

type excelHeader struct {
	position int
	value    string
}

func openExcelSource(request SourceRequest) (*excelize.File, error) {
	path := request.SnapshotPath
	if path == "" {
		path = request.Source.Location
	}
	if path == "" {
		return nil, errors.New("Excel source path is required")
	}
	book, err := excelize.OpenFile(path, excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, fmt.Errorf("open Excel source %q: %w", path, err)
	}
	return book, nil
}

func excelSheet(sourceID, name string, index int) table.Sheet {
	return table.Sheet{ID: table.SheetID(sourceID + ":excel:" + strconv.Itoa(index)), SourceID: table.SourceID(sourceID), Name: name, Index: index}
}

func selectExcelSheet(source contract.Source, names []string) (table.Sheet, error) {
	if len(names) == 0 {
		return table.Sheet{}, errors.New("Excel workbook contains no sheets")
	}
	if len(source.Sheets) == 0 {
		if len(names) != 1 {
			return table.Sheet{}, errors.New("Excel workbook has multiple sheets; select exactly one sheet by name or zero-based index")
		}
		return excelSheet(source.ID, names[0], 0), nil
	}
	if len(source.Sheets) != 1 {
		return table.Sheet{}, fmt.Errorf("Excel row reading requires exactly one sheet selection, got %d", len(source.Sheets))
	}
	selection := source.Sheets[0]
	if selection.Name == "" && selection.Index == nil {
		return table.Sheet{}, errors.New("Excel sheet selection requires a name or zero-based index")
	}
	resolved := -1
	if selection.Index != nil {
		resolved = *selection.Index
		if resolved < 0 || resolved >= len(names) {
			return table.Sheet{}, fmt.Errorf("Excel sheet index %d is outside [0,%d)", resolved, len(names))
		}
	}
	if selection.Name != "" {
		nameIndex := -1
		for i, name := range names {
			if name == selection.Name {
				nameIndex = i
				break
			}
		}
		if nameIndex < 0 {
			return table.Sheet{}, fmt.Errorf("Excel sheet %q does not exist", selection.Name)
		}
		if resolved >= 0 && nameIndex != resolved {
			return table.Sheet{}, fmt.Errorf("Excel sheet name %q and index %d select different sheets", selection.Name, resolved)
		}
		resolved = nameIndex
	}
	return excelSheet(source.ID, names[resolved], resolved), nil
}

func readExcelHeader(ctx context.Context, rows *excelize.Rows) ([]excelHeader, int, error) {
	physicalRow := 0
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return nil, 0, err
		}
		physicalRow++
		values, err := rows.Columns(excelize.Options{RawCellValue: true})
		if err != nil {
			return nil, 0, err
		}
		headers := make([]excelHeader, 0, len(values))
		for i, value := range values {
			if value != "" {
				headers = append(headers, excelHeader{position: i + 1, value: value})
			}
		}
		if len(headers) > 0 {
			return headers, physicalRow, nil
		}
	}
	if err := rows.Error(); err != nil {
		return nil, 0, err
	}
	return nil, 0, errors.New("Excel sheet has no non-empty header row")
}

func buildExcelTable(request SourceRequest, selected table.Sheet, headers []excelHeader) (table.Table, []int, error) {
	canonical := table.Table{ID: table.TableID(request.Source.ID + ":excel:" + strconv.Itoa(selected.Index)), Sheet: selected, Columns: make([]table.Column, len(headers))}
	sourceColumns := make([]int, len(headers))
	for i, header := range headers {
		canonical.Columns[i] = table.Column{ID: table.ColumnID(header.value), PhysicalPosition: spreadsheetColumnName(header.position), OriginalHeader: header.value, Type: table.KindNull, TypeAuthority: table.TypeInferred, Header: header.value}
		sourceColumns[i] = header.position
	}
	resolved := make(map[int]struct{}, len(request.Mappings))
	for mappingIndex, mapping := range request.Mappings {
		if mapping.SourceID != "" && mapping.SourceID != request.Source.ID {
			return table.Table{}, nil, fmt.Errorf("mappings[%d].source_id: got %q, want %q", mappingIndex, mapping.SourceID, request.Source.ID)
		}
		if mapping.Sheet != "" && mapping.Sheet != selected.Name {
			return table.Table{}, nil, fmt.Errorf("mappings[%d].sheet: got %q, want %q", mappingIndex, mapping.Sheet, selected.Name)
		}
		columnIndex, err := resolveProjectionColumn(canonical.Columns, mapping.Column)
		if err != nil {
			return table.Table{}, nil, fmt.Errorf("mappings[%d].column: %w", mappingIndex, err)
		}
		if _, exists := resolved[columnIndex]; exists {
			return table.Table{}, nil, fmt.Errorf("mappings[%d].column: column is mapped more than once", mappingIndex)
		}
		resolved[columnIndex] = struct{}{}
		if mapping.Alias != "" {
			canonical.Columns[columnIndex].GlobalAlias = mapping.Alias
			canonical.Columns[columnIndex].ID = table.ColumnID(mapping.Alias)
		}
	}
	if err := table.ValidateTable(canonical); err != nil {
		return table.Table{}, nil, fmt.Errorf("validate Excel basic mapping: %w", err)
	}
	return canonical, sourceColumns, nil
}

func validateExcelTable(canonical table.Table, headers []excelHeader) ([]int, error) {
	if len(headers) != len(canonical.Columns) {
		return nil, fmt.Errorf("Excel header count %d does not match canonical table column count %d", len(headers), len(canonical.Columns))
	}
	sourceColumns := make([]int, len(headers))
	for i, header := range headers {
		position := spreadsheetColumnName(header.position)
		if canonical.Columns[i].PhysicalPosition != position || canonical.Columns[i].OriginalHeader != header.value {
			return nil, fmt.Errorf("Excel header %d is %q at %s, canonical table has %q at %s", i, header.value, position, canonical.Columns[i].OriginalHeader, canonical.Columns[i].PhysicalPosition)
		}
		sourceColumns[i] = header.position
	}
	return sourceColumns, nil
}

func excelTypesResolved(columns []table.Column) bool {
	for _, column := range columns {
		if column.Type == table.KindNull {
			return false
		}
	}
	return true
}

func readExcelValue(book *excelize.File, sheet string, physicalRow, sourceColumn int, values []string) (table.Value, error) {
	if sourceColumn > len(values) || values[sourceColumn-1] == "" {
		return table.NullValue(), nil
	}
	raw := values[sourceColumn-1]
	cell, err := excelize.CoordinatesToCellName(sourceColumn, physicalRow)
	if err != nil {
		return table.Value{}, fmt.Errorf("resolve Excel cell at row %d column %d: %w", physicalRow, sourceColumn, err)
	}
	cellType, err := book.GetCellType(sheet, cell)
	if err != nil {
		return table.Value{}, fmt.Errorf("read Excel cell type %s!%s: %w", sheet, cell, err)
	}
	value, err := canonicalExcelValue(raw, cellType)
	if err != nil {
		return table.Value{}, fmt.Errorf("read Excel value %s!%s: %w", sheet, cell, err)
	}
	return value, nil
}

func canonicalExcelValue(raw string, cellType excelize.CellType) (table.Value, error) {
	switch cellType {
	case excelize.CellTypeBool:
		switch strings.ToLower(raw) {
		case "1", "true":
			return table.BooleanValue(true), nil
		case "0", "false":
			return table.BooleanValue(false), nil
		default:
			return table.Value{}, fmt.Errorf("invalid Boolean value %q", raw)
		}
	case excelize.CellTypeNumber:
		if value, err := table.IntegerValue(raw); err == nil {
			return value, nil
		}
		return table.DecimalValue(raw)
	case excelize.CellTypeDate:
		for _, layout := range []string{time.RFC3339Nano, "2006-01-02"} {
			if parsed, err := time.Parse(layout, raw); err == nil {
				return table.DateValue(parsed.Year(), parsed.Month(), parsed.Day())
			}
		}
		return table.StringValue(raw), nil
	case excelize.CellTypeSharedString, excelize.CellTypeInlineString, excelize.CellTypeError:
		return table.StringValue(raw), nil
	case excelize.CellTypeFormula, excelize.CellTypeUnset:
		if value, err := table.IntegerValue(raw); err == nil {
			return value, nil
		}
		if value, err := table.DecimalValue(raw); err == nil {
			return value, nil
		}
		return table.StringValue(raw), nil
	default:
		return table.StringValue(raw), nil
	}
}

type excelRowStream struct {
	book          *excelize.File
	rows          *excelize.Rows
	sourceID      table.SourceID
	sheet         table.Sheet
	columns       []table.Column
	sourceColumns []int
	physicalRow   int
	ordinal       uint64
	closed        bool
}

func (stream *excelRowStream) Next(ctx context.Context) (table.Row, error) {
	if stream.closed {
		return table.Row{}, errors.New("Excel row stream is closed")
	}
	if err := ctx.Err(); err != nil {
		return table.Row{}, err
	}
	if !stream.rows.Next() {
		if err := stream.rows.Error(); err != nil {
			return table.Row{}, fmt.Errorf("iterate Excel sheet %q: %w", stream.sheet.Name, err)
		}
		return table.Row{}, io.EOF
	}
	stream.physicalRow++
	values, err := stream.rows.Columns(excelize.Options{RawCellValue: true})
	if err != nil {
		return table.Row{}, fmt.Errorf("read Excel sheet %q row %d: %w", stream.sheet.Name, stream.physicalRow, err)
	}
	rowID := table.RowID(string(stream.sourceID) + ":sheet:" + strconv.Itoa(stream.sheet.Index) + ":row:" + strconv.Itoa(stream.physicalRow))
	row := table.Row{ID: rowID, SourceID: stream.sourceID, SheetID: stream.sheet.ID, Ordinal: stream.ordinal, Cells: make([]table.Cell, len(stream.columns))}
	for i, column := range stream.columns {
		value, err := readExcelValue(stream.book, stream.sheet.Name, stream.physicalRow, stream.sourceColumns[i], values)
		if err != nil {
			return table.Row{}, err
		}
		row.Cells[i] = table.Cell{ColumnID: column.ID, Value: value, Provenance: table.Provenance{SourceID: stream.sourceID, SheetID: stream.sheet.ID, RowID: rowID, ColumnID: column.ID, PhysicalRow: stream.physicalRow, PhysicalColumn: column.PhysicalPosition}}
	}
	stream.ordinal++
	return row, nil
}

func (stream *excelRowStream) Close() error {
	if stream.closed {
		return nil
	}
	stream.closed = true
	rowsErr := stream.rows.Close()
	bookErr := stream.book.Close()
	if rowsErr != nil {
		return rowsErr
	}
	return bookErr
}
