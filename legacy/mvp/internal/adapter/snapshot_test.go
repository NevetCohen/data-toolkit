package adapter

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestProcessingUsesValidatedImmutableSnapshot(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "source.csv")
	original := []byte("name\nנועה 🧭\n")
	if err := os.WriteFile(sourcePath, original, 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err := CreateSourceSnapshot(context.Background(), sourcePath, filepath.Join(root, "run"), RetryPolicy{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(snapshot.SnapshotPath, 0o600) })
	if snapshot.OriginalPath == snapshot.SnapshotPath || snapshot.Size != int64(len(original)) {
		t.Fatalf("snapshot provenance = %#v", snapshot)
	}

	if err := os.WriteFile(sourcePath, []byte("name\nchanged\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := snapshot.OpenForProcessing()
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	got, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("processing bytes = %q, want snapshotted %q", got, original)
	}
	digest, size, err := hashFile(snapshot.SnapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	if digest != snapshot.SHA256 || size != snapshot.Size {
		t.Fatal("snapshot provenance no longer matches snapshot")
	}
}
