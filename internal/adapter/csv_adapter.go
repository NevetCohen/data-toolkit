package adapter

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

type CSVAdapter struct{}

func NewCSVAdapter() *CSVAdapter {
	return &CSVAdapter{}
}

func (adapter *CSVAdapter) Format() contract.DataFormat {
	return contract.FormatCSV
}

func (adapter *CSVAdapter) Capabilities() Capabilities {
	return Capabilities{
		CapabilityInspect: true, CapabilityMapBasic: true, CapabilityMapExtended: true,
		CapabilityRead: true, CapabilityWrite: true, CapabilityValidate: true,
	}
}

func (adapter *CSVAdapter) Inspect(_ context.Context, request SourceRequest) (Inspection, error) {
	file, reader, headers, err := openCSVSource(request)
	if err != nil {
		return Inspection{}, err
	}
	file.Close()
	_ = reader
	_ = headers
	return Inspection{
		SourceID: table.SourceID(request.Source.ID),
		Sheets: []table.Sheet{{
			ID: table.SheetID(request.Source.ID + ":csv"), SourceID: table.SourceID(request.Source.ID), Name: "csv",
		}},
	}, nil
}

func (adapter *CSVAdapter) Map(ctx context.Context, request SourceRequest, mode MappingMode) (MappingReport, error) {
	file, reader, headers, err := openCSVSource(request)
	if err != nil {
		return MappingReport{}, err
	}
	defer file.Close()
	canonicalTable, err := csvTable(request.Source.ID, headers, request.Mappings)
	if err != nil {
		return MappingReport{}, err
	}
	report := MappingReport{
		SourceID: table.SourceID(request.Source.ID), Tables: []table.Table{canonicalTable},
		Columns: make([]ColumnMapping, len(headers)),
	}
	for index := range report.Columns {
		report.Columns[index] = newColumnMapping(canonicalTable.Columns[index])
	}
	for rowNumber := 2; ; rowNumber++ {
		if err := ctx.Err(); err != nil {
			return MappingReport{}, err
		}
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return MappingReport{}, fmt.Errorf("read CSV row %d: %w", rowNumber, err)
		}
		if len(record) > len(headers) {
			return MappingReport{}, fmt.Errorf("CSV row %d has %d fields, header has %d", rowNumber, len(record), len(headers))
		}
		for index := range headers {
			if index >= len(record) || record[index] == "" {
				recordMappedValue(&report.Columns[index], table.NullValue(), "")
				continue
			}
			if !utf8.ValidString(record[index]) {
				return MappingReport{}, fmt.Errorf("CSV row %d column %d is not valid UTF-8", rowNumber, index+1)
			}
			value, err := inferCSVValue(record[index])
			if err != nil {
				return MappingReport{}, err
			}
			recordMappedValue(&report.Columns[index], value, csvObservedFormat(record[index], value))
			if canonicalTable.Columns[index].Type == table.KindNull {
				canonicalTable.Columns[index].Type = value.Kind()
			}
		}
		if mode == MappingBasic {
			allResolved := true
			for _, column := range canonicalTable.Columns {
				if column.Type == table.KindNull {
					allResolved = false
				}
			}
			if allResolved {
				break
			}
		}
	}
	for index := range canonicalTable.Columns {
		if canonicalTable.Columns[index].Type == table.KindNull {
			canonicalTable.Columns[index].Type = table.KindString
		}
	}
	report.Tables[0] = canonicalTable
	finalizeMappingReport(&report, canonicalTable)
	return report, nil
}

func (adapter *CSVAdapter) OpenRows(_ context.Context, request SourceRequest, canonicalTable table.Table) (table.RowStream, error) {
	file, reader, headers, err := openCSVSource(request)
	if err != nil {
		return nil, err
	}
	if len(headers) != len(canonicalTable.Columns) {
		file.Close()
		return nil, errors.New("CSV header count does not match canonical table")
	}
	return &csvRowStream{
		file: file, reader: reader, sourceID: table.SourceID(request.Source.ID),
		sheetID: canonicalTable.Sheet.ID, columns: canonicalTable.Columns,
	}, nil
}

func (adapter *CSVAdapter) NewWriter(_ context.Context, request OutputRequest, canonicalTable table.Table) (RowWriter, error) {
	file, err := os.OpenFile(request.StagedPath, os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open staged CSV output: %w", err)
	}
	indices := make([]int, len(request.Output.Projection))
	headers := make([]string, len(request.Output.Projection))
	for projectionIndex, projection := range request.Output.Projection {
		columnIndex, err := resolveProjectionColumn(canonicalTable.Columns, projection.Column)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("output projection[%d]: %w", projectionIndex, err)
		}
		indices[projectionIndex] = columnIndex
		headers[projectionIndex] = projection.OutputAs
	}
	writer := csv.NewWriter(file)
	if err := writer.Write(headers); err != nil {
		file.Close()
		return nil, fmt.Errorf("write CSV header: %w", err)
	}
	return &csvRowWriter{file: file, writer: writer, indices: indices, render: request.Render}, nil
}

func (adapter *CSVAdapter) ValidateOutput(_ context.Context, request OutputRequest) error {
	file, err := os.Open(request.StagedPath)
	if err != nil {
		return fmt.Errorf("open staged CSV for validation: %w", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read staged CSV header: %w", err)
	}
	if len(headers) != len(request.Output.Projection) {
		return fmt.Errorf("CSV header count = %d, want %d", len(headers), len(request.Output.Projection))
	}
	for index, projection := range request.Output.Projection {
		if headers[index] != projection.OutputAs {
			return fmt.Errorf("CSV header[%d] = %q, want %q", index, headers[index], projection.OutputAs)
		}
	}
	var rows int64
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read staged CSV row %d: %w", rows+2, err)
		}
		if len(record) != len(headers) {
			return fmt.Errorf("CSV row %d field count = %d, want %d", rows+2, len(record), len(headers))
		}
		for index, value := range record {
			if !utf8.ValidString(value) {
				return fmt.Errorf("CSV row %d field %d is not valid UTF-8", rows+2, index+1)
			}
		}
		rows++
	}
	for _, validation := range request.Output.Validations {
		if validation.Kind == contract.ValidationRowCount {
			expected, err := strconv.ParseInt(validation.Expected, 10, 64)
			if err != nil {
				return fmt.Errorf("row-count validation %q has invalid expected value: %w", validation.ID, err)
			}
			if rows != expected {
				return fmt.Errorf("row-count validation %q failed: got %d, want %d", validation.ID, rows, expected)
			}
		}
	}
	return nil
}

func openCSVSource(request SourceRequest) (*os.File, *csv.Reader, []string, error) {
	path := request.SnapshotPath
	if path == "" {
		path = request.Source.Location
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open CSV source %q: %w", path, err)
	}
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	headers, err := reader.Read()
	if err != nil {
		file.Close()
		return nil, nil, nil, fmt.Errorf("read CSV header: %w", err)
	}
	if len(headers) > 0 {
		headers[0] = strings.TrimPrefix(headers[0], "\uFEFF")
	}
	seen := make(map[string]struct{}, len(headers))
	for index, header := range headers {
		if !utf8.ValidString(header) {
			file.Close()
			return nil, nil, nil, fmt.Errorf("CSV header %d is not valid UTF-8", index+1)
		}
		if header == "" {
			file.Close()
			return nil, nil, nil, fmt.Errorf("CSV header %d is empty", index+1)
		}
		if _, exists := seen[header]; exists {
			file.Close()
			return nil, nil, nil, fmt.Errorf("CSV header %q is duplicated", header)
		}
		seen[header] = struct{}{}
	}
	return file, reader, headers, nil
}

func csvTable(sourceID string, headers []string, mappings []contract.FieldMapping) (table.Table, error) {
	canonicalTable := table.Table{
		ID: table.TableID(sourceID + ":csv"),
		Sheet: table.Sheet{
			ID: table.SheetID(sourceID + ":csv"), SourceID: table.SourceID(sourceID), Name: "csv",
		},
		Columns: make([]table.Column, len(headers)),
	}
	for index, header := range headers {
		canonicalTable.Columns[index] = table.Column{
			ID: table.ColumnID(header), PhysicalPosition: spreadsheetColumnName(index + 1),
			OriginalHeader: header, Header: header, Type: table.KindNull, TypeAuthority: table.TypeInferred,
		}
	}
	resolved := make(map[int]struct{}, len(mappings))
	for mappingIndex, mapping := range mappings {
		if mapping.SourceID != "" && mapping.SourceID != sourceID {
			return table.Table{}, fmt.Errorf("mappings[%d].source_id: got %q, want %q", mappingIndex, mapping.SourceID, sourceID)
		}
		if mapping.Sheet != "" && mapping.Sheet != canonicalTable.Sheet.Name {
			return table.Table{}, fmt.Errorf("mappings[%d].sheet: got %q, want %q", mappingIndex, mapping.Sheet, canonicalTable.Sheet.Name)
		}
		columnIndex, err := resolveProjectionColumn(canonicalTable.Columns, mapping.Column)
		if err != nil {
			return table.Table{}, fmt.Errorf("mappings[%d].column: %w", mappingIndex, err)
		}
		if _, exists := resolved[columnIndex]; exists {
			return table.Table{}, fmt.Errorf("mappings[%d].column: column is mapped more than once", mappingIndex)
		}
		resolved[columnIndex] = struct{}{}
		if mapping.Alias != "" {
			canonicalTable.Columns[columnIndex].GlobalAlias = mapping.Alias
			canonicalTable.Columns[columnIndex].ID = table.ColumnID(mapping.Alias)
		}
	}
	if err := table.ValidateTable(canonicalTable); err != nil {
		return table.Table{}, fmt.Errorf("validate CSV basic mapping: %w", err)
	}
	return canonicalTable, nil
}

type csvRowStream struct {
	file     *os.File
	reader   *csv.Reader
	sourceID table.SourceID
	sheetID  table.SheetID
	columns  []table.Column
	ordinal  uint64
	closed   bool
}

func (stream *csvRowStream) Next(ctx context.Context) (table.Row, error) {
	if stream.closed {
		return table.Row{}, errors.New("CSV row stream is closed")
	}
	if err := ctx.Err(); err != nil {
		return table.Row{}, err
	}
	record, err := stream.reader.Read()
	if err != nil {
		return table.Row{}, err
	}
	if len(record) > len(stream.columns) {
		return table.Row{}, fmt.Errorf("CSV row %d has %d fields, want at most %d", stream.ordinal+2, len(record), len(stream.columns))
	}
	rowID := table.RowID(string(stream.sourceID) + ":row:" + strconv.FormatUint(stream.ordinal+1, 10))
	row := table.Row{
		ID: rowID, SourceID: stream.sourceID, SheetID: stream.sheetID,
		Ordinal: stream.ordinal, Cells: make([]table.Cell, len(stream.columns)),
	}
	for index, column := range stream.columns {
		value := table.NullValue()
		if index < len(record) && record[index] != "" {
			value, err = inferCSVValue(record[index])
			if err != nil {
				return table.Row{}, err
			}
		}
		row.Cells[index] = table.Cell{
			ColumnID: column.ID, Value: value,
			Provenance: table.Provenance{
				SourceID: stream.sourceID, SheetID: stream.sheetID, RowID: rowID, ColumnID: column.ID,
				PhysicalRow: int(stream.ordinal) + 2, PhysicalColumn: column.PhysicalPosition,
			},
		}
	}
	stream.ordinal++
	return row, nil
}

func (stream *csvRowStream) Close() error {
	if stream.closed {
		return nil
	}
	stream.closed = true
	return stream.file.Close()
}

type csvRowWriter struct {
	file    *os.File
	writer  *csv.Writer
	indices []int
	render  table.RenderOptions
	closed  bool
}

func (writer *csvRowWriter) WriteRow(ctx context.Context, row table.Row) error {
	if writer.closed {
		return errors.New("CSV writer is closed")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	record := make([]string, len(writer.indices))
	for outputIndex, cellIndex := range writer.indices {
		if cellIndex >= len(row.Cells) {
			return fmt.Errorf("row %q has no cell at projection index %d", row.ID, cellIndex)
		}
		rendered, err := table.Render(row.Cells[cellIndex].Value, writer.render)
		if err != nil {
			return fmt.Errorf("render row %q cell %d: %w", row.ID, cellIndex, err)
		}
		record[outputIndex] = rendered
	}
	if err := writer.writer.Write(record); err != nil {
		return fmt.Errorf("write CSV row %q: %w", row.ID, err)
	}
	return nil
}

func (writer *csvRowWriter) Close() error {
	if writer.closed {
		return nil
	}
	writer.closed = true
	writer.writer.Flush()
	writeError := writer.writer.Error()
	closeError := writer.file.Close()
	if writeError != nil {
		return writeError
	}
	return closeError
}

func inferCSVValue(raw string) (table.Value, error) {
	if raw == "true" || raw == "false" {
		return table.BooleanValue(raw == "true"), nil
	}
	if integer, err := table.IntegerValue(raw); err == nil {
		return integer, nil
	}
	if decimal, err := table.DecimalValue(raw); err == nil {
		return decimal, nil
	}
	for _, layout := range []string{"2006-01-02", "02/01/2006", "02/01/06"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return table.DateValue(parsed.Year(), parsed.Month(), parsed.Day())
		}
	}
	for _, layout := range []string{"15:04", "15:04:05"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return table.TimeValue(parsed.Hour(), parsed.Minute(), parsed.Second(), parsed.Nanosecond())
		}
	}
	return table.StringValue(raw), nil
}

func resolveProjectionColumn(columns []table.Column, reference contract.ColumnRef) (int, error) {
	matches := make([]int, 0, 1)
	for index, column := range columns {
		if (reference.Position != "" && column.PhysicalPosition == reference.Position) ||
			(reference.Header != "" && (column.OriginalHeader == reference.Header || column.Header == reference.Header)) ||
			(reference.Alias != "" && column.GlobalAlias == reference.Alias) {
			matches = append(matches, index)
		}
	}
	if len(matches) != 1 {
		return 0, fmt.Errorf("column reference resolved to %d columns", len(matches))
	}
	return matches[0], nil
}

func spreadsheetColumnName(position int) string {
	var name []byte
	for position > 0 {
		position--
		name = append([]byte{byte('A' + position%26)}, name...)
		position /= 26
	}
	return string(name)
}
