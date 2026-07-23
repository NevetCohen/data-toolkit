package contract

// ValidateStartedResult checks the invariant of the pre-execution result.
func ValidateStartedResult(result StartedResult) error {
	findings := make(ValidationErrors, 0)
	if result.SchemaVersion != CurrentSchemaVersion {
		addFinding(&findings, "$.schema_version", "unsupported schema version %q", result.SchemaVersion)
	}
	if result.RunID == "" {
		addFinding(&findings, "$.run_id", "run id is required")
	}
	if result.WorkflowID == "" {
		addFinding(&findings, "$.workflow_id", "workflow id is required")
	}
	if result.Status != RunStatusStarted {
		addFinding(&findings, "$.status", "started result requires status %q", RunStatusStarted)
	}
	if result.StartedAt.IsZero() {
		addFinding(&findings, "$.started_at", "start time is required")
	}
	if len(findings) > 0 {
		return findings
	}
	return nil
}

// ValidateFinalResult checks the invariant of a terminal result.
func ValidateFinalResult(result FinalResult) error {
	findings := make(ValidationErrors, 0)
	if result.SchemaVersion != CurrentSchemaVersion {
		addFinding(&findings, "$.schema_version", "unsupported schema version %q", result.SchemaVersion)
	}
	if result.RunID == "" {
		addFinding(&findings, "$.run_id", "run id is required")
	}
	if result.WorkflowID == "" {
		addFinding(&findings, "$.workflow_id", "workflow id is required")
	}
	if result.Status != RunStatusSucceeded && result.Status != RunStatusFailed {
		addFinding(&findings, "$.status", "final result requires succeeded or failed status")
	}
	if result.StartedAt.IsZero() {
		addFinding(&findings, "$.started_at", "start time is required")
	}
	if result.FinishedAt.IsZero() {
		addFinding(&findings, "$.finished_at", "finish time is required")
	}
	if result.FinishedAt.Before(result.StartedAt) {
		addFinding(&findings, "$.finished_at", "finish time cannot precede start time")
	}
	if result.ResolvedSettings == nil {
		addFinding(&findings, "$.resolved_settings", "resolved non-secret settings are required")
	}
	if result.Status == RunStatusFailed && len(result.Errors) == 0 {
		addFinding(&findings, "$.errors", "failed result requires at least one detailed error")
	}
	if len(findings) > 0 {
		return findings
	}
	return nil
}
