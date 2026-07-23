package adapter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"data-toolkit/internal/contract"
)

type StagedOutput struct {
	finalPath       string
	stagedPath      string
	validated       bool
	validationError error
	published       bool
}

func NewStagedOutput(finalPath string) (*StagedOutput, error) {
	absoluteFinal, err := filepath.Abs(finalPath)
	if err != nil {
		return nil, fmt.Errorf("resolve output path: %w", err)
	}
	directory := filepath.Dir(absoluteFinal)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}
	file, err := os.CreateTemp(directory, ".data-toolkit-*.stage")
	if err != nil {
		return nil, fmt.Errorf("create staged output: %w", err)
	}
	stagedPath := file.Name()
	if err := file.Close(); err != nil {
		os.Remove(stagedPath)
		return nil, fmt.Errorf("close staged output: %w", err)
	}
	return &StagedOutput{finalPath: absoluteFinal, stagedPath: stagedPath}, nil
}

func (output *StagedOutput) Path() string {
	return output.stagedPath
}

func (output *StagedOutput) Validate(ctx context.Context, validate func(context.Context, string) error) error {
	if output.published {
		return errors.New("output is already published")
	}
	if err := ctx.Err(); err != nil {
		output.validationError = err
		return err
	}
	info, err := os.Stat(output.stagedPath)
	if err != nil {
		output.validationError = fmt.Errorf("stat staged output: %w", err)
		return output.validationError
	}
	if !info.Mode().IsRegular() {
		output.validationError = errors.New("staged output is not a regular file")
		return output.validationError
	}
	if validate != nil {
		if err := validate(ctx, output.stagedPath); err != nil {
			output.validationError = fmt.Errorf("validate staged output: %w", err)
			return output.validationError
		}
	}
	output.validated = true
	output.validationError = nil
	return nil
}

func (output *StagedOutput) Publish(action contract.CollisionAction) (string, error) {
	if output.published {
		return "", errors.New("output is already published")
	}
	if !output.validated {
		if output.validationError != nil {
			return "", fmt.Errorf("staged output failed validation: %w", output.validationError)
		}
		return "", errors.New("staged output must be validated before publication")
	}

	target, err := resolvePublicationPath(output.finalPath, action)
	if err != nil {
		return "", err
	}
	if action == contract.CollisionOverwrite {
		if _, err := os.Stat(target); err == nil {
			if err := os.Remove(target); err != nil {
				return "", fmt.Errorf("overwrite output %q: %w", target, err)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("inspect output %q: %w", target, err)
		}
	}
	if err := os.Rename(output.stagedPath, target); err != nil {
		return "", fmt.Errorf("publish staged output %q to %q: %w", output.stagedPath, target, err)
	}
	output.published = true
	return target, nil
}

// Abort removes an unpublished stage unless failure retention is requested.
// It returns the retained path when the stage is kept.
func (output *StagedOutput) Abort(retain bool) (string, error) {
	if output.published {
		return "", nil
	}
	if retain {
		return output.stagedPath, nil
	}
	if err := os.Remove(output.stagedPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("remove staged output %q: %w", output.stagedPath, err)
	}
	return "", nil
}

func resolvePublicationPath(finalPath string, action contract.CollisionAction) (string, error) {
	_, err := os.Stat(finalPath)
	exists := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("inspect output %q: %w", finalPath, err)
	}
	switch action {
	case contract.CollisionBlock:
		if exists {
			return "", fmt.Errorf("output target %q already exists and collision policy is block", finalPath)
		}
		return finalPath, nil
	case contract.CollisionOverwrite:
		return finalPath, nil
	case contract.CollisionAlternateName:
		if !exists {
			return finalPath, nil
		}
		extension := filepath.Ext(finalPath)
		base := finalPath[:len(finalPath)-len(extension)]
		for suffix := 1; ; suffix++ {
			candidate := fmt.Sprintf("%s (%d)%s", base, suffix, extension)
			if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
				return candidate, nil
			} else if err != nil {
				return "", fmt.Errorf("inspect alternate output %q: %w", candidate, err)
			}
		}
	default:
		return "", fmt.Errorf("unsupported collision action %q", action)
	}
}
