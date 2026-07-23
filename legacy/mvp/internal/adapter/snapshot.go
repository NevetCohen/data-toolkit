package adapter

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type SourceSnapshot struct {
	OriginalPath string
	SnapshotPath string
	Size         int64
	SHA256       [sha256.Size]byte
	CapturedAt   time.Time
}

func CreateSourceSnapshot(ctx context.Context, sourcePath, runDirectory string, retry RetryPolicy) (SourceSnapshot, error) {
	absoluteSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return SourceSnapshot{}, fmt.Errorf("resolve source path: %w", err)
	}
	if runDirectory == "" {
		return SourceSnapshot{}, errors.New("run snapshot directory is required")
	}
	if err := os.MkdirAll(runDirectory, 0o700); err != nil {
		return SourceSnapshot{}, fmt.Errorf("create run snapshot directory: %w", err)
	}

	source, err := OpenSharedRead(ctx, absoluteSource, retry)
	if err != nil {
		return SourceSnapshot{}, err
	}
	defer source.Close()

	sourceInfo, err := source.Stat()
	if err != nil {
		return SourceSnapshot{}, fmt.Errorf("stat source %q: %w", absoluteSource, err)
	}
	if !sourceInfo.Mode().IsRegular() {
		return SourceSnapshot{}, fmt.Errorf("source %q is not a regular file", absoluteSource)
	}

	extension := filepath.Ext(absoluteSource)
	snapshotFile, err := os.CreateTemp(runDirectory, "source-*"+extension)
	if err != nil {
		return SourceSnapshot{}, fmt.Errorf("create source snapshot: %w", err)
	}
	snapshotPath := snapshotFile.Name()
	keepSnapshot := false
	defer func() {
		snapshotFile.Close()
		if !keepSnapshot {
			os.Remove(snapshotPath)
		}
	}()

	hash := sha256.New()
	copied, err := io.Copy(io.MultiWriter(snapshotFile, hash), &contextReader{ctx: ctx, reader: source})
	if err != nil {
		return SourceSnapshot{}, fmt.Errorf("copy source snapshot: %w", err)
	}
	if copied != sourceInfo.Size() {
		return SourceSnapshot{}, fmt.Errorf("validate source snapshot size: copied %d bytes, expected %d", copied, sourceInfo.Size())
	}
	if err := snapshotFile.Sync(); err != nil {
		return SourceSnapshot{}, fmt.Errorf("sync source snapshot: %w", err)
	}
	if err := snapshotFile.Close(); err != nil {
		return SourceSnapshot{}, fmt.Errorf("close source snapshot: %w", err)
	}

	var digest [sha256.Size]byte
	copy(digest[:], hash.Sum(nil))
	validatedDigest, validatedSize, err := hashFile(snapshotPath)
	if err != nil {
		return SourceSnapshot{}, fmt.Errorf("validate source snapshot: %w", err)
	}
	if validatedSize != copied || validatedDigest != digest {
		return SourceSnapshot{}, errors.New("source snapshot failed size or SHA-256 validation")
	}
	if err := os.Chmod(snapshotPath, 0o400); err != nil {
		return SourceSnapshot{}, fmt.Errorf("make source snapshot read-only: %w", err)
	}

	absoluteSnapshot, err := filepath.Abs(snapshotPath)
	if err != nil {
		return SourceSnapshot{}, fmt.Errorf("resolve snapshot path: %w", err)
	}
	keepSnapshot = true
	return SourceSnapshot{
		OriginalPath: absoluteSource,
		SnapshotPath: absoluteSnapshot,
		Size:         copied,
		SHA256:       digest,
		CapturedAt:   time.Now().UTC(),
	}, nil
}

func (snapshot SourceSnapshot) OpenForProcessing() (*os.File, error) {
	file, err := os.Open(snapshot.SnapshotPath)
	if err != nil {
		return nil, fmt.Errorf("open validated source snapshot %q: %w", snapshot.SnapshotPath, err)
	}
	return file, nil
}

func hashFile(path string) ([sha256.Size]byte, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return [sha256.Size]byte{}, 0, err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return [sha256.Size]byte{}, 0, err
	}
	var digest [sha256.Size]byte
	copy(digest[:], hash.Sum(nil))
	return digest, size, nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (reader *contextReader) Read(buffer []byte) (int, error) {
	if err := reader.ctx.Err(); err != nil {
		return 0, err
	}
	return reader.reader.Read(buffer)
}
