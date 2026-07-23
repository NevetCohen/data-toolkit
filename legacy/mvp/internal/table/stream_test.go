package table

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
)

func TestBoundedSpoolPreservesStableOrderUnicodeAndProvenance(t *testing.T) {
	rows := []Row{
		testRow("row-1", 0, "נועה 🧭"),
		testRow("row-2", 1, "עברית"),
		testRow("row-3", 2, "🙂"),
	}
	spool, err := NewBoundedSpool(t.TempDir(), 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if err := spool.Append(row); err != nil {
			t.Fatal(err)
		}
	}
	if !spool.Spilled() {
		t.Fatal("expected rows to spill under one-byte memory budget")
	}
	spoolPath := spool.path
	stream, err := spool.Stream()
	if err != nil {
		t.Fatal(err)
	}
	for index, want := range rows {
		got, err := stream.Next(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != want.ID || got.Ordinal != want.Ordinal ||
			got.Cells[0].Provenance.RowID != want.ID {
			t.Fatalf("row[%d] identity changed: %#v", index, got)
		}
		text, _ := got.Cells[0].Value.StringContent()
		wantText, _ := want.Cells[0].Value.StringContent()
		if text != wantText {
			t.Fatalf("row[%d] text = %q, want %q", index, text, wantText)
		}
	}
	if _, err := stream.Next(context.Background()); !errors.Is(err, io.EOF) {
		t.Fatalf("end error = %v, want io.EOF", err)
	}
	if err := spool.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(spoolPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("spool file was not removed: %v", err)
	}
}

func TestCanonicalRowSerializationPreservesUnicodeAndProvenance(t *testing.T) {
	before := testRow("row-10", 9, "שלום 🙂")
	encoded, err := json.Marshal(before)
	if err != nil {
		t.Fatal(err)
	}
	var after Row
	if err := json.Unmarshal(encoded, &after); err != nil {
		t.Fatal(err)
	}
	if after.ID != before.ID || after.SourceID != before.SourceID ||
		after.SheetID != before.SheetID || after.Ordinal != before.Ordinal ||
		after.Cells[0].Provenance != before.Cells[0].Provenance {
		t.Fatalf("row identity or provenance changed: %#v", after)
	}
	text, _ := after.Cells[0].Value.StringContent()
	if text != "שלום 🙂" {
		t.Fatalf("Unicode changed: %q", text)
	}
}

func TestSliceStreamHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stream := NewSliceRowStream([]Row{testRow("row", 0, "value")})
	if _, err := stream.Next(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func testRow(id RowID, ordinal uint64, text string) Row {
	return Row{
		ID: id, SourceID: "source", SheetID: "sheet", Ordinal: ordinal,
		Cells: []Cell{{
			ColumnID: "name",
			Value:    StringValue(text),
			Provenance: Provenance{
				SourceID: "source", SheetID: "sheet", RowID: id, ColumnID: "name",
				PhysicalRow: int(ordinal) + 2, PhysicalColumn: "A",
			},
		}},
	}
}
