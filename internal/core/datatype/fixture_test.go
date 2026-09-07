package datatype

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"golang.org/x/text/unicode/norm"
)

type canonicalFixture struct {
	SchemaVersion string                  `json:"schema_version"`
	Values        []canonicalFixtureValue `json:"values"`
}

type canonicalFixtureValue struct {
	Name                  string          `json:"name"`
	TypeID                ID              `json:"type_id"`
	Input                 json.RawMessage `json:"input"`
	ExpectedEncoded       string          `json:"expected_encoded"`
	NormalizationInput    string          `json:"normalization_input"`
	NormalizationExpected string          `json:"normalization_expected"`
}

func TestCanonicalDataFixturePreservesRequiredValues(t *testing.T) {
	fixtureBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "contracts", "canonical-data", "primitive-values.json"))
	if err != nil {
		t.Fatal(err)
	}

	var fixture canonicalFixture
	if err := json.Unmarshal(fixtureBytes, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SchemaVersion != "canonical-data-fixtures-v1" {
		t.Fatalf("schema_version = %q", fixture.SchemaVersion)
	}

	registry, err := NewBuiltinRegistry()
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool, len(fixture.Values))
	for _, test := range fixture.Values {
		t.Run(test.Name, func(t *testing.T) {
			seen[test.Name] = true
			input, err := fixtureInput(test)
			if err != nil {
				t.Fatal(err)
			}
			value, err := registry.Parse(test.TypeID, input)
			if err != nil {
				t.Fatalf("parse %s: %v", test.TypeID, err)
			}
			if got := string(value.Encoded()); got != test.ExpectedEncoded {
				t.Fatalf("encoded = %q, want %q", got, test.ExpectedEncoded)
			}
			if test.NormalizationExpected != "" {
				if norm.NFC.IsNormalString(test.NormalizationInput) {
					t.Fatalf("normalization_input is unexpectedly already NFC: %q", test.NormalizationInput)
				}
				if got := norm.NFC.String(test.NormalizationInput); got != test.NormalizationExpected {
					t.Fatalf("NFC normalization = %q, want %q", got, test.NormalizationExpected)
				}
				if !norm.NFC.IsNormalString(test.NormalizationExpected) {
					t.Fatalf("normalization_expected is not NFC: %q", test.NormalizationExpected)
				}
			}
		})
	}

	for _, name := range []string{"hebrew", "emoji", "nfc-normalization", "multiline", "null", "empty", "hyphen", "boolean-false", "integer-zero", "exact-integer", "exact-decimal"} {
		if !seen[name] {
			t.Errorf("fixture is missing %q", name)
		}
	}
}

func fixtureInput(value canonicalFixtureValue) (string, error) {
	switch value.TypeID {
	case NullID:
		if string(value.Input) != "null" {
			return "", strconv.ErrSyntax
		}
		return "null", nil
	case BooleanID:
		var boolean bool
		if err := json.Unmarshal(value.Input, &boolean); err != nil {
			return "", err
		}
		return strconv.FormatBool(boolean), nil
	default:
		var text string
		if err := json.Unmarshal(value.Input, &text); err != nil {
			return "", err
		}
		return text, nil
	}
}
