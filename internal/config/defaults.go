package config

import (
	"path/filepath"
	"time"

	"data-toolkit/internal/contract"
)

const (
	defaultMaximumMemory = int64(512 * 1024 * 1024)
)

// BuiltInDefaults returns conservative product defaults. The caller supplies
// the workspace root so machine-specific paths are not hard-coded.
func BuiltInDefaults(temporaryRoot string) Resolved {
	bold := true
	return Resolved{
		SchemaVersion:          CurrentSchemaVersion,
		NullDisplay:            "-",
		DecimalPrecision:       3,
		DateFormat:             "02/01/06",
		TimeFormat:             "15:04",
		PhoneFormat:            "052-6105412",
		TextEncoding:           "UTF-8",
		Styles:                 []TableStyle{{Name: "default", BoldHeaders: &bold}},
		OutputDirectory:        ".",
		FilenameTemplate:       "{{workflow_id}}-{{output_id}}-{{timestamp}}",
		TableStyle:             "default",
		TXTDelimiter:           "\t",
		Collision:              contract.CollisionBlock,
		MixedTypeAction:        contract.MixedTypeBlock,
		DeduplicateKeep:        contract.KeepFirst,
		ExceptionDetail:        "summary",
		AutoNormalizeFormats:   true,
		TrimDetectedWhitespace: true,
		NotScope:               contract.NotScopeBoth,
		MaximumMemoryBytes:     defaultMaximumMemory,
		LockRetryCount:         5,
		LockRetryInterval:      500 * time.Millisecond,
		LockTimeout:            5 * time.Second,
		TemporaryWorkspace:     filepath.Join(temporaryRoot, "data-toolkit"),
		FailedRunRetention:     24 * time.Hour,
		RecentWorkflows:        20,
	}
}
