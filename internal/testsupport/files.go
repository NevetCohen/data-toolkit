// Package testsupport provides shared invariants for local integration and
// acceptance tests. It is not part of the runtime dependency graph.
package testsupport

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// FileSnapshot records the identity and content of an immutable source file.
type FileSnapshot struct {
	Path   string
	Size   int64
	SHA256 [sha256.Size]byte
}

// SnapshotFile captures the current state of a fixture or source file.
func SnapshotFile(path string) (FileSnapshot, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return FileSnapshot{}, fmt.Errorf("resolve source path: %w", err)
	}

	file, err := os.Open(absolutePath)
	if err != nil {
		return FileSnapshot{}, fmt.Errorf("open source %q: %w", absolutePath, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return FileSnapshot{}, fmt.Errorf("stat source %q: %w", absolutePath, err)
	}
	if !info.Mode().IsRegular() {
		return FileSnapshot{}, fmt.Errorf("source %q is not a regular file", absolutePath)
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return FileSnapshot{}, fmt.Errorf("hash source %q: %w", absolutePath, err)
	}

	var digest [sha256.Size]byte
	copy(digest[:], hash.Sum(nil))
	return FileSnapshot{Path: absolutePath, Size: info.Size(), SHA256: digest}, nil
}

// AssertFileUnchanged fails the test when the snapshotted source was removed,
// replaced, truncated, or modified in place.
func AssertFileUnchanged(tb testing.TB, before FileSnapshot) {
	tb.Helper()

	after, err := SnapshotFile(before.Path)
	if err != nil {
		tb.Fatalf("immutable source changed: %v", err)
	}
	if after.Size != before.Size || after.SHA256 != before.SHA256 {
		tb.Fatalf("immutable source %q changed during the test", before.Path)
	}
}

// NewGeneratedDir creates an isolated disposable directory beneath
// testdata/generated and registers cleanup with the calling test.
func NewGeneratedDir(tb testing.TB, repositoryRoot, prefix string) string {
	tb.Helper()

	generatedRoot := filepath.Join(repositoryRoot, "testdata", "generated")
	if err := os.MkdirAll(generatedRoot, 0o755); err != nil {
		tb.Fatalf("create generated-output root: %v", err)
	}

	directory, err := os.MkdirTemp(generatedRoot, prefix+"-")
	if err != nil {
		tb.Fatalf("create generated-output directory: %v", err)
	}
	tb.Cleanup(func() {
		if err := os.RemoveAll(directory); err != nil {
			tb.Errorf("remove generated-output directory %q: %v", directory, err)
		}
	})
	return directory
}
