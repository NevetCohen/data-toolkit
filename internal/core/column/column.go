// Package column defines canonical column metadata without materialized values.
package column

import (
	"fmt"
	"unicode/utf8"

	"data-toolkit/internal/core"

	"golang.org/x/text/unicode/norm"
)

// ColumnDescriptor is canonical schema metadata. Cell values remain in the row stream.
type ColumnDescriptor struct {
	ID               core.ColumnID      `json:"id"`
	PhysicalPosition string             `json:"physical_position"`
	SourceHeader     string             `json:"source_header"`
	Subheader        *string            `json:"subheader,omitempty"`
	Alias            *string            `json:"alias,omitempty"`
	SemanticType     *string            `json:"semantic_type,omitempty"`
	DataType         core.DataTypeID    `json:"data_type"`
	TypeAuthority    core.TypeAuthority `json:"type_authority"`
}

// Descriptor is retained as a compatibility alias while table construction is
// migrated to the exact ColumnDescriptor name.
type Descriptor = ColumnDescriptor

func (descriptor ColumnDescriptor) Validate() error {
	if err := descriptor.ID.Validate(); err != nil {
		return fmt.Errorf("column.id: %w", err)
	}
	if err := validateString("column.physical_position", descriptor.PhysicalPosition, true); err != nil {
		return err
	}
	if err := validateString("column.source_header", descriptor.SourceHeader, false); err != nil {
		return err
	}
	optionalFields := []struct {
		name  string
		value *string
	}{
		{name: "column.subheader", value: descriptor.Subheader},
		{name: "column.alias", value: descriptor.Alias},
		{name: "column.semantic_type", value: descriptor.SemanticType},
	}
	for _, field := range optionalFields {
		if field.value != nil {
			if err := validateString(field.name, *field.value, false); err != nil {
				return err
			}
		}
	}
	if err := descriptor.DataType.Validate(); err != nil {
		return fmt.Errorf("column.data_type: %w", err)
	}
	if err := descriptor.TypeAuthority.Validate(); err != nil {
		return fmt.Errorf("column.type_authority: %w", err)
	}
	return nil
}

func validateString(name, value string, required bool) error {
	if required && value == "" {
		return fmt.Errorf("%s is required", name)
	}
	if !utf8.ValidString(value) || !norm.NFC.IsNormalString(value) {
		return fmt.Errorf("%s must be UTF-8 NFC", name)
	}
	return nil
}
