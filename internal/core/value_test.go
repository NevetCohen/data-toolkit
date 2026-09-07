package core

import "testing"

func TestValueCopiesConstructorAndAccessorBytes(t *testing.T) {
	encoded := []byte("canonical")
	value, err := NewValue(DataTypeID("text"), encoded)
	if err != nil {
		t.Fatalf("NewValue() error = %v", err)
	}

	encoded[0] = 'X'
	if got := string(value.Encoded()); got != "canonical" {
		t.Fatalf("constructor bytes mutated value: got %q", got)
	}

	accessed := value.Encoded()
	accessed[0] = 'Y'
	if got := string(value.Encoded()); got != "canonical" {
		t.Fatalf("accessor bytes mutated value: got %q", got)
	}
}

func TestValueAllowsEmptyBytesOnlyForString(t *testing.T) {
	value, err := NewValue(DataTypeID("string"), nil)
	if err != nil {
		t.Fatalf("NewValue(string, nil) error = %v", err)
	}
	if got := len(value.Encoded()); got != 0 {
		t.Fatalf("encoded length = %d, want 0", got)
	}

	if _, err := NewValue(DataTypeID("null"), nil); err == nil {
		t.Fatal("NewValue(null, nil) succeeded")
	}
}
