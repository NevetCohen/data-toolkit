// Package report provides deterministic ordering and formatting for execution
// diagnostics.
package report

import (
	"sort"

	"data-toolkit/internal/contract"
)

func OrderExceptions(exceptions []contract.ExceptionRecord) []contract.ExceptionRecord {
	ordered := append([]contract.ExceptionRecord(nil), exceptions...)
	sort.SliceStable(ordered, func(left, right int) bool {
		a, b := ordered[left], ordered[right]
		if a.SourceOrdinal != b.SourceOrdinal {
			return a.SourceOrdinal < b.SourceOrdinal
		}
		if a.RowOrdinal != b.RowOrdinal {
			return a.RowOrdinal < b.RowOrdinal
		}
		if a.ColumnOrdinal != b.ColumnOrdinal {
			return a.ColumnOrdinal < b.ColumnOrdinal
		}
		return a.Sequence < b.Sequence
	})
	return ordered
}
