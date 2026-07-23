package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type IdentityKind string

const (
	IdentityLocal  IdentityKind = "local"
	IdentityGoogle IdentityKind = "google"
)

type ResourceIdentity struct {
	Kind  IdentityKind
	Value string
}

func LocalIdentity(path string) (ResourceIdentity, error) {
	if path == "" {
		return ResourceIdentity{}, fmt.Errorf("local resource path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return ResourceIdentity{}, fmt.Errorf("resolve local resource path: %w", err)
	}
	canonical := filepath.Clean(absolute)
	if resolved, err := filepath.EvalSymlinks(canonical); err == nil {
		canonical = resolved
	} else if os.IsNotExist(err) {
		parent, parentErr := filepath.EvalSymlinks(filepath.Dir(canonical))
		if parentErr == nil {
			canonical = filepath.Join(parent, filepath.Base(canonical))
		}
	} else {
		return ResourceIdentity{}, fmt.Errorf("resolve local resource identity: %w", err)
	}
	if runtime.GOOS == "windows" {
		canonical = strings.ToLower(canonical)
	}
	return ResourceIdentity{Kind: IdentityLocal, Value: canonical}, nil
}

func GoogleIdentity(fileID string) (ResourceIdentity, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return ResourceIdentity{}, fmt.Errorf("Google file id is required")
	}
	return ResourceIdentity{Kind: IdentityGoogle, Value: fileID}, nil
}

func ValidateDistinctOutputIdentities(sources, outputs []ResourceIdentity) error {
	for outputIndex, output := range outputs {
		if output.Kind != IdentityLocal && output.Kind != IdentityGoogle {
			return fmt.Errorf("outputs[%d]: unsupported resource identity kind %q", outputIndex, output.Kind)
		}
		for sourceIndex, source := range sources {
			if source.Kind == output.Kind && source.Value == output.Value {
				return fmt.Errorf("outputs[%d] resolves to source[%d] identity %q; source files are immutable", outputIndex, sourceIndex, output.Value)
			}
		}
	}
	return nil
}
