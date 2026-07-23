package datatype

import "testing"

func TestBuiltinRegistryUsesOneTemplateForCanonicalTypes(t *testing.T) {
	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	want := []ID{NullID, StringID, BooleanID, IntegerID, DecimalID, DateID, TimeID, JSONObjectID, JSONArrayID}
	if got := len(registry.Descriptors()); got != len(want) {
		t.Fatalf("descriptor count = %d, want %d", got, len(want))
	}
	for _, identity := range want {
		if _, err := registry.Require(identity); err != nil {
			t.Errorf("required built-in %q: %v", identity, err)
		}
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
	if _, err := registry.Parse(JSONObjectID, `["not","object"]`); err == nil {
		t.Fatal("object parser accepted an array")
	}
}
