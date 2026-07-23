package adapter

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"data-toolkit/internal/contract"
)

func TestStagedOutputPublishesOnlyAfterValidation(t *testing.T) {
	finalPath := filepath.Join(t.TempDir(), "output.csv")
	output, err := NewStagedOutput(finalPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output.Path(), []byte("name\nנועה 🧭\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Publish(contract.CollisionBlock); err == nil {
		t.Fatal("unvalidated output was published")
	}
	if _, err := os.Stat(finalPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("final output exists before validation: %v", err)
	}
	if err := output.Validate(context.Background(), func(_ context.Context, path string) error {
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(contents), "נועה 🧭") {
			return errors.New("missing expected Unicode row")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	published, err := output.Publish(contract.CollisionBlock)
	if err != nil {
		t.Fatal(err)
	}
	if published != finalPath {
		t.Fatalf("published path = %q, want %q", published, finalPath)
	}
}

func TestCollisionPolicies(t *testing.T) {
	root := t.TempDir()
	finalPath := filepath.Join(root, "report.csv")
	if err := os.WriteFile(finalPath, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}

	blocked := validatedStage(t, finalPath, "blocked")
	if _, err := blocked.Publish(contract.CollisionBlock); err == nil {
		t.Fatal("block policy replaced an existing output")
	}
	if _, err := blocked.Abort(false); err != nil {
		t.Fatal(err)
	}

	alternate := validatedStage(t, finalPath, "alternate")
	openTarget, err := os.Open(finalPath)
	if err != nil {
		t.Fatal(err)
	}
	alternatePath, err := alternate.Publish(contract.CollisionAlternateName)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(alternatePath) != "report (1).csv" {
		t.Fatalf("alternate path = %q", alternatePath)
	}
	if err := openTarget.Close(); err != nil {
		t.Fatal(err)
	}

	overwrite := validatedStage(t, finalPath, "replacement")
	if _, err := overwrite.Publish(contract.CollisionOverwrite); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "replacement" {
		t.Fatalf("overwritten contents = %q", contents)
	}
}

func validatedStage(t *testing.T, finalPath, contents string) *StagedOutput {
	t.Helper()
	output, err := NewStagedOutput(finalPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output.Path(), []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := output.Validate(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	return output
}
