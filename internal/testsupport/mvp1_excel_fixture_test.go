package testsupport

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xuri/excelize/v2"
)

type mvp1FixtureAssertions struct {
	Workbook          string     `json:"workbook"`
	Sheet             string     `json:"sheet"`
	SourceSHA256      string     `json:"source_sha256"`
	FixtureSHA256     string     `json:"fixture_sha256"`
	SourceBytes       int64      `json:"source_bytes"`
	Headers           []string   `json:"headers"`
	CityHeader        string     `json:"city_header"`
	CityColumn        int        `json:"city_column"`
	FixtureRows       int        `json:"fixture_rows"`
	ExpectedMatchRows []int      `json:"expected_match_rows"`
	ExpectedValues    [][]string `json:"expected_values"`
}

const approvedMVP1Source = `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\all_final\all_final.xlsx`

func TestMVP1ExcelFixtureReadbackAndHash(t *testing.T) {
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	fixtureRoot := filepath.Join(repositoryRoot, "testdata", "fixtures", "mvp1-excel-city-filter")
	assertions := readMVP1FixtureAssertions(t, filepath.Join(fixtureRoot, "expected-assertions.json"))
	fixturePath := filepath.Join(fixtureRoot, assertions.Workbook)
	before, err := SnapshotFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	defer AssertFileUnchanged(t, before)
	if got := hex.EncodeToString(before.SHA256[:]); got != lower(assertions.FixtureSHA256) {
		t.Fatalf("fixture SHA-256 = %s, want %s", got, assertions.FixtureSHA256)
	}
	if assertions.SourceSHA256 != "F7D3861B458E548127ED8CCEA4C69B70CCDB5EEC0B464BC9F70F6D9B8E5C399C" || assertions.SourceBytes != 7065978 {
		t.Fatalf("source provenance = %#v", assertions)
	}
	sourceBefore, err := SnapshotFile(approvedMVP1Source)
	if err != nil {
		t.Fatal(err)
	}
	defer AssertFileUnchanged(t, sourceBefore)
	if got := hex.EncodeToString(sourceBefore.SHA256[:]); got != lower(assertions.SourceSHA256) || sourceBefore.Size != assertions.SourceBytes {
		t.Fatalf("approved source identity = SHA-256 %s, size %d; want SHA-256 %s, size %d", got, sourceBefore.Size, assertions.SourceSHA256, assertions.SourceBytes)
	}

	workbook, err := excelize.OpenFile(fixturePath, excelize.Options{RawCellValue: true})
	if err != nil {
		t.Fatal(err)
	}
	defer workbook.Close()
	if sheets := workbook.GetSheetList(); !reflect.DeepEqual(sheets, []string{assertions.Sheet}) {
		t.Fatalf("sheets = %#v, want [%q]", sheets, assertions.Sheet)
	}
	rows, err := workbook.GetRows(assertions.Sheet, excelize.Options{RawCellValue: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != assertions.FixtureRows+1 || !reflect.DeepEqual(rows[0], assertions.Headers) || !reflect.DeepEqual(rows[1:], assertions.ExpectedValues) {
		t.Fatalf("readback rows = %#v", rows)
	}
	if assertions.CityColumn != 6 || assertions.CityHeader != "ישוב" || !reflect.DeepEqual(assertions.ExpectedMatchRows, []int{3, 4, 6}) {
		t.Fatalf("filter contract = %#v", assertions)
	}
	for _, rowNumber := range assertions.ExpectedMatchRows {
		if rows[rowNumber-1][assertions.CityColumn-1] != "בת ים" {
			t.Fatalf("row %d city = %q", rowNumber, rows[rowNumber-1][assertions.CityColumn-1])
		}
	}
}

func readMVP1FixtureAssertions(t *testing.T, path string) mvp1FixtureAssertions {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var assertions mvp1FixtureAssertions
	if err := json.Unmarshal(content, &assertions); err != nil {
		t.Fatal(err)
	}
	return assertions
}

func lower(value string) string {
	result := make([]byte, len(value))
	for index := range value {
		if value[index] >= 'A' && value[index] <= 'Z' {
			result[index] = value[index] + ('a' - 'A')
		} else {
			result[index] = value[index]
		}
	}
	return string(result)
}
