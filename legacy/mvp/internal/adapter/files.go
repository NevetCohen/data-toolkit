package adapter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

var ErrFileLocked = errors.New("file is locked")

type RetryPolicy struct {
	RetryCount int
	Interval   time.Duration
	Timeout    time.Duration
}

type LockError struct {
	Path     string
	Attempts int
	Cause    error
}

func (err *LockError) Error() string {
	return fmt.Sprintf("source %q remained exclusively locked after %d attempts: %v; close the program holding the file and retry", err.Path, err.Attempts, err.Cause)
}

func (err *LockError) Unwrap() error {
	return err.Cause
}

func OpenSharedRead(ctx context.Context, path string, policy RetryPolicy) (*os.File, error) {
	return openSharedRead(ctx, path, policy, os.Open)
}

func openSharedRead(ctx context.Context, path string, policy RetryPolicy, opener func(string) (*os.File, error)) (*os.File, error) {
	if path == "" {
		return nil, errors.New("source path is required")
	}
	if policy.RetryCount < 0 || policy.Interval < 0 || policy.Timeout < 0 {
		return nil, errors.New("lock retry values must not be negative")
	}
	started := time.Now()
	attempts := policy.RetryCount + 1
	var lastError error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		file, err := opener(path)
		if err == nil {
			return file, nil
		}
		lastError = err
		if !isRetryableLock(err) {
			return nil, fmt.Errorf("open source %q for shared read: %w", path, err)
		}
		if attempt == attempts || (policy.Timeout > 0 && time.Since(started) >= policy.Timeout) {
			return nil, &LockError{Path: path, Attempts: attempt, Cause: lastError}
		}
		wait := policy.Interval
		if policy.Timeout > 0 {
			remaining := policy.Timeout - time.Since(started)
			if remaining <= 0 {
				return nil, &LockError{Path: path, Attempts: attempt, Cause: lastError}
			}
			if wait > remaining {
				wait = remaining
			}
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, &LockError{Path: path, Attempts: attempts, Cause: lastError}
}

func isRetryableLock(err error) bool {
	if errors.Is(err, ErrFileLocked) {
		return true
	}
	var pathError *os.PathError
	if errors.As(err, &pathError) {
		err = pathError.Err
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == syscall.Errno(32) || errno == syscall.Errno(33)
	}
	return false
}
