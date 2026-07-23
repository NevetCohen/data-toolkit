package cli

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"data-toolkit/internal/contract"
)

func TestWorkflowValidateSupportsYAMLAndJSONWithoutReadingSources(t *testing.T) {
	fixtureRoot := mvp2FixtureRoot(t)
	yamlPayload, err := os.ReadFile(filepath.Join(fixtureRoot, "workflow.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	temporaryRoot := t.TempDir()
	yamlPath := filepath.Join(temporaryRoot, "workflow.yaml")
	if err := os.WriteFile(yamlPath, yamlPayload, 0o600); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{yamlPath, writeJSONWorkflow(t, yamlPath)} {
		t.Run(filepath.Ext(path), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Execute(context.Background(), []string{"workflow", "validate", path}, &stdout, &stderr)
			if code != ExitSuccess {
				t.Fatalf("exit = %d, stderr = %s", code, stderr.String())
			}
			var result struct {
				Status string `json:"status"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("decode stdout: %v\n%s", err, stdout.String())
			}
			if result.Status != "valid" || stderr.Len() != 0 {
				t.Fatalf("stdout = %s, stderr = %s", stdout.String(), stderr.String())
			}
		})
	}
}

func TestWorkflowValidateRejectsInvalidWorkflow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"v1"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Execute(context.Background(), []string{"workflow", "validate", path}, &stdout, &stderr)
	if code != ExitFailure || stdout.Len() != 0 || !strings.Contains(stderr.String(), "error:") {
		t.Fatalf("exit = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
}

func TestWorkflowRunRejectsInvalidConfigBeforeSourceIO(t *testing.T) {
	workflowPath := filepath.Join(mvp2FixtureRoot(t), "workflow.yaml")
	configPath := filepath.Join(t.TempDir(), "invalid-config.json")
	payload := []byte(`{"schema_version":"v1","output":{"collision":"explode"}}`)
	if err := os.WriteFile(configPath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Execute(
		context.Background(),
		[]string{"workflow", "run", workflowPath, "--config", configPath},
		&stdout,
		&stderr,
	)
	if code != ExitFailure || stdout.Len() != 0 ||
		!strings.Contains(stderr.String(), "decode configuration") {
		t.Fatalf("exit = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
}

func TestWorkflowRunResolvesRelativePathsAndMatchesMVP2(t *testing.T) {
	fixtureRoot := mvp2FixtureRoot(t)
	runRoot := t.TempDir()
	for _, name := range []string{"input.json", "workflow.yaml", "expected.csv"} {
		payload, err := os.ReadFile(filepath.Join(fixtureRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(runRoot, name), payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var stdout, stderr bytes.Buffer
	code := Execute(
		context.Background(),
		[]string{"workflow", "run", filepath.Join(runRoot, "workflow.yaml")},
		&stdout,
		&stderr,
	)
	if code != ExitSuccess {
		t.Fatalf("exit = %d, stdout = %s, stderr = %s", code, stdout.String(), stderr.String())
	}
	var final contract.FinalResult
	if err := json.Unmarshal(stdout.Bytes(), &final); err != nil {
		t.Fatalf("decode final result: %v\n%s", err, stdout.String())
	}
	if final.Status != contract.RunStatusSucceeded || len(final.Outputs) != 1 {
		t.Fatalf("final = %#v", final)
	}
	wantLocation := filepath.Join(runRoot, "generated", "mvp2.csv")
	if final.Outputs[0].Location != wantLocation {
		t.Fatalf("output location = %q, want %q", final.Outputs[0].Location, wantLocation)
	}
	if !reflect.DeepEqual(readCSV(t, wantLocation), readCSV(t, filepath.Join(runRoot, "expected.csv"))) {
		t.Fatal("executable MVP 2 output differs from expected.csv")
	}
	var started contract.StartedResult
	if err := json.Unmarshal(stderr.Bytes(), &started); err != nil {
		t.Fatalf("decode started status: %v\n%s", err, stderr.String())
	}
	if started.Status != contract.RunStatusStarted {
		t.Fatalf("started = %#v", started)
	}
}

func TestWorkflowUsageErrorsExitTwo(t *testing.T) {
	for _, args := range [][]string{
		{"workflow", "run"},
		{"workflow", "validate"},
		{"workflow", "unknown"},
		{"workflow", "run", "one.yaml", "two.yaml"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Execute(context.Background(), args, &stdout, &stderr); code != ExitUsage {
			t.Fatalf("args %v exit = %d, stderr = %q", args, code, stderr.String())
		}
	}
}

func writeJSONWorkflow(t *testing.T, yamlPath string) string {
	t.Helper()
	workflow, err := loadWorkflowFile(yamlPath)
	if err != nil {
		t.Fatal(err)
	}
	workflow.Sources[0].Location = "missing-input.json"
	workflow.Outputs[0].Location = filepath.Join("generated", "missing.csv")
	payload, err := json.Marshal(workflow)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(filepath.Dir(yamlPath), "workflow.json")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func mvp2FixtureRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "testdata", "fixtures", "mvp2-json-to-csv"))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func readCSV(t *testing.T, path string) [][]string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	return rows
}
