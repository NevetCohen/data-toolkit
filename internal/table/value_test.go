package table

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCanonicalValueKindsRemainDistinct(t *testing.T) {
	integer, err := IntegerValue("9007199254740993")
	if err != nil {
		t.Fatal(err)
	}
	decimal, err := DecimalValue("0.1000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	date, err := DateValue(2026, time.February, 2)
	if err != nil {
		t.Fatal(err)
	}
	localTime, err := TimeValue(14, 34, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	object, err := JSONValue([]byte("{\"nested\":{\"emoji\":\"🧭\"}}"))
	if err != nil {
		t.Fatal(err)
	}
	array, err := JSONValue([]byte("[1,\"עברית\"]"))
	if err != nil {
		t.Fatal(err)
	}

	values := []Value{
		NullValue(),
		StringValue("עברית 🧭"),
		BooleanValue(true),
		integer,
		decimal,
		date,
		localTime,
		object,
		array,
	}
	wantKinds := []ValueKind{
		KindNull, KindString, KindBoolean, KindInteger, KindDecimal,
		KindDate, KindTime, KindJSONObject, KindJSONArray,
	}
	for index := range values {
		if values[index].Kind() != wantKinds[index] {
			t.Fatalf("kind[%d] = %q, want %q", index, values[index].Kind(), wantKinds[index])
		}
		encoded, err := json.Marshal(values[index])
		if err != nil {
			t.Fatal(err)
		}
		var decoded Value
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.Kind() != values[index].Kind() {
			t.Fatalf("round-trip kind[%d] = %q", index, decoded.Kind())
		}
	}
}

func TestExactNumericComparisonAndFormatting(t *testing.T) {
	larger, _ := IntegerValue("9007199254740993")
	smaller, _ := IntegerValue("9007199254740992")
	if comparison, err := CompareNumeric(larger, smaller); err != nil || comparison <= 0 {
		t.Fatalf("exact integer comparison = %d, %v", comparison, err)
	}

	precise, _ := DecimalValue("0.1000000000000000000000000001")
	tenth, _ := DecimalValue("0.1")
	if comparison, err := CompareNumeric(precise, tenth); err != nil || comparison <= 0 {
		t.Fatalf("exact decimal comparison = %d, %v", comparison, err)
	}
	if raw, ok := precise.RawNumericLexeme(); !ok || raw != "0.1000000000000000000000000001" {
		t.Fatalf("raw numeric lexeme = %q, %v", raw, ok)
	}

	value, _ := DecimalValue("123.456789")
	formatted, err := FormatNumeric(value, 3)
	if err != nil {
		t.Fatal(err)
	}
	if formatted != "123.457" {
		t.Fatalf("formatted decimal = %q", formatted)
	}
}
