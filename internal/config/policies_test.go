package config

import (
	"path/filepath"
	"testing"
	"time"

	"data-toolkit/internal/contract"
)

func TestOutputCleaningAndRepairSettingsResolve(t *testing.T) {
	directory := "configured-output"
	template := "{{workflow_id}}.txt"
	delimiter := ","
	collision := contract.CollisionOverwrite
	mixed := contract.MixedTypeReport
	detail := "full"
	autoNormalize := false
	trim := false
	user := Config{
		SchemaVersion: CurrentSchemaVersion,
		Output: OutputConfig{
			Directory:        &directory,
			FilenameTemplate: &template,
			TXTDelimiter:     &delimiter,
			Collision:        &collision,
		},
		Cleaning: CleaningConfig{
			MixedTypeAction:        &mixed,
			ExceptionDetail:        &detail,
			AutoNormalizeFormats:   &autoNormalize,
			TrimDetectedWhitespace: &trim,
		},
	}

	resolved, err := Resolve(BuiltInDefaults("temporary"), user, contract.WorkflowOverrides{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.OutputDirectory != directory ||
		resolved.FilenameTemplate != template ||
		resolved.TXTDelimiter != delimiter ||
		resolved.Collision != collision ||
		resolved.MixedTypeAction != mixed ||
		resolved.ExceptionDetail != detail ||
		resolved.AutoNormalizeFormats ||
		resolved.TrimDetectedWhitespace {
		t.Fatalf("settings were not resolved: %#v", resolved)
	}
}

func TestRuntimeSafetySettingsResolveWithoutFixedMachinePath(t *testing.T) {
	memory := int64(256 * 1024 * 1024)
	retries := 3
	interval := "250ms"
	timeout := "2s"
	workspace := filepath.Join("custom", "temporary")
	retention := "6h"
	recent := 7
	user := Config{
		SchemaVersion: CurrentSchemaVersion,
		Runtime: RuntimeConfig{
			MaximumMemoryBytes: &memory,
			LockRetryCount:     &retries,
			LockRetryInterval:  &interval,
			LockTimeout:        &timeout,
			TemporaryWorkspace: &workspace,
			FailedRunRetention: &retention,
			RecentWorkflows:    &recent,
		},
	}

	resolved, err := Resolve(BuiltInDefaults("caller-owned-root"), user, contract.WorkflowOverrides{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.MaximumMemoryBytes != memory ||
		resolved.LockRetryCount != retries ||
		resolved.LockRetryInterval != 250*time.Millisecond ||
		resolved.LockTimeout != 2*time.Second ||
		resolved.TemporaryWorkspace != workspace ||
		resolved.FailedRunRetention != 6*time.Hour ||
		resolved.RecentWorkflows != recent {
		t.Fatalf("runtime settings were not resolved: %#v", resolved)
	}
}

func TestWorkflowCanOverrideMaximumMemory(t *testing.T) {
	resolved, err := Resolve(
		BuiltInDefaults("temporary"),
		Config{SchemaVersion: CurrentSchemaVersion},
		contract.WorkflowOverrides{MaximumMemory: 128 * 1024 * 1024},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.MaximumMemoryBytes != 128*1024*1024 {
		t.Fatalf("memory = %d", resolved.MaximumMemoryBytes)
	}
}
