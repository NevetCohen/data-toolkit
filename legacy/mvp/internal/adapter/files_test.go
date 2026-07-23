package adapter

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenSharedReadRetriesExclusiveLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.csv")
	if err := os.WriteFile(path, []byte("a\n1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	attempts := 0
	file, err := openSharedRead(context.Background(), path, RetryPolicy{
		RetryCount: 3, Interval: time.Millisecond, Timeout: time.Second,
	}, func(path string) (*os.File, error) {
		attempts++
		if attempts < 3 {
			return nil, ErrFileLocked
		}
		return os.Open(path)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestOpenSharedReadReturnsPreciseLockError(t *testing.T) {
	_, err := openSharedRead(context.Background(), "locked.xlsx", RetryPolicy{
		RetryCount: 2, Interval: 0, Timeout: time.Second,
	}, func(string) (*os.File, error) {
		return nil, ErrFileLocked
	})
	var lockError *LockError
	if !errors.As(err, &lockError) {
		t.Fatalf("error type = %T, want LockError: %v", err, err)
	}
	if lockError.Path != "locked.xlsx" || lockError.Attempts != 3 {
		t.Fatalf("lock error = %#v", lockError)
	}
}

func TestOpenSharedReadHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := OpenSharedRead(ctx, "source.csv", RetryPolicy{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
