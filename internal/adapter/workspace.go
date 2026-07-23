package adapter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const retentionMarker = ".retain-until"

var runIDPattern = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9._-]*$")

type RunWorkspace struct {
	root      string
	path      string
	retention time.Duration
	finalized bool
}

func NewRunWorkspace(root, runID string, failedRunRetention time.Duration) (*RunWorkspace, error) {
	if failedRunRetention < 0 {
		return nil, errors.New("failed-run retention must not be negative")
	}
	if !runIDPattern.MatchString(runID) || runID == "." || runID == ".." {
		return nil, fmt.Errorf("invalid run id %q", runID)
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}
	if err := os.MkdirAll(absoluteRoot, 0o700); err != nil {
		return nil, fmt.Errorf("create workspace root: %w", err)
	}
	path := filepath.Join(absoluteRoot, runID)
	if err := validateChildPath(absoluteRoot, path); err != nil {
		return nil, err
	}
	if err := os.Mkdir(path, 0o700); err != nil {
		return nil, fmt.Errorf("create run workspace: %w", err)
	}
	return &RunWorkspace{root: absoluteRoot, path: path, retention: failedRunRetention}, nil
}

func (workspace *RunWorkspace) Path() string {
	return workspace.path
}

// Finalize intentionally completes cleanup even when ctx is already canceled.
func (workspace *RunWorkspace) Finalize(_ context.Context, success bool) (bool, error) {
	if workspace.finalized {
		return false, nil
	}
	workspace.finalized = true
	if !success && workspace.retention > 0 {
		retainUntil := time.Now().UTC().Add(workspace.retention).Format(time.RFC3339Nano)
		if err := os.WriteFile(filepath.Join(workspace.path, retentionMarker), []byte(retainUntil), 0o600); err != nil {
			return false, fmt.Errorf("write failed-run retention marker: %w", err)
		}
		return true, nil
	}
	return false, removeRunWorkspace(workspace.root, workspace.path)
}

func CleanupExpiredWorkspaces(root string, now time.Time) error {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve workspace root: %w", err)
	}
	entries, err := os.ReadDir(absoluteRoot)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read workspace root: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(absoluteRoot, entry.Name())
		marker, err := os.ReadFile(filepath.Join(path, retentionMarker))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read retention marker for %q: %w", path, err)
		}
		retainUntil, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(string(marker)))
		if err != nil {
			return fmt.Errorf("parse retention marker for %q: %w", path, err)
		}
		if !now.Before(retainUntil) {
			if err := removeRunWorkspace(absoluteRoot, path); err != nil {
				return err
			}
		}
	}
	return nil
}

func removeRunWorkspace(root, path string) error {
	if err := validateChildPath(root, path); err != nil {
		return err
	}
	_ = filepath.WalkDir(path, func(current string, entry os.DirEntry, walkErr error) error {
		if walkErr == nil {
			if entry.IsDir() {
				_ = os.Chmod(current, 0o700)
			} else {
				_ = os.Chmod(current, 0o600)
			}
		}
		return nil
	})
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove run workspace %q: %w", path, err)
	}
	return nil
}

func validateChildPath(root, path string) error {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return fmt.Errorf("validate run workspace path: %w", err)
	}
	if relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return fmt.Errorf("run workspace %q escapes root %q", path, root)
	}
	return nil
}
