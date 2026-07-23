// Package column defines canonical column metadata without materialized values.
package column

import (
	"fmt"
	"strings"

	"data-toolkit/internal/core"
	"data-toolkit/internal/core/datatype"
)

type TypeAuthority string

const (
	TypeInferred TypeAuthority = "inferred"
	TypeDeclared TypeAuthority = "declared"
)

type Range struct {
	StartRow    uint64 `json:"start_row,omitempty"`
	EndRow      uint64 `json:"end_row,omitempty"`
	StartColumn string `json:"start_column,omitempty"`
	EndColumn   string `json:"end_column,omitempty"`
	Address     string `json:"address,omitempty"`
}

type SourceBinding struct {
	SourceID core.SourceID `json:"source_id"`
	SheetID  core.SheetID  `json:"sheet_id"`
	Kind     string        `json:"kind"`
	Locator  string        `json:"locator"`
}

// Descriptor is metadata only. Cell values remain in table.RowStream.
type Descriptor struct {
	ID            core.ColumnID `json:"id"`
	Header        string        `json:"header"`
	Subheader     string        `json:"subheader,omitempty"`
	DataType      datatype.ID   `json:"data_type"`
	TypeAuthority TypeAuthority `json:"type_authority"`
	Range         *Range        `json:"range,omitempty"`
	Binding       SourceBinding `json:"binding"`
}

func (descriptor Descriptor) Validate() error {
	if descriptor.ID == "" {
		return fmt.Errorf("column.id is required")
	}
	if strings.TrimSpace(descriptor.Header) == "" {
		return fmt.Errorf("column.header is required")
	}
	if descriptor.DataType == "" {
		return fmt.Errorf("column.data_type is required")
	}
	switch descriptor.TypeAuthority {
	case TypeInferred, TypeDeclared:
	default:
		return fmt.Errorf("column.type_authority %q is invalid", descriptor.TypeAuthority)
	}
	if descriptor.Binding.SourceID == "" {
		return fmt.Errorf("column.binding.source_id is required")
	}
	if strings.TrimSpace(descriptor.Binding.Kind) == "" || strings.TrimSpace(descriptor.Binding.Locator) == "" {
		return fmt.Errorf("column.binding kind and locator are required")
	}
	if descriptor.Range != nil {
		if descriptor.Range.EndRow != 0 && descriptor.Range.EndRow < descriptor.Range.StartRow {
			return fmt.Errorf("column.range end_row precedes start_row")
		}
		if descriptor.Range.EndColumn != "" && descriptor.Range.StartColumn == "" {
			return fmt.Errorf("column.range start_column is required when end_column is set")
		}
	}
	return nil
}
