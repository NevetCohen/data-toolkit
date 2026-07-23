package adapter

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"reflect"
	"testing"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
	"github.com/xuri/excelize/v2"
)

func TestExcelAdapterInspectsAndStreamsMVP1FixtureInStableOrder(t *testing.T) {
	fixture := filepath.Join("..", "..", "testdata", "fixtures", "mvp1-excel-city-filter", "all_final.xlsx")
	request := SourceRequest{Source: contract.Source{ID: "mvp1", Format: contract.FormatExcel, Location: filepath.Join(t.TempDir(), "must-not-be-read.xlsx"), Sheets: []contract.SheetSelection{{Name: "all_final"}}}, SnapshotPath: fixture}
	selected := NewExcelAdapter()
	inspection, err := selected.Inspect(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	wantSheets := []table.Sheet{{ID: "mvp1:excel:0", SourceID: "mvp1", Name: "all_final", Index: 0}}
	if inspection.SourceID != "mvp1" || !reflect.DeepEqual(inspection.Sheets, wantSheets) {
		t.Fatalf("inspection = %#v, want sheets %#v", inspection, wantSheets)
	}
	report, err := selected.Map(context.Background(), request, MappingBasic)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Tables) != 1 || len(report.Tables[0].Columns) != 7 {
		t.Fatalf("mapping report = %#v", report)
	}
	wantHeaders := []string{"תעודת זהות", "שם פרטי ", "שם משפחה", "כתובת מייל", "כתובת", "ישוב", "טלפון נייד"}
	for i, column := range report.Tables[0].Columns {
		if column.PhysicalPosition != spreadsheetColumnName(i+1) || column.OriginalHeader != wantHeaders[i] || column.Header != wantHeaders[i] {
			t.Fatalf("column[%d] = %#v", i, column)
		}
	}
	if report.Tables[0].Columns[5].PhysicalPosition != "F" || report.Tables[0].Columns[5].OriginalHeader != "ישוב" {
		t.Fatalf("city column = %#v", report.Tables[0].Columns[5])
	}
	stream, err := selected.OpenRows(context.Background(), request, report.Tables[0])
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	wantCities := []string{"fixture-non-match", "בת ים", "בת ים", "fixture-non-match", "בת ים"}
	for i, wantCity := range wantCities {
		row, err := stream.Next(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		city, ok := row.Cells[5].Value.StringContent()
		if !ok || city != wantCity || row.Ordinal != uint64(i) {
			t.Fatalf("row[%d] = %#v, city = %q", i, row, city)
		}
		if row.Cells[5].Provenance.PhysicalRow != i+2 || row.Cells[5].Provenance.PhysicalColumn != "F" || row.SheetID != "mvp1:excel:0" {
			t.Fatalf("row[%d] provenance = %#v", i, row.Cells[5].Provenance)
		}
	}
	if _, err := stream.Next(context.Background()); !errors.Is(err, io.EOF) {
		t.Fatalf("end error = %v, want io.EOF", err)
	}
}

func TestExcelAdapterSelectsByZeroBasedIndexAndPreservesPhysicalColumns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "physical-columns.xlsx")
	book := excelize.NewFile()
	book.SetSheetName("Sheet1", "ignored")
	if _, err := book.NewSheet("selected"); err != nil {
		t.Fatal(err)
	}
	for cell, value := range map[string]string{"C1": "name", "E1": "city", "C2": "נועה 🧭", "E2": "בת ים", "C4": "דן", "E4": "חיפה"} {
		if err := book.SetCellValue("selected", cell, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := book.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	if err := book.Close(); err != nil {
		t.Fatal(err)
	}
	index := 1
	request := SourceRequest{Source: contract.Source{ID: "physical", Format: contract.FormatExcel, Location: path, Sheets: []contract.SheetSelection{{Name: "selected", Index: &index}}}}
	report, err := NewExcelAdapter().Map(context.Background(), request, MappingBasic)
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{"C", "E"} {
		if report.Tables[0].Columns[i].PhysicalPosition != want {
			t.Fatalf("column[%d] position = %q, want %q", i, report.Tables[0].Columns[i].PhysicalPosition, want)
		}
	}
	stream, err := NewExcelAdapter().OpenRows(context.Background(), request, report.Tables[0])
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	for ordinal, physicalRow := range []int{2, 3, 4} {
		row, err := stream.Next(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if row.Ordinal != uint64(ordinal) || row.Cells[0].Provenance.PhysicalRow != physicalRow {
			t.Fatalf("row[%d] = %#v", ordinal, row)
		}
		if physicalRow == 3 && (!row.Cells[0].Value.IsNull() || !row.Cells[1].Value.IsNull()) {
			t.Fatalf("blank row was not preserved as nulls: %#v", row)
		}
	}
	if _, err := stream.Next(context.Background()); !errors.Is(err, io.EOF) {
		t.Fatalf("end error = %v, want io.EOF", err)
	}
}

func TestExcelAdapterRejectsAmbiguousOrInvalidSheetSelection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sheets.xlsx")
	book := excelize.NewFile()
	if _, err := book.NewSheet("second"); err != nil {
		t.Fatal(err)
	}
	if err := book.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	if err := book.Close(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		sheets []contract.SheetSelection
	}{
		{name: "multiple sheets require selection"},
		{name: "unknown name", sheets: []contract.SheetSelection{{Name: "missing"}}},
		{name: "multiple selections", sheets: []contract.SheetSelection{{Name: "Sheet1"}, {Name: "second"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := SourceRequest{Source: contract.Source{ID: "sheets", Format: contract.FormatExcel, Location: path, Sheets: test.sheets}}
			if _, err := NewExcelAdapter().Map(context.Background(), request, MappingBasic); err == nil {
				t.Fatal("expected sheet selection error")
			}
		})
	}
}

func TestExcelAdapterCapabilitiesRemainReadOnlyAndFocused(t *testing.T) {
	capabilities := NewExcelAdapter().Capabilities()
	if !capabilities.Supports(CapabilityInspect, CapabilityMapBasic, CapabilityRead, CapabilitySheets) {
		t.Fatalf("capabilities = %#v", capabilities)
	}
	if capabilities.Supports(CapabilityMapExtended, CapabilityWrite, CapabilityValidate, CapabilityStyles) {
		t.Fatalf("unexpected capability: %#v", capabilities)
	}
}
