package report

import (
	"testing"

	"data-toolkit/internal/contract"
)

func TestExceptionsAreOrderedByStableSourceProvenance(t *testing.T) {
	exceptions := []contract.ExceptionRecord{
		{Reference: "third", SourceOrdinal: 1, RowOrdinal: 0, ColumnOrdinal: 0, Sequence: 2},
		{Reference: "second", SourceOrdinal: 0, RowOrdinal: 4, ColumnOrdinal: 2, Sequence: 1},
		{Reference: "first", SourceOrdinal: 0, RowOrdinal: 4, ColumnOrdinal: 1, Sequence: 0},
	}
	ordered := OrderExceptions(exceptions)
	for index, want := range []string{"first", "second", "third"} {
		if ordered[index].Reference != want {
			t.Fatalf("ordered[%d] = %q, want %q", index, ordered[index].Reference, want)
		}
	}
}
