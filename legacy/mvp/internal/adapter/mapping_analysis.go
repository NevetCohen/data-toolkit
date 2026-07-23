package adapter

import (
	"fmt"
	"strings"

	"data-toolkit/internal/table"
)

const (
	formatBoolean           = "boolean"
	formatInteger           = "integer"
	formatDecimalFixed      = "decimal:fixed"
	formatDecimalScientific = "decimal:scientific"
	formatDateISO           = "date:yyyy-mm-dd"
	formatDateDMYLong       = "date:dd/mm/yyyy"
	formatDateDMYShort      = "date:dd/mm/yy"
	formatTimeMinute        = "time:hh:mm"
	formatTimeSecond        = "time:hh:mm:ss"
	formatText              = "text"
	formatJSONObject        = "json:object"
	formatJSONArray         = "json:array"
)

func newColumnMapping(column table.Column) ColumnMapping {
	return ColumnMapping{
		Column:          column,
		ColumnID:        column.ID,
		TypeCounts:      make(map[table.ValueKind]int64),
		ObservedFormats: make(map[string]int64),
	}
}

func recordMappedValue(mapping *ColumnMapping, value table.Value, observedFormat string) {
	if value.IsNull() {
		mapping.MissingCount++
		return
	}
	mapping.ValueCount++
	mapping.TypeCounts[value.Kind()]++
	if observedFormat == "" {
		observedFormat = canonicalObservedFormat(value)
	}
	mapping.ObservedFormats[observedFormat]++
	if text, ok := value.StringContent(); ok && strings.TrimSpace(text) != text {
		mapping.WhitespaceCount++
	}
}

func finalizeMappingReport(report *MappingReport, canonicalTable table.Table) {
	for index := range canonicalTable.Columns {
		column := canonicalTable.Columns[index]
		mapping := &report.Columns[index]
		mapping.Column = column
		mapping.ColumnID = column.ID
		if mapping.WhitespaceCount > 0 {
			report.Findings = append(report.Findings, Finding{
				SourceID: report.SourceID,
				SheetID:  canonicalTable.Sheet.ID,
				ColumnID: column.ID,
				Code:     "whitespace",
				Path:     fmt.Sprintf("$.columns[%d]", index),
				Message:  "leading or trailing whitespace detected",
				Count:    mapping.WhitespaceCount,
			})
		}
		if len(mapping.TypeCounts) > 1 {
			report.Findings = append(report.Findings, Finding{
				SourceID: report.SourceID,
				SheetID:  canonicalTable.Sheet.ID,
				ColumnID: column.ID,
				Code:     "mixed-types",
				Path:     fmt.Sprintf("$.columns[%d]", index),
				Message:  "column contains multiple inferred canonical types",
				Count:    mapping.ValueCount,
			})
		}
	}
}

func canonicalObservedFormat(value table.Value) string {
	switch value.Kind() {
	case table.KindBoolean:
		return formatBoolean
	case table.KindInteger:
		return formatInteger
	case table.KindDecimal:
		lexeme, _ := value.RawNumericLexeme()
		if strings.ContainsAny(lexeme, "eE") {
			return formatDecimalScientific
		}
		return formatDecimalFixed
	case table.KindDate:
		return formatDateISO
	case table.KindTime:
		return formatTimeSecond
	case table.KindJSONObject:
		return formatJSONObject
	case table.KindJSONArray:
		return formatJSONArray
	default:
		return formatText
	}
}

func csvObservedFormat(raw string, value table.Value) string {
	switch value.Kind() {
	case table.KindDate:
		switch {
		case len(raw) == len("2006-01-02") && raw[4] == '-' && raw[7] == '-':
			return formatDateISO
		case len(raw) == len("02/01/2006"):
			return formatDateDMYLong
		default:
			return formatDateDMYShort
		}
	case table.KindTime:
		if len(raw) == len("15:04") {
			return formatTimeMinute
		}
		return formatTimeSecond
	default:
		return canonicalObservedFormat(value)
	}
}
