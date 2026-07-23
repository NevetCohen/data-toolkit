package adapter

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

func TestCSVAdapterMapsHeadersMissingFieldsUnicodeAndStableOrder(t *testing.T) {
	request := SourceRequest{Source: contract.Source{
		ID: "csv", Format: contract.FormatCSV, Location: "testdata/csv/rows.csv",
	}}
	selected := NewCSVAdapter()
	report, err := selected.Map(context.Background(), request, MappingExtended)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Tables) != 1 || len(report.Tables[0].Columns) != 3 {
		t.Fatalf("mapping report = %#v", report)
	}
	if report.Tables[0].Columns[0].OriginalHeader != "name" ||
		report.Tables[0].Columns[0].PhysicalPosition != "A" ||
		report.Columns[2].MissingCount != 1 {
		t.Fatalf("column mapping = %#v %#v", report.Tables[0].Columns, report.Columns)
	}
	stream, err := selected.OpenRows(context.Background(), request, report.Tables[0])
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	first, err := stream.Next(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	name, _ := first.Cells[0].Value.StringContent()
	id, _ := first.Cells[1].Value.RawNumericLexeme()
	if first.Ordinal != 0 || name != "נועה 🧭" || id != "9007199254740993" {
		t.Fatalf("first row = %#v", first)
	}
	second, err := stream.Next(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if second.Ordinal != 1 || !second.Cells[2].Value.IsNull() {
		t.Fatalf("second row = %#v", second)
	}
	if _, err := stream.Next(context.Background()); !errors.Is(err, io.EOF) {
		t.Fatalf("end error = %v", err)
	}
}

func TestCSVAdapterBasicMappingPreservesCompleteColumnIdentity(t *testing.T) {
	request := SourceRequest{
		Source: contract.Source{ID: "csv", Format: contract.FormatCSV, Location: "testdata/csv/rows.csv"},
		Mappings: []contract.FieldMapping{
			{SourceID: "csv", Sheet: "csv", Column: contract.ColumnRef{Position: "B"}, Alias: "MemberID"},
			{SourceID: "csv", Column: contract.ColumnRef{Header: "city"}, Alias: "City"},
		},
	}
	report, err := NewCSVAdapter().Map(context.Background(), request, MappingBasic)
	if err != nil {
		t.Fatal(err)
	}
	if report.SourceID != "csv" || len(report.Tables) != 1 {
		t.Fatalf("mapping identity = %#v", report)
	}
	mapped := report.Tables[0]
	if mapped.Sheet.SourceID != "csv" || mapped.Sheet.Name != "csv" || len(mapped.Columns) != 3 {
		t.Fatalf("table identity = %#v", mapped)
	}
	wantPositions := []string{"A", "B", "C"}
	wantHeaders := []string{"name", "id", "city"}
	wantAliases := []string{"", "MemberID", "City"}
	wantIDs := []table.ColumnID{"name", "MemberID", "City"}
	for index, column := range mapped.Columns {
		if column.PhysicalPosition != wantPositions[index] || column.OriginalHeader != wantHeaders[index] ||
			column.GlobalAlias != wantAliases[index] || column.ID != wantIDs[index] ||
			column.Type == table.KindNull || column.TypeAuthority != table.TypeInferred {
			t.Fatalf("column[%d] = %#v", index, column)
		}
	}
}

func TestCSVAdapterBasicMappingRejectsAmbiguousOrWrongSheetMappings(t *testing.T) {
	tests := []struct {
		name     string
		mappings []contract.FieldMapping
	}{
		{
			name: "wrong sheet",
			mappings: []contract.FieldMapping{{
				SourceID: "csv", Sheet: "Sheet1", Column: contract.ColumnRef{Header: "name"}, Alias: "Name",
			}},
		},
		{
			name: "same column twice",
			mappings: []contract.FieldMapping{
				{SourceID: "csv", Column: contract.ColumnRef{Header: "name"}, Alias: "Name"},
				{SourceID: "csv", Column: contract.ColumnRef{Position: "A"}, Alias: "OtherName"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := SourceRequest{
				Source:   contract.Source{ID: "csv", Format: contract.FormatCSV, Location: "testdata/csv/rows.csv"},
				Mappings: test.mappings,
			}
			if _, err := NewCSVAdapter().Map(context.Background(), request, MappingBasic); err == nil {
				t.Fatal("expected invalid mapping error")
			}
		})
	}
}

func TestCSVAdapterExtendedMappingReportsCountsFormatsMetadataAndCleanliness(t *testing.T) {
	request := SourceRequest{Source: contract.Source{
		ID: "csv", Format: contract.FormatCSV, Location: "testdata/csv/extended_mapping.csv",
	}}
	report, err := NewCSVAdapter().Map(context.Background(), request, MappingExtended)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Tables) != 1 || len(report.Columns) != 5 {
		t.Fatalf("extended mapping = %#v", report)
	}
	date := report.Columns[1]
	if date.ValueCount != 3 || date.MissingCount != 0 || date.TypeCounts[table.KindDate] != 3 ||
		date.ObservedFormats[formatDateISO] != 1 || date.ObservedFormats[formatDateDMYLong] != 1 ||
		date.ObservedFormats[formatDateDMYShort] != 1 {
		t.Fatalf("date mapping = %#v", date)
	}
	clock := report.Columns[2]
	if clock.ValueCount != 2 || clock.MissingCount != 1 ||
		clock.ObservedFormats[formatTimeMinute] != 1 || clock.ObservedFormats[formatTimeSecond] != 1 {
		t.Fatalf("time mapping = %#v", clock)
	}
	amount := report.Columns[3]
	if amount.TypeCounts[table.KindDecimal] != 3 || amount.ObservedFormats[formatDecimalFixed] != 2 ||
		amount.ObservedFormats[formatDecimalScientific] != 1 {
		t.Fatalf("amount mapping = %#v", amount)
	}
	mixed := report.Columns[4]
	if mixed.ValueCount != 2 || mixed.MissingCount != 1 || mixed.TypeCounts[table.KindInteger] != 1 ||
		mixed.TypeCounts[table.KindString] != 1 {
		t.Fatalf("mixed mapping = %#v", mixed)
	}
	for index, mapping := range report.Columns {
		column := report.Tables[0].Columns[index]
		if mapping.Column != column || mapping.ColumnID != column.ID || mapping.Column.OriginalHeader == "" ||
			mapping.Column.PhysicalPosition == "" || mapping.Column.Header == "" {
			t.Fatalf("column metadata[%d] = %#v, table column = %#v", index, mapping, column)
		}
	}
	assertFinding := func(code string, column table.ColumnID, count int64) {
		t.Helper()
		for _, finding := range report.Findings {
			if finding.Code == code && finding.ColumnID == column {
				if finding.SourceID != "csv" || finding.SheetID != "csv:csv" || finding.Count != count {
					t.Fatalf("finding %q = %#v", code, finding)
				}
				return
			}
		}
		t.Fatalf("finding %q for column %q was not reported: %#v", code, column, report.Findings)
	}
	assertFinding("whitespace", "name", 1)
	assertFinding("mixed-types", "mixed", 2)
}

func TestCSVWriterUsesExplicitProjectionOrderAndValidatesStructure(t *testing.T) {
	sourceRequest := SourceRequest{Source: contract.Source{
		ID: "csv", Format: contract.FormatCSV, Location: "testdata/csv/rows.csv",
	}}
	selected := NewCSVAdapter()
	report, err := selected.Map(context.Background(), sourceRequest, MappingExtended)
	if err != nil {
		t.Fatal(err)
	}
	stream, err := selected.OpenRows(context.Background(), sourceRequest, report.Tables[0])
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	finalPath := filepath.Join(t.TempDir(), "projected.csv")
	stage, err := NewStagedOutput(finalPath)
	if err != nil {
		t.Fatal(err)
	}
	output := contract.Output{
		ID: "projected", Input: "csv", Format: contract.FormatCSV, Location: finalPath,
		Projection: []contract.ProjectionField{
			{Column: contract.ColumnRef{Header: "city"}, OutputAs: "ישוב"},
			{Column: contract.ColumnRef{Header: "name"}, OutputAs: "שם"},
		},
		Validations: []contract.Validation{{ID: "rows", Kind: contract.ValidationRowCount, Expected: "2"}},
	}
	request := OutputRequest{
		Output: output, StagedPath: stage.Path(),
		Render: table.RenderOptions{NullDisplay: "-", DecimalPrecision: 3, DateFormat: "02/01/06", TimeFormat: "15:04"},
	}
	writer, err := selected.NewWriter(context.Background(), request, report.Tables[0])
	if err != nil {
		t.Fatal(err)
	}
	for {
		row, err := stream.Next(context.Background())
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := writer.WriteRow(context.Background(), row); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := selected.ValidateOutput(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if err := stage.Validate(context.Background(), func(ctx context.Context, _ string) error {
		return selected.ValidateOutput(ctx, request)
	}); err != nil {
		t.Fatal(err)
	}
	published, err := stage.Publish(contract.CollisionBlock)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(published)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if rows[0][0] != "ישוב" || rows[0][1] != "שם" ||
		rows[1][0] != "בת ים" || rows[1][1] != "נועה 🧭" ||
		rows[2][0] != "-" {
		t.Fatalf("projected CSV = %#v", rows)
	}
}
