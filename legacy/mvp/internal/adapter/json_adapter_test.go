package adapter

import (
	"context"
	"errors"
	"io"
	"testing"

	"data-toolkit/internal/contract"
	"data-toolkit/internal/table"
)

func TestJSONAdapterMapsExplicitPathsAndStreamsExactRows(t *testing.T) {
	request := SourceRequest{
		Source: contract.Source{
			ID: "json", Format: contract.FormatJSON, Location: "testdata/json/rows.json",
			Options: contract.SourceOptions{JSONRoot: "/Data"},
		},
		Mappings: []contract.FieldMapping{
			{SourceID: "json", Column: contract.ColumnRef{Header: "name"}, JSONPath: "/name"},
			{SourceID: "json", Column: contract.ColumnRef{Header: "id"}, JSONPath: "/id"},
			{SourceID: "json", Column: contract.ColumnRef{Header: "amount"}, JSONPath: "/amount"},
			{SourceID: "json", Column: contract.ColumnRef{Header: "city"}, JSONPath: "/nested/city"},
		},
	}
	selected := NewJSONAdapter()
	report, err := selected.Map(context.Background(), request, MappingExtended)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Tables) != 1 || len(report.Columns) != 4 {
		t.Fatalf("mapping report = %#v", report)
	}
	if report.Columns[0].ValueCount != 2 || report.Columns[0].MissingCount != 1 {
		t.Fatalf("name counts = %#v", report.Columns[0])
	}
	if report.Columns[2].TypeCounts[table.KindDecimal] != 2 || report.Columns[2].MissingCount != 1 {
		t.Fatalf("amount counts = %#v", report.Columns[2])
	}
	if report.Columns[2].ObservedFormats[formatDecimalFixed] != 2 ||
		report.Columns[0].Column != report.Tables[0].Columns[0] {
		t.Fatalf("extended JSON mapping = %#v", report.Columns)
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
	if first.Ordinal != 0 || first.Cells[0].Provenance.JSONPointer != "/name" {
		t.Fatalf("first row provenance = %#v", first)
	}
	name, _ := first.Cells[0].Value.StringContent()
	if name != "נועה 🧭" {
		t.Fatalf("name = %q", name)
	}
	id, _ := first.Cells[1].Value.RawNumericLexeme()
	amount, _ := first.Cells[2].Value.RawNumericLexeme()
	if id != "9007199254740993" || amount != "0.1000000000000000000000000001" {
		t.Fatalf("exact numbers = %q, %q", id, amount)
	}
	for index := 1; index < 3; index++ {
		row, err := stream.Next(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if row.Ordinal != uint64(index) {
			t.Fatalf("row ordinal = %d, want %d", row.Ordinal, index)
		}
	}
	if _, err := stream.Next(context.Background()); !errors.Is(err, io.EOF) {
		t.Fatalf("end error = %v", err)
	}
}
