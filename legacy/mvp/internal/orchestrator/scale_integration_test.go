//go:build scale

package orchestrator

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"testing"
	"time"

	"data-toolkit/internal/adapter"
	"data-toolkit/internal/config"
	"data-toolkit/internal/contract"
	"data-toolkit/internal/testsupport"
)

const (
	scaleJSONTargetBytes = int64(40 * 1024 * 1024)
	scaleCSVRows         = 500_000
	scaleMemoryBudget    = int64(64 * 1024 * 1024)
)

func TestScaleFixturesAreBoundedStableUnicodeAndDeterministic(t *testing.T) {
	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	generated := testsupport.NewGeneratedDir(t, repositoryRoot, "scale")

	t.Run("JSON_near_40_MiB", func(t *testing.T) {
		fixture, err := testsupport.GenerateJSONScaleFixture(filepath.Join(generated, "scale.json"), scaleJSONTargetBytes)
		if err != nil {
			t.Fatal(err)
		}
		if fixture.Size < scaleJSONTargetBytes || fixture.Size-scaleJSONTargetBytes > 32*1024 {
			t.Fatalf("JSON fixture size = %d, want [%d, %d]", fixture.Size, scaleJSONTargetBytes, scaleJSONTargetBytes+32*1024)
		}
		runScaleCase(t, fixture, contract.FormatJSON, generated)
	})

	t.Run("CSV_500000_rows", func(t *testing.T) {
		fixture, err := testsupport.GenerateCSVScaleFixture(filepath.Join(generated, "scale.csv"), scaleCSVRows)
		if err != nil {
			t.Fatal(err)
		}
		if fixture.Rows != scaleCSVRows {
			t.Fatalf("CSV fixture rows = %d, want %d", fixture.Rows, scaleCSVRows)
		}
		runScaleCase(t, fixture, contract.FormatCSV, generated)
	})
}

func runScaleCase(t *testing.T, fixture testsupport.ScaleFixture, format contract.DataFormat, generatedRoot string) {
	t.Helper()
	sourceBefore, err := testsupport.SnapshotFile(fixture.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer testsupport.AssertFileUnchanged(t, sourceBefore)

	outputPaths := [2]string{
		filepath.Join(generatedRoot, string(format)+"-first.csv"),
		filepath.Join(generatedRoot, string(format)+"-second.csv"),
	}
	results := [2]contract.FinalResult{}
	run := func(index int) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		workflow := scaleWorkflow(format, fixture.Path, outputPaths[index], fixture.Rows)
		result, err := runScaleWorkflow(ctx, workflow, filepath.Join(generatedRoot, fmt.Sprintf("workspace-%s-%d", format, index)), index)
		if err != nil {
			return fmt.Errorf("run %s scale workflow %d: %w", format, index+1, err)
		}
		if result.Status != contract.RunStatusSucceeded || len(result.Outputs) != 1 || result.Outputs[0].Rows != int64(fixture.Rows) {
			return fmt.Errorf("%s scale result %d = %#v", format, index+1, result)
		}
		if err := assertReportedMemoryBudget(result, scaleMemoryBudget); err != nil {
			return err
		}
		results[index] = result
		return nil
	}
	measurement, err := measurePeakHeap(scaleMemoryBudget, func() error {
		return run(0)
	})
	if err != nil {
		t.Fatal(err)
	}
	if measurement.PeakDelta > uint64(scaleMemoryBudget) {
		t.Fatalf("%s peak heap delta = %d bytes, budget = %d", format, measurement.PeakDelta, scaleMemoryBudget)
	}
	if err := run(1); err != nil {
		t.Fatal(err)
	}

	firstOutput, err := testsupport.SnapshotFile(results[0].Outputs[0].Location)
	if err != nil {
		t.Fatal(err)
	}
	secondOutput, err := testsupport.SnapshotFile(results[1].Outputs[0].Location)
	if err != nil {
		t.Fatal(err)
	}
	if firstOutput.SHA256 != secondOutput.SHA256 || firstOutput.Size != secondOutput.Size {
		t.Fatalf("%s repeated outputs differ: first=%x/%d second=%x/%d", format, firstOutput.SHA256, firstOutput.Size, secondOutput.SHA256, secondOutput.Size)
	}
	verifyScaleCSV(t, results[0].Outputs[0].Location, fixture.Rows)
	t.Logf(
		"format=%s source_bytes=%d rows=%d source_sha256=%x output_bytes=%d output_sha256=%x baseline_heap=%d peak_heap=%d peak_heap_delta=%d budget=%d soft_limit=%d",
		format, sourceBefore.Size, fixture.Rows, sourceBefore.SHA256, firstOutput.Size, firstOutput.SHA256,
		measurement.Baseline, measurement.Peak, measurement.PeakDelta, scaleMemoryBudget, measurement.SoftLimit,
	)
}

func runScaleWorkflow(ctx context.Context, workflow contract.Workflow, temporaryRoot string, run int) (contract.FinalResult, error) {
	registry := adapter.NewRegistry()
	if err := registry.Register(adapter.NewJSONAdapter()); err != nil {
		return contract.FinalResult{}, err
	}
	if err := registry.Register(adapter.NewCSVAdapter()); err != nil {
		return contract.FinalResult{}, err
	}
	defaults := config.BuiltInDefaults(temporaryRoot)
	defaults.MaximumMemoryBytes = scaleMemoryBudget
	defaults.FailedRunRetention = 0
	service := Service{
		Adapters: registry, Defaults: defaults,
		UserConfig: config.Config{SchemaVersion: config.CurrentSchemaVersion},
		RunID:      func() string { return fmt.Sprintf("scale-%s-%d", workflow.Sources[0].Format, run+1) },
		Backend:    &LocalPipelineBackend{Adapters: registry},
	}
	return service.Run(ctx, contract.RunRequest{Workflow: workflow})
}

func scaleWorkflow(format contract.DataFormat, sourcePath, outputPath string, rows int) contract.Workflow {
	source := contract.Source{ID: "scale-source", Format: format, Location: sourcePath}
	var mappings []contract.FieldMapping
	if format == contract.FormatJSON {
		source.Options.JSONRoot = "/Data"
		for _, field := range []string{"id", "name", "city", "marker"} {
			mappings = append(mappings, contract.FieldMapping{
				SourceID: source.ID, Column: contract.ColumnRef{Header: field}, JSONPath: "/" + field,
			})
		}
	}
	return contract.Workflow{
		SchemaVersion: contract.CurrentSchemaVersion,
		ID:            "scale-" + string(format),
		Sources:       []contract.Source{source},
		Mappings:      mappings,
		Outputs: []contract.Output{{
			ID: "output", Input: source.ID, Format: contract.FormatCSV, Location: outputPath,
			Projection: []contract.ProjectionField{
				{Column: contract.ColumnRef{Header: "id"}, OutputAs: "מזהה"},
				{Column: contract.ColumnRef{Header: "name"}, OutputAs: "שם"},
				{Column: contract.ColumnRef{Header: "city"}, OutputAs: "ישוב"},
				{Column: contract.ColumnRef{Header: "marker"}, OutputAs: "סמן"},
			},
			Validations: []contract.Validation{{ID: "rows", Kind: contract.ValidationRowCount, Expected: strconv.Itoa(rows)}},
		}},
		Interface: contract.InterfaceMetadata{Name: "scale-test"},
		Exception: contract.ExceptionPolicy{Action: contract.ExceptionBlock},
	}
}

func verifyScaleCSV(t *testing.T, path string, wantRows int) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.ReuseRecord = true
	header, err := reader.Read()
	if err != nil {
		t.Fatal(err)
	}
	wantHeader := []string{"מזהה", "שם", "ישוב", "סמן"}
	for index := range wantHeader {
		if header[index] != wantHeader[index] {
			t.Fatalf("header[%d] = %q, want %q", index, header[index], wantHeader[index])
		}
	}
	rows := 0
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("read output row %d: %v", rows+1, err)
		}
		rows++
		wantName := fmt.Sprintf("רשומה-%06d 🧭", rows)
		wantCity := "תל אביב"
		if rows%2 == 1 {
			wantCity = "בת ים"
		}
		if record[0] != strconv.Itoa(rows) || record[1] != wantName || record[2] != wantCity || record[3] != "עברית, English 🙂" {
			t.Fatalf("output row %d is out of order or changed Unicode: %#v", rows, record)
		}
	}
	if rows != wantRows {
		t.Fatalf("output rows = %d, want %d", rows, wantRows)
	}
}

func assertReportedMemoryBudget(result contract.FinalResult, want int64) error {
	runtimeSettings, ok := result.ResolvedSettings["runtime"].(map[string]any)
	if !ok {
		return errors.New("final result has no resolved runtime settings")
	}
	got, ok := runtimeSettings["maximum_memory_bytes"].(int64)
	if !ok || got != want {
		return fmt.Errorf("reported maximum memory = %#v, want %d", runtimeSettings["maximum_memory_bytes"], want)
	}
	return nil
}

type heapMeasurement struct {
	Baseline  uint64
	Peak      uint64
	PeakDelta uint64
	SoftLimit int64
}

func measurePeakHeap(budget int64, action func() error) (heapMeasurement, error) {
	runtime.GC()
	var baseline runtime.MemStats
	runtime.ReadMemStats(&baseline)
	runtimeFootprint := baseline.Sys - baseline.HeapReleased
	softLimit := int64(runtimeFootprint) + budget
	previousLimit := debug.SetMemoryLimit(-1)
	if previousLimit >= 0 && previousLimit < softLimit {
		softLimit = previousLimit
	}
	debug.SetMemoryLimit(softLimit)
	defer debug.SetMemoryLimit(previousLimit)

	stop := make(chan struct{})
	done := make(chan uint64, 1)
	go func() {
		peak := baseline.HeapAlloc
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var sample runtime.MemStats
				runtime.ReadMemStats(&sample)
				if sample.HeapAlloc > peak {
					peak = sample.HeapAlloc
				}
			case <-stop:
				var sample runtime.MemStats
				runtime.ReadMemStats(&sample)
				if sample.HeapAlloc > peak {
					peak = sample.HeapAlloc
				}
				done <- peak
				return
			}
		}
	}()

	actionError := action()
	close(stop)
	peak := <-done
	delta := uint64(0)
	if peak > baseline.HeapAlloc {
		delta = peak - baseline.HeapAlloc
	}
	return heapMeasurement{
		Baseline: baseline.HeapAlloc, Peak: peak, PeakDelta: delta, SoftLimit: softLimit,
	}, actionError
}
