package testsupport

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const approvedMVP3Script = `C:\Users\nevet\האחסון שלי\5 - פעילות פוליטית\מפקד הדמוקרטים 2026\נתוני מתפקדים\scripts\Invoke-RawMemberCleanup.Streaming.py`
const mvp3BehaviorDoc = `..\..\docs\mvp3-cleanup-script-behavior.md`

type mvp3FixtureCase struct {
	Filter   []string          `json:"filter"`
	Counters map[string]string `json:"counters"`
	Headers  []string          `json:"headers"`
	Rows     [][]string        `json:"rows"`
	IDLines  []string          `json:"id_lines"`
}

type mvp3FixtureAssertions struct {
	Normal       mvp3FixtureCase   `json:"normal"`
	ZeroAccepted mvp3FixtureCase   `json:"zero_accepted"`
	MaxIndex     mvp3FixtureCase   `json:"max_index"`
	Errors       map[string]string `json:"errors"`
}

type mvp3RunOptions struct {
	Input, CSVSeed, IDsSeed string
	Filter                  []string
	DryRun, IDsOnly         bool
	TrimCSVFinalNewline     bool
}

type mvp3ReferenceRun struct {
	Summary                        map[string]string
	CSV, IDs, BeforeCSV, BeforeIDs []byte
	CSVExists                      bool
}

func TestMVP3CleanupReferenceFixtures(t *testing.T) {
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	fixtureRoot := filepath.Join(repositoryRoot, "testdata", "fixtures", "mvp3-cleanup-script-replacement")
	for _, snapshot := range snapshotMVP3FixtureTree(t, fixtureRoot) {
		defer AssertFileUnchanged(t, snapshot)
	}

	scriptBefore, err := SnapshotFile(approvedMVP3Script)
	if err != nil {
		t.Fatal(err)
	}
	defer AssertFileUnchanged(t, scriptBefore)
	behaviorBefore, err := SnapshotFile(mvp3BehaviorDoc)
	if err != nil {
		t.Fatal(err)
	}
	defer AssertFileUnchanged(t, behaviorBefore)
	if got := hex.EncodeToString(scriptBefore.SHA256[:]); got != "2d13534e0c58df08db9ae9cff94c08b2701d21fd48887180762b2343e7636dfe" {
		t.Fatalf("reference script SHA-256 = %s", got)
	}
	for _, name := range []string{"normal-input.json", "normal-seed-entity-ids.txt"} {
		content, err := os.ReadFile(filepath.Join(fixtureRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.HasPrefix(content, []byte{0xef, 0xbb, 0xbf}) {
			t.Fatalf("%s must start with a UTF-8 BOM", name)
		}
	}

	var assertions mvp3FixtureAssertions
	content, err := os.ReadFile(filepath.Join(fixtureRoot, "assertions.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(content, &assertions); err != nil {
		t.Fatal(err)
	}

	t.Run("normal", func(t *testing.T) {
		run := runMVP3Reference(t, fixtureRoot, mvp3RunOptions{
			Input: "normal-input.json", CSVSeed: "normal-seed.csv", IDsSeed: "normal-seed-entity-ids.txt",
			Filter: assertions.Normal.Filter, TrimCSVFinalNewline: true,
		})
		assertMVP3FixtureCase(t, run, assertions.Normal)
		if !bytes.Contains(run.CSV, []byte("\r\n3,00123")) {
			t.Fatal("no-final-newline CSV repair was not preserved")
		}
	})

	t.Run("ids-only", func(t *testing.T) {
		run := runMVP3Reference(t, fixtureRoot, mvp3RunOptions{
			Input: "normal-input.json", IDsSeed: "normal-seed-entity-ids.txt", Filter: assertions.Normal.Filter, IDsOnly: true,
		})
		if run.CSVExists {
			t.Fatal("ids-only run created a CSV output")
		}
		wantIDs := append([]string(nil), assertions.Normal.IDLines...)
		for index := range wantIDs {
			wantIDs[index] = strings.TrimSuffix(wantIDs[index], ",")
		}
		if got := mvp3Lines(run.IDs); !reflect.DeepEqual(got, wantIDs) {
			t.Fatalf("ID lines = %#v, want %#v", got, wantIDs)
		}
		if run.Summary["IdsOnly"] != "True" || run.Summary["CsvLastIndexBeforeRun"] != "0" {
			t.Fatalf("ids-only counters = %#v", run.Summary)
		}
	})

	t.Run("dry-run", func(t *testing.T) {
		run := runMVP3Reference(t, fixtureRoot, mvp3RunOptions{
			Input: "normal-input.json", CSVSeed: "normal-seed.csv", IDsSeed: "normal-seed-entity-ids.txt",
			Filter: assertions.Normal.Filter, DryRun: true, TrimCSVFinalNewline: true,
		})
		assertMVP3Counters(t, run.Summary, assertions.Normal.Counters)
		if !bytes.Equal(run.CSV, run.BeforeCSV) || !bytes.Equal(run.IDs, run.BeforeIDs) {
			t.Fatal("dry-run changed disposable outputs")
		}
	})

	t.Run("zero-accepted-default-header", func(t *testing.T) {
		run := runMVP3Reference(t, fixtureRoot, mvp3RunOptions{Input: "zero-accepted-input.json", Filter: assertions.ZeroAccepted.Filter})
		assertMVP3FixtureCase(t, run, assertions.ZeroAccepted)
	})

	t.Run("maximum-index", func(t *testing.T) {
		run := runMVP3Reference(t, fixtureRoot, mvp3RunOptions{Input: "max-index-input.json", CSVSeed: "max-index-seed.csv"})
		assertMVP3FixtureCase(t, run, assertions.MaxIndex)
	})

	for input, expectedError := range assertions.Errors {
		t.Run(input, func(t *testing.T) {
			if output := runMVP3ReferenceError(t, fixtureRoot, input); !strings.Contains(output, expectedError) {
				t.Fatalf("error output = %q, want substring %q", output, expectedError)
			}
		})
	}
}

func runMVP3Reference(t *testing.T, fixtureRoot string, options mvp3RunOptions) mvp3ReferenceRun {
	t.Helper()
	directory := t.TempDir()
	inputPath := copyMVP3FixtureFile(t, fixtureRoot, options.Input, directory, "input.json")
	idsPath := filepath.Join(directory, "entity_ids.txt")
	if options.IDsSeed != "" {
		copyMVP3FixtureFile(t, fixtureRoot, options.IDsSeed, directory, "entity_ids.txt")
	}
	csvPath := filepath.Join(directory, "out.csv")
	if options.CSVSeed != "" {
		copyMVP3FixtureFile(t, fixtureRoot, options.CSVSeed, directory, "out.csv")
		if options.TrimCSVFinalNewline {
			content, err := os.ReadFile(csvPath)
			if err != nil {
				t.Fatal(err)
			}
			content = bytes.TrimSuffix(content, []byte("\r\n"))
			content = bytes.TrimSuffix(content, []byte("\n"))
			if err := os.WriteFile(csvPath, content, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	beforeCSV, _ := os.ReadFile(csvPath)
	beforeIDs, _ := os.ReadFile(idsPath)

	arguments := []string{"-3", approvedMVP3Script, "--json-path", inputPath, "--entity-ids-path", idsPath, "--csv-path", csvPath}
	if options.DryRun {
		arguments = append(arguments, "--dry-run")
	}
	if options.IDsOnly {
		arguments = append(arguments, "--ids-only")
	}
	if len(options.Filter) == 2 {
		arguments = append(arguments, "--filter-field", options.Filter[0], "--filter-value", options.Filter[1])
	}
	output, err := exec.Command("py", arguments...).CombinedOutput()
	if err != nil {
		t.Fatalf("reference script failed: %v\n%s", err, output)
	}
	csvContent, csvErr := os.ReadFile(csvPath)
	idsContent, idsErr := os.ReadFile(idsPath)
	if idsErr != nil {
		t.Fatal(idsErr)
	}
	return mvp3ReferenceRun{Summary: parseMVP3Summary(t, string(output)), CSV: csvContent, IDs: idsContent, BeforeCSV: beforeCSV, BeforeIDs: beforeIDs, CSVExists: csvErr == nil}
}

func runMVP3ReferenceError(t *testing.T, fixtureRoot, inputName string) string {
	t.Helper()
	directory := t.TempDir()
	inputPath := copyMVP3FixtureFile(t, fixtureRoot, inputName, directory, "input.json")
	idsPath := filepath.Join(directory, "entity_ids.txt")
	output, err := exec.Command("py", "-3", approvedMVP3Script, "--json-path", inputPath, "--entity-ids-path", idsPath, "--csv-path", filepath.Join(directory, "out.csv"), "--dry-run").CombinedOutput()
	if err == nil {
		t.Fatalf("%s unexpectedly succeeded", inputName)
	}
	if _, err := os.Stat(idsPath); !os.IsNotExist(err) {
		t.Fatal("failing run wrote an ID output")
	}
	return string(output)
}

func copyMVP3FixtureFile(t *testing.T, root, source, destinationDirectory, destination string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, source))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(destinationDirectory, destination)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func parseMVP3Summary(t *testing.T, output string) map[string]string {
	t.Helper()
	result := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if key, value, found := strings.Cut(line, ": "); found {
			result[key] = value
		}
	}
	return result
}

func assertMVP3FixtureCase(t *testing.T, run mvp3ReferenceRun, expected mvp3FixtureCase) {
	t.Helper()
	assertMVP3Counters(t, run.Summary, expected.Counters)
	if filepath.Base(run.Summary["CsvPath"]) != "out.csv" || filepath.Base(run.Summary["EntityIdsPath"]) != "entity_ids.txt" {
		t.Fatalf("dynamic paths = %#v", run.Summary)
	}
	rows, err := csv.NewReader(bytes.NewReader(run.CSV)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 || !reflect.DeepEqual(rows[0], expected.Headers) || !reflect.DeepEqual(rows[1:], expected.Rows) {
		t.Fatalf("CSV rows = %#v", rows)
	}
	if got := mvp3Lines(run.IDs); !reflect.DeepEqual(got, expected.IDLines) {
		t.Fatalf("ID lines = %#v, want %#v", got, expected.IDLines)
	}
}

func assertMVP3Counters(t *testing.T, actual, expected map[string]string) {
	t.Helper()
	for key, want := range expected {
		if got := actual[key]; got != want {
			t.Fatalf("%s = %q, want %q; summary = %#v", key, got, want, actual)
		}
	}
}

func mvp3Lines(content []byte) []string {
	text := strings.TrimSuffix(string(content), "\n")
	if text == "" {
		return []string{}
	}
	return strings.Split(text, "\n")
}

func snapshotMVP3FixtureTree(t *testing.T, root string) []FileSnapshot {
	t.Helper()
	var snapshots []FileSnapshot
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		snapshot, err := SnapshotFile(path)
		if err != nil {
			return err
		}
		snapshots = append(snapshots, snapshot)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snapshots
}
