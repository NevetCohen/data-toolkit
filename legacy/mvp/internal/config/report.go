package config

// ReportSettings returns a complete non-secret snapshot suitable for
// contract.FinalResult.ResolvedSettings.
func (resolved Resolved) ReportSettings() map[string]any {
	return map[string]any{
		"schema_version": resolved.SchemaVersion,
		"display": map[string]any{
			"null_display":      resolved.NullDisplay,
			"decimal_precision": resolved.DecimalPrecision,
			"date_format":       resolved.DateFormat,
			"time_format":       resolved.TimeFormat,
			"phone_format":      resolved.PhoneFormat,
			"text_encoding":     resolved.TextEncoding,
		},
		"aliases": append([]ColumnAlias(nil), resolved.Aliases...),
		"styles":  append([]TableStyle(nil), resolved.Styles...),
		"output": map[string]any{
			"directory":         resolved.OutputDirectory,
			"filename_template": resolved.FilenameTemplate,
			"table_style":       resolved.TableStyle,
			"txt_delimiter":     resolved.TXTDelimiter,
			"collision":         resolved.Collision,
			"format":            resolved.OutputFormat,
		},
		"cleaning": map[string]any{
			"mixed_type_action":        resolved.MixedTypeAction,
			"deduplicate_keep":         resolved.DeduplicateKeep,
			"exception_detail":         resolved.ExceptionDetail,
			"auto_normalize_formats":   resolved.AutoNormalizeFormats,
			"trim_detected_whitespace": resolved.TrimDetectedWhitespace,
			"not_scope":                resolved.NotScope,
		},
		"runtime": map[string]any{
			"maximum_memory_bytes": resolved.MaximumMemoryBytes,
			"lock_retry_count":     resolved.LockRetryCount,
			"lock_retry_interval":  resolved.LockRetryInterval.String(),
			"lock_timeout":         resolved.LockTimeout.String(),
			"temporary_workspace":  resolved.TemporaryWorkspace,
			"failed_run_retention": resolved.FailedRunRetention.String(),
			"recent_workflows":     resolved.RecentWorkflows,
		},
	}
}
