package datatype

import "testing"

func TestBuiltinRegistryUsesOneTemplateForCanonicalTypes(t *testing.T) {
	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	want := []ID{NullID, StringID, IntegerID, DecimalID, BooleanID, DateID, TimeID}
	if got := len(registry.Descriptors()); got != len(want) {
		t.Fatalf("descriptor count = %d, want %d", got, len(want))
	}
	for _, identity := range want {
		handler, err := registry.Require(identity)
		if err != nil {
			t.Errorf("required built-in %q: %v", identity, err)
			continue
		}
		if descriptor := handler.Descriptor(); descriptor.ID != identity || descriptor.Version != CurrentVersion {
			t.Errorf("descriptor for %q = %#v, want ID %q at %q", identity, descriptor, identity, CurrentVersion)
		}
	}
	if _, err := registry.Require(ID("json_object")); err == nil {
		t.Fatal("json_object was registered outside the seven canonical V1 types")
	}

	integer, err := registry.Parse(IntegerID, "900719925474099312345")
	if err != nil {
		t.Fatal(err)
	}
	smaller, err := registry.Parse(IntegerID, "900719925474099312344")
	if err != nil {
		t.Fatal(err)
	}
	comparison, err := registry.Compare(integer, smaller)
	if err != nil || comparison <= 0 {
		t.Fatalf("exact integer comparison = %d, %v", comparison, err)
	}
	if _, err := registry.Parse(IntegerID, "1.0"); err == nil {
		t.Fatal("integer parser accepted a decimal")
	}
}

func TestCanonicalNullEmptyStringAndHyphenAreDistinct(t *testing.T) {
	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}

	nullValue, err := registry.Parse(NullID, "null")
	if err != nil {
		t.Fatalf("parse canonical null: %v", err)
	}
	emptyValue, err := registry.Parse(StringID, "")
	if err != nil {
		t.Fatalf("parse empty string: %v", err)
	}
	hyphenValue, err := registry.Parse(StringID, "-")
	if err != nil {
		t.Fatalf("parse literal hyphen: %v", err)
	}

	if nullValue.TypeID() != NullID || string(nullValue.Encoded()) != "null" {
		t.Fatalf("canonical null = (%q, %q)", nullValue.TypeID(), nullValue.Encoded())
	}
	if emptyValue.TypeID() != StringID || len(emptyValue.Encoded()) != 0 {
		t.Fatalf("empty string = (%q, %q)", emptyValue.TypeID(), emptyValue.Encoded())
	}
	if hyphenValue.TypeID() != StringID || string(hyphenValue.Encoded()) != `"-"` {
		t.Fatalf("literal hyphen = (%q, %q)", hyphenValue.TypeID(), hyphenValue.Encoded())
	}
	if nullValue.Equal(emptyValue) || nullValue.Equal(hyphenValue) || emptyValue.Equal(hyphenValue) {
		t.Fatal("canonical null, empty string, and literal hyphen must remain distinct")
	}
}

func TestBuiltinNumericValuesRemainExactWithoutFloat64Conversion(t *testing.T) {
	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		id   ID
		text string
	}{
		{
			name: "integer beyond float64 precision",
			id:   IntegerID,
			text: "900719925474099312345678901234567890123456789",
		},
		{
			name: "decimal beyond float64 precision",
			id:   DecimalID,
			text: "9007199254740993.123456789012345678901234567890123456789",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := registry.Parse(test.id, test.text)
			if err != nil {
				t.Fatalf("parse %s: %v", test.id, err)
			}
			if got := string(value.Encoded()); got != test.text {
				t.Fatalf("encoded %s = %q, want exact input %q", test.id, got, test.text)
			}
		})
	}

	integer, err := registry.Parse(IntegerID, "900719925474099312345678901234567890123456789")
	if err != nil {
		t.Fatal(err)
	}
	nextInteger, err := registry.Parse(IntegerID, "900719925474099312345678901234567890123456790")
	if err != nil {
		t.Fatal(err)
	}
	if comparison, err := registry.Compare(integer, nextInteger); err != nil || comparison >= 0 {
		t.Fatalf("exact integer comparison = %d, %v; adjacent values must remain ordered", comparison, err)
	}

	decimal, err := registry.Parse(DecimalID, "1.000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	nextDecimal, err := registry.Parse(DecimalID, "1.000000000000000000000000000000000002")
	if err != nil {
		t.Fatal(err)
	}
	if comparison, err := registry.Compare(decimal, nextDecimal); err != nil || comparison >= 0 {
		t.Fatalf("exact decimal comparison = %d, %v; adjacent values must remain ordered", comparison, err)
	}
}
