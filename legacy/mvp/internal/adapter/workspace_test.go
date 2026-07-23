package adapter

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"data-toolkit/internal/contract"
)

func TestCanceledRunStillCleansTemporaryWorkspace(t *testing.T) {
	root := t.TempDir()
	workspace, err := NewRunWorkspace(root, "run-canceled", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace.Path(), "temporary.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	retained, err := workspace.Finalize(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	if retained {
		t.Fatal("canceled workspace was retained despite zero retention")
	}
	if _, err := os.Stat(workspace.Path()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("workspace still exists: %v", err)
	}
}

func TestFailedRunRetentionAndExpiry(t *testing.T) {
	root := t.TempDir()
	workspace, err := NewRunWorkspace(root, "run-failed", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	retained, err := workspace.Finalize(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !retained {
		t.Fatal("failed workspace was not retained")
	}
	if err := CleanupExpiredWorkspaces(root, time.Now().UTC().Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(workspace.Path()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expired workspace still exists: %v", err)
	}
}

func TestFailedValidationNeverPublishesAndHonorsStageRetention(t *testing.T) {
	finalPath := filepath.Join(t.TempDir(), "output.csv")
	stage, err := NewStagedOutput(finalPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stage.Path(), []byte("invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := stage.Validate(context.Background(), func(context.Context, string) error {
		return errors.New("structural validation failed")
	}); err == nil {
		t.Fatal("expected validation failure")
	}
	if _, err := stage.Publish(contract.CollisionBlock); err == nil {
		t.Fatal("failed stage was published")
	}
	if _, err := os.Stat(finalPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("final output exists after failed validation: %v", err)
	}
	retainedPath, err := stage.Abort(true)
	if err != nil {
		t.Fatal(err)
	}
	if retainedPath == "" {
		t.Fatal("failed stage was not retained")
	}
	if _, err := os.Stat(retainedPath); err != nil {
		t.Fatal(err)
	}
	if _, err := stage.Abort(false); err != nil {
		t.Fatal(err)
	}
}
