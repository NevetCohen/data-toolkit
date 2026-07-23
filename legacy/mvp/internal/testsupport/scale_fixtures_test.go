package testsupport

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
)

func TestScaleFixtureGeneratorsAreDeterministic(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		const targetBytes = int64(64 * 1024)
		first, err := GenerateJSONScaleFixture(filepath.Join(t.TempDir(), "first.json"), targetBytes)
		if err != nil {
			t.Fatal(err)
		}
		second, err := GenerateJSONScaleFixture(filepath.Join(t.TempDir(), "second.json"), targetBytes)
		if err != nil {
			t.Fatal(err)
		}
		if first.Size < targetBytes || first.Size-targetBytes > 32*1024 || first.Rows == 0 || first.Rows != second.Rows {
			t.Fatalf("JSON fixtures = %#v and %#v", first, second)
		}
		firstSnapshot, err := SnapshotFile(first.Path)
		if err != nil {
			t.Fatal(err)
		}
		secondSnapshot, err := SnapshotFile(second.Path)
		if err != nil {
			t.Fatal(err)
		}
		if firstSnapshot.SHA256 != secondSnapshot.SHA256 {
			t.Fatal("same JSON scale parameters produced different bytes")
		}
		content, err := os.ReadFile(first.Path)
		if err != nil {
			t.Fatal(err)
		}
		if !containsBytes(content, []byte("רשומה-000001 🧭")) || !containsBytes(content, []byte("בת ים")) {
			t.Fatal("JSON scale fixture does not contain the Unicode sentinels")
		}
	})

	t.Run("CSV", func(t *testing.T) {
		const rows = 25
		first, err := GenerateCSVScaleFixture(filepath.Join(t.TempDir(), "first.csv"), rows)
		if err != nil {
			t.Fatal(err)
		}
		second, err := GenerateCSVScaleFixture(filepath.Join(t.TempDir(), "second.csv"), rows)
		if err != nil {
			t.Fatal(err)
		}
		firstSnapshot, err := SnapshotFile(first.Path)
		if err != nil {
			t.Fatal(err)
		}
		secondSnapshot, err := SnapshotFile(second.Path)
		if err != nil {
			t.Fatal(err)
		}
		if firstSnapshot.SHA256 != secondSnapshot.SHA256 {
			t.Fatal("same CSV scale parameters produced different bytes")
		}
		file, err := os.Open(first.Path)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		records, err := csv.NewReader(file).ReadAll()
		if err != nil {
			t.Fatal(err)
		}
		if len(records) != rows+1 || records[1][0] != "1" || records[1][1] != "רשומה-000001 🧭" || records[rows][0] != "25" {
			t.Fatalf("CSV scale fixture records = %#v", records)
		}
	})
}

func containsBytes(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return true
	}
	for offset := 0; offset+len(needle) <= len(haystack); offset++ {
		match := true
		for index := range needle {
			if haystack[offset+index] != needle[index] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
