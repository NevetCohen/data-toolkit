package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

type JSONAdapter struct{}

func NewJSONAdapter() *JSONAdapter {
	return &JSONAdapter{}
}

func (adapter *JSONAdapter) Format() contract.DataFormat {
	return contract.FormatJSON
}

func (adapter *JSONAdapter) Capabilities() Capabilities {
	return Capabilities{
		CapabilityInspect: true, CapabilityMapBasic: true,
		CapabilityMapExtended: true, CapabilityRead: true,
	}
}

func (adapter *JSONAdapter) Inspect(ctx context.Context, request SourceRequest) (Inspection, error) {
	mappings, err := jsonMappings(ctx, request)
	if err != nil {
		return Inspection{}, err
	}
	name := "root"
	if request.Source.Options.JSONRoot != "" {
		segments, err := parseSimpleRootPointer(request.Source.Options.JSONRoot)
		if err != nil {
			return Inspection{}, err
		}
		name = segments[0]
	}
	_ = mappings
	return Inspection{
		SourceID: table.SourceID(request.Source.ID),
		Sheets: []table.Sheet{{
			ID: table.SheetID(request.Source.ID + ":json"), SourceID: table.SourceID(request.Source.ID), Name: name,
		}},
	}, nil
}

func (adapter *JSONAdapter) Map(ctx context.Context, request SourceRequest, mode MappingMode) (MappingReport, error) {
	mappings, err := jsonMappings(ctx, request)
	if err != nil {
		return MappingReport{}, err
	}
	canonicalTable := jsonTable(request.Source, mappings)
	report := MappingReport{
		SourceID: table.SourceID(request.Source.ID),
		Tables:   []table.Table{canonicalTable},
		Columns:  make([]ColumnMapping, len(mappings)),
	}
	for index, mapping := range mappings {
		report.Columns[index] = newColumnMapping(canonicalTable.Columns[index])
		report.Columns[index].JSONPointer = mapping.JSONPath
	}

	stream, err := adapter.OpenRows(ctx, request, canonicalTable)
	if err != nil {
		return MappingReport{}, err
	}
	defer stream.Close()
	for {
		row, err := stream.Next(ctx)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return MappingReport{}, fmt.Errorf("map JSON rows: %w", err)
		}
		for index, cell := range row.Cells {
			column := &report.Columns[index]
			recordMappedValue(column, cell.Value, canonicalObservedFormat(cell.Value))
			if !cell.Value.IsNull() && canonicalTable.Columns[index].Type == table.KindNull {
				canonicalTable.Columns[index].Type = cell.Value.Kind()
			}
		}
		if mode == MappingBasic {
			allResolved := true
			for _, column := range canonicalTable.Columns {
				if column.Type == table.KindNull {
					allResolved = false
					break
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

func (adapter *JSONAdapter) OpenRows(ctx context.Context, request SourceRequest, canonicalTable table.Table) (table.RowStream, error) {
	path := request.SnapshotPath
	if path == "" {
		path = request.Source.Location
	}
	file, decoder, err := openJSONRootArray(ctx, path, request.Source.Options.JSONRoot)
	if err != nil {
		return nil, err
	}
	mappings, err := jsonMappings(ctx, request)
	if err != nil {
		file.Close()
		return nil, err
	}
	if len(mappings) != len(canonicalTable.Columns) {
		file.Close()
		return nil, errors.New("JSON mapping count does not match canonical table")
	}
	projections := make([]table.JSONProjection, len(mappings))
	for index, mapping := range mappings {
		projections[index] = table.JSONProjection{Name: string(canonicalTable.Columns[index].ID), Pointer: mapping.JSONPath}
	}
	return &jsonRowStream{
		file: file, decoder: decoder, sourceID: table.SourceID(request.Source.ID),
		sheetID: canonicalTable.Sheet.ID, columns: canonicalTable.Columns, projections: projections,
	}, nil
}

func (adapter *JSONAdapter) NewWriter(context.Context, OutputRequest, table.Table) (RowWriter, error) {
	return nil, errors.New("JSON writing is not implemented")
}

func (adapter *JSONAdapter) ValidateOutput(context.Context, OutputRequest) error {
	return errors.New("JSON output validation is not implemented")
}

func jsonMappings(ctx context.Context, request SourceRequest) ([]contract.FieldMapping, error) {
	if len(request.Mappings) > 0 {
		result := append([]contract.FieldMapping(nil), request.Mappings...)
		for index := range result {
			if result[index].JSONPath == "" {
				header := result[index].Column.Header
				if header == "" {
					return nil, fmt.Errorf("mappings[%d]: JSON path or original header is required", index)
				}
				result[index].JSONPath = "/" + escapeJSONPointerSegment(header)
			}
		}
		return result, nil
	}

	path := request.SnapshotPath
	if path == "" {
		path = request.Source.Location
	}
	file, decoder, err := openJSONRootArray(ctx, path, request.Source.Options.JSONRoot)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if !decoder.More() {
		return nil, errors.New("JSON row array is empty and has no declared mappings")
	}
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("read first JSON row: %w", err)
	}
	keys, err := orderedTopLevelKeys(raw)
	if err != nil {
		return nil, err
	}
	mappings := make([]contract.FieldMapping, len(keys))
	for index, key := range keys {
		mappings[index] = contract.FieldMapping{
			SourceID: request.Source.ID,
			Column:   contract.ColumnRef{Header: key},
			JSONPath: "/" + escapeJSONPointerSegment(key),
		}
	}
	return mappings, nil
}

func jsonTable(source contract.Source, mappings []contract.FieldMapping) table.Table {
	sheetName := "root"
	if source.Options.JSONRoot != "" {
		if segments, err := parseSimpleRootPointer(source.Options.JSONRoot); err == nil && len(segments) == 1 {
			sheetName = segments[0]
		}
	}
	canonicalTable := table.Table{
		ID: table.TableID(source.ID + ":json"),
		Sheet: table.Sheet{
			ID: table.SheetID(source.ID + ":json"), SourceID: table.SourceID(source.ID), Name: sheetName,
		},
		Columns: make([]table.Column, len(mappings)),
	}
	for index, mapping := range mappings {
		header := mapping.Column.Header
		if header == "" {
			segments := strings.Split(mapping.JSONPath, "/")
			header = strings.ReplaceAll(strings.ReplaceAll(segments[len(segments)-1], "~1", "/"), "~0", "~")
		}
		id := header
		if mapping.Alias != "" {
			id = mapping.Alias
		}
		canonicalTable.Columns[index] = table.Column{
			ID: table.ColumnID(id), PhysicalPosition: mapping.JSONPath,
			OriginalHeader: header, GlobalAlias: mapping.Alias,
			Type: table.KindNull, TypeAuthority: table.TypeInferred, Header: header,
		}
	}
	return canonicalTable
}

type jsonRowStream struct {
	file        *os.File
	decoder     *json.Decoder
	sourceID    table.SourceID
	sheetID     table.SheetID
	columns     []table.Column
	projections []table.JSONProjection
	ordinal     uint64
	closed      bool
}

func (stream *jsonRowStream) Next(ctx context.Context) (table.Row, error) {
	if stream.closed {
		return table.Row{}, errors.New("JSON row stream is closed")
	}
	if err := ctx.Err(); err != nil {
		return table.Row{}, err
	}
	if !stream.decoder.More() {
		return table.Row{}, io.EOF
	}
	var raw json.RawMessage
	if err := stream.decoder.Decode(&raw); err != nil {
		return table.Row{}, fmt.Errorf("decode JSON row %d: %w", stream.ordinal+1, err)
	}
	projected, err := table.ProjectJSON(raw, stream.projections)
	if err != nil {
		return table.Row{}, fmt.Errorf("project JSON row %d: %w", stream.ordinal+1, err)
	}
	rowID := table.RowID(string(stream.sourceID) + ":row:" + strconv.FormatUint(stream.ordinal+1, 10))
	row := table.Row{
		ID: rowID, SourceID: stream.sourceID, SheetID: stream.sheetID,
		Ordinal: stream.ordinal, Cells: make([]table.Cell, len(projected)),
	}
	for index, value := range projected {
		row.Cells[index] = table.Cell{
			ColumnID: stream.columns[index].ID,
			Value:    value.Value,
			Provenance: table.Provenance{
				SourceID: stream.sourceID, SheetID: stream.sheetID, RowID: rowID,
				ColumnID: stream.columns[index].ID, PhysicalRow: int(stream.ordinal) + 1,
				JSONPointer: value.Pointer,
			},
		}
	}
	stream.ordinal++
	return row, nil
}

func (stream *jsonRowStream) Close() error {
	if stream.closed {
		return nil
	}
	stream.closed = true
	return stream.file.Close()
}

func openJSONRootArray(ctx context.Context, path, rootPointer string) (*os.File, *json.Decoder, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open JSON source %q: %w", path, err)
	}
	decoder := json.NewDecoder(file)
	decoder.UseNumber()
	token, err := decoder.Token()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("read JSON root: %w", err)
	}
	if rootPointer == "" {
		if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
			file.Close()
			return nil, nil, errors.New("JSON root must be an array or json_root must select a top-level array")
		}
		return file, decoder, nil
	}
	segments, err := parseSimpleRootPointer(rootPointer)
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		file.Close()
		return nil, nil, errors.New("json_root requires an object root")
	}
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			file.Close()
			return nil, nil, fmt.Errorf("read JSON root key: %w", err)
		}
		key, ok := keyToken.(string)
		if !ok {
			file.Close()
			return nil, nil, errors.New("JSON object key is not a string")
		}
		if key == segments[0] {
			arrayToken, err := decoder.Token()
			if err != nil {
				file.Close()
				return nil, nil, fmt.Errorf("read selected JSON array: %w", err)
			}
			if delimiter, ok := arrayToken.(json.Delim); !ok || delimiter != '[' {
				file.Close()
				return nil, nil, fmt.Errorf("json_root %q does not select an array", rootPointer)
			}
			return file, decoder, nil
		}
		var skipped json.RawMessage
		if err := decoder.Decode(&skipped); err != nil {
			file.Close()
			return nil, nil, fmt.Errorf("skip JSON root field %q: %w", key, err)
		}
	}
	file.Close()
	return nil, nil, fmt.Errorf("json_root %q was not found", rootPointer)
}

func parseSimpleRootPointer(pointer string) ([]string, error) {
	if !strings.HasPrefix(pointer, "/") || strings.Count(pointer, "/") != 1 {
		return nil, errors.New("json_root currently supports exactly one top-level JSON Pointer segment")
	}
	segment := strings.TrimPrefix(pointer, "/")
	segment = strings.ReplaceAll(strings.ReplaceAll(segment, "~1", "/"), "~0", "~")
	if segment == "" {
		return nil, errors.New("json_root segment is empty")
	}
	return []string{segment}, nil
}

func orderedTopLevelKeys(raw []byte) ([]string, error) {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return nil, errors.New("JSON row must be an object")
	}
	keys := make([]string, 0)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key := token.(string)
		keys = append(keys, key)
		var skipped json.RawMessage
		if err := decoder.Decode(&skipped); err != nil {
			return nil, err
		}
	}
	return keys, nil
}

func escapeJSONPointerSegment(segment string) string {
	return strings.ReplaceAll(strings.ReplaceAll(segment, "~", "~0"), "/", "~1")
}
