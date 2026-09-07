package column

import (
	"reflect"
	"strings"
	"testing"

	"data-toolkit/internal/core"
)

func TestColumnDescriptorHasExactV1Fields(t *testing.T) {
	typeOfDescriptor := reflect.TypeFor[ColumnDescriptor]()
	want := []string{
		"ID",
		"PhysicalPosition",
		"SourceHeader",
		"Subheader",
		"Alias",
		"SemanticType",
		"DataType",
		"TypeAuthority",
	}
	if typeOfDescriptor.NumField() != len(want) {
		t.Fatalf("ColumnDescriptor field count = %d, want %d", typeOfDescriptor.NumField(), len(want))
	}
	for index, name := range want {
		field := typeOfDescriptor.Field(index)
		if field.Name != name {
			t.Errorf("field %d = %q, want %q", index, field.Name, name)
		}
	}
}

func TestColumnDescriptorValidation(t *testing.T) {
	valid := ColumnDescriptor{
		ID:               "email",
		PhysicalPosition: "C",
		SourceHeader:     "Email",
		DataType:         core.DataTypeID("string"),
		TypeAuthority:    core.TypeAuthorityInferred,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid descriptor: %v", err)
	}

	tests := []struct {
		name string
		edit func(*ColumnDescriptor)
		want string
	}{
		{name: "missing ID", edit: func(value *ColumnDescriptor) { value.ID = "" }, want: "column.id"},
		{name: "missing physical position", edit: func(value *ColumnDescriptor) { value.PhysicalPosition = "" }, want: "column.physical_position"},
		{name: "non NFC source header", edit: func(value *ColumnDescriptor) { value.SourceHeader = "e\u0301" }, want: "column.source_header"},
		{name: "non NFC optional alias", edit: func(value *ColumnDescriptor) { alias := "e\u0301"; value.Alias = &alias }, want: "column.alias"},
		{name: "invalid data type", edit: func(value *ColumnDescriptor) { value.DataType = "not valid" }, want: "column.data_type"},
		{name: "invalid type authority", edit: func(value *ColumnDescriptor) { value.TypeAuthority = "unknown" }, want: "column.type_authority"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := valid
			test.edit(&value)
			if err := value.Validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestColumnDescriptorOptionalFields(t *testing.T) {
	value := ColumnDescriptor{
		ID:               "email",
		PhysicalPosition: "C",
		SourceHeader:     "",
		DataType:         core.DataTypeID("string"),
		TypeAuthority:    core.TypeAuthorityDeclared,
	}
	if err := value.Validate(); err != nil {
		t.Fatalf("descriptor without optional fields: %v", err)
	}
	if value.Subheader != nil || value.Alias != nil || value.SemanticType != nil {
		t.Fatal("optional fields must remain omitted when unset")
	}
}
