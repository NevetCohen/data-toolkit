package testsupport

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSnapshotFileDetectsContentChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.csv")
	if err := os.WriteFile(path, []byte("a,b\n1,2\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	before, err := SnapshotFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("a,b\n1,3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	after, err := SnapshotFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if before.SHA256 == after.SHA256 {
		t.Fatal("content change did not change the snapshot digest")
	}
}

func TestNewGeneratedDirUsesAndCleansGeneratedRoot(t *testing.T) {
	repositoryRoot := t.TempDir()
	generatedDir := NewGeneratedDir(t, repositoryRoot, "workflow")
	wantParent := filepath.Join(repositoryRoot, "testdata", "generated")
	if filepath.Dir(generatedDir) != wantParent {
		t.Fatalf("generated directory parent = %q, want %q", filepath.Dir(generatedDir), wantParent)
	}
}
