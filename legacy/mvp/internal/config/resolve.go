package config

import (
	"fmt"
	"strings"
	"time"

	"data-toolkit/internal/contract"
)

// ValidateConfig checks identities and values in a partial user document.
func ValidateConfig(configuration Config) error {
	if configuration.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("$.schema_version: unsupported configuration version %q", configuration.SchemaVersion)
	}
	aliases := make(map[string]struct{}, len(configuration.Aliases))
	for index, alias := range configuration.Aliases {
		key := strings.ToLower(alias.Name)
		if _, exists := aliases[key]; exists {
			return fmt.Errorf("$.aliases[%d].name: duplicate alias identity %q", index, alias.Name)
		}
		aliases[key] = struct{}{}
	}
	styles := make(map[string]struct{}, len(configuration.Styles))
	for index, style := range configuration.Styles {
		key := strings.ToLower(style.Name)
		if _, exists := styles[key]; exists {
			return fmt.Errorf("$.styles[%d].name: duplicate style identity %q", index, style.Name)
		}
		styles[key] = struct{}{}
	}
	for path, value := range map[string]*string{
		"$.runtime.lock_retry_interval":  configuration.Runtime.LockRetryInterval,
		"$.runtime.lock_timeout":         configuration.Runtime.LockTimeout,
		"$.runtime.failed_run_retention": configuration.Runtime.FailedRunRetention,
	} {
		if value != nil {
			if _, err := time.ParseDuration(*value); err != nil {
				return fmt.Errorf("%s: invalid duration %q: %w", path, *value, err)
			}
		}
	}
	return nil
}

// Resolve applies built-ins, user settings, workflow overrides, and one
// output override in ascending precedence order.
func Resolve(defaults Resolved, user Config, workflow contract.WorkflowOverrides, output *contract.OutputOverrides) (Resolved, error) {
	if err := ValidateConfig(user); err != nil {
		return Resolved{}, err
	}
	resolved := defaults
	applyUserConfig(&resolved, user)
	applyWorkflowOverrides(&resolved, workflow)
	if output != nil {
		applyOutputOverrides(&resolved, *output)
	}
	if err := ValidateResolved(resolved); err != nil {
		return Resolved{}, err
	}
	return resolved, nil
}

func applyUserConfig(resolved *Resolved, user Config) {
	if user.Display.NullDisplay != nil {
		resolved.NullDisplay = *user.Display.NullDisplay
	}
	if user.Display.DecimalPrecision != nil {
		resolved.DecimalPrecision = *user.Display.DecimalPrecision
	}
	if user.Display.DateFormat != nil {
		resolved.DateFormat = *user.Display.DateFormat
	}
	if user.Display.TimeFormat != nil {
		resolved.TimeFormat = *user.Display.TimeFormat
	}
	if user.Display.PhoneFormat != nil {
		resolved.PhoneFormat = *user.Display.PhoneFormat
	}
	if user.Display.TextEncoding != nil {
		resolved.TextEncoding = *user.Display.TextEncoding
	}
	if user.Aliases != nil {
		resolved.Aliases = append([]ColumnAlias(nil), user.Aliases...)
	}
	if user.Styles != nil {
		resolved.Styles = append([]TableStyle(nil), user.Styles...)
	}
	if user.Output.Directory != nil {
		resolved.OutputDirectory = *user.Output.Directory
	}
	if user.Output.FilenameTemplate != nil {
		resolved.FilenameTemplate = *user.Output.FilenameTemplate
	}
	if user.Output.DefaultStyle != nil {
		resolved.TableStyle = *user.Output.DefaultStyle
	}
	if user.Output.TXTDelimiter != nil {
		resolved.TXTDelimiter = *user.Output.TXTDelimiter
	}
	if user.Output.Collision != nil {
		resolved.Collision = *user.Output.Collision
	}
	if user.Cleaning.MixedTypeAction != nil {
		resolved.MixedTypeAction = *user.Cleaning.MixedTypeAction
	}
	if user.Cleaning.DeduplicateKeep != nil {
		resolved.DeduplicateKeep = *user.Cleaning.DeduplicateKeep
	}
	if user.Cleaning.ExceptionDetail != nil {
		resolved.ExceptionDetail = *user.Cleaning.ExceptionDetail
	}
	if user.Cleaning.AutoNormalizeFormats != nil {
		resolved.AutoNormalizeFormats = *user.Cleaning.AutoNormalizeFormats
	}
	if user.Cleaning.TrimDetectedWhitespace != nil {
		resolved.TrimDetectedWhitespace = *user.Cleaning.TrimDetectedWhitespace
	}
	if user.Cleaning.NotScope != nil {
		resolved.NotScope = *user.Cleaning.NotScope
	}
	if user.Runtime.MaximumMemoryBytes != nil {
		resolved.MaximumMemoryBytes = *user.Runtime.MaximumMemoryBytes
	}
	if user.Runtime.LockRetryCount != nil {
		resolved.LockRetryCount = *user.Runtime.LockRetryCount
	}
	applyDuration(user.Runtime.LockRetryInterval, &resolved.LockRetryInterval)
	applyDuration(user.Runtime.LockTimeout, &resolved.LockTimeout)
	if user.Runtime.TemporaryWorkspace != nil {
		resolved.TemporaryWorkspace = *user.Runtime.TemporaryWorkspace
	}
	applyDuration(user.Runtime.FailedRunRetention, &resolved.FailedRunRetention)
	if user.Runtime.RecentWorkflows != nil {
		resolved.RecentWorkflows = *user.Runtime.RecentWorkflows
	}
}

func applyWorkflowOverrides(resolved *Resolved, workflow contract.WorkflowOverrides) {
	if workflow.NullDisplay != "" {
		resolved.NullDisplay = workflow.NullDisplay
	}
	if workflow.DecimalPrecision != nil {
		resolved.DecimalPrecision = *workflow.DecimalPrecision
	}
	if workflow.DateFormat != "" {
		resolved.DateFormat = workflow.DateFormat
	}
	if workflow.TimeFormat != "" {
		resolved.TimeFormat = workflow.TimeFormat
	}
	if workflow.PhoneFormat != "" {
		resolved.PhoneFormat = workflow.PhoneFormat
	}
	if workflow.TextEncoding != "" {
		resolved.TextEncoding = workflow.TextEncoding
	}
	if workflow.OutputDirectory != "" {
		resolved.OutputDirectory = workflow.OutputDirectory
	}
	if workflow.FilenameTemplate != "" {
		resolved.FilenameTemplate = workflow.FilenameTemplate
	}
	if workflow.TableStyle != "" {
		resolved.TableStyle = workflow.TableStyle
	}
	if workflow.TXTDelimiter != "" {
		resolved.TXTDelimiter = workflow.TXTDelimiter
	}
	if workflow.MixedTypeAction != "" {
		resolved.MixedTypeAction = workflow.MixedTypeAction
	}
	if workflow.DeduplicateKeep != "" {
		resolved.DeduplicateKeep = workflow.DeduplicateKeep
	}
	if workflow.ExceptionDetail != "" {
		resolved.ExceptionDetail = workflow.ExceptionDetail
	}
	if workflow.AutoNormalizeFormats != nil {
		resolved.AutoNormalizeFormats = *workflow.AutoNormalizeFormats
	}
	if workflow.TrimDetectedWhitespace != nil {
		resolved.TrimDetectedWhitespace = *workflow.TrimDetectedWhitespace
	}
	if workflow.NotScope != "" {
		resolved.NotScope = workflow.NotScope
	}
	if workflow.Collision != "" {
		resolved.Collision = workflow.Collision
	}
	if workflow.MaximumMemory > 0 {
		resolved.MaximumMemoryBytes = workflow.MaximumMemory
	}
}

func applyOutputOverrides(resolved *Resolved, output contract.OutputOverrides) {
	if output.Directory != "" {
		resolved.OutputDirectory = output.Directory
	}
	if output.Filename != "" {
		resolved.FilenameTemplate = output.Filename
	}
	if output.Format != "" {
		resolved.OutputFormat = output.Format
	}
	if output.Style != "" {
		resolved.TableStyle = output.Style
	}
	if output.Delimiter != "" {
		resolved.TXTDelimiter = output.Delimiter
	}
	if output.Collision != "" {
		resolved.Collision = output.Collision
	}
}

func applyDuration(value *string, target *time.Duration) {
	if value == nil {
		return
	}
	parsed, err := time.ParseDuration(*value)
	if err == nil {
		*target = parsed
	}
}

// ValidateResolved protects callers that construct configuration without the
// document loader.
func ValidateResolved(resolved Resolved) error {
	if resolved.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("schema_version: unsupported configuration version %q", resolved.SchemaVersion)
	}
	if resolved.MaximumMemoryBytes <= 0 {
		return fmt.Errorf("runtime.maximum_memory_bytes: must be greater than zero")
	}
	if resolved.DecimalPrecision < 0 {
		return fmt.Errorf("display.decimal_precision: must not be negative")
	}
	if resolved.TextEncoding != "UTF-8" {
		return fmt.Errorf("display.text_encoding: unsupported encoding %q", resolved.TextEncoding)
	}
	if resolved.OutputDirectory == "" || resolved.FilenameTemplate == "" || resolved.TXTDelimiter == "" {
		return fmt.Errorf("output: directory, filename_template, and txt_delimiter are required")
	}
	if !validCollision(resolved.Collision) {
		return fmt.Errorf("output.collision: unsupported action %q", resolved.Collision)
	}
	if !validMixedTypeAction(resolved.MixedTypeAction) {
		return fmt.Errorf("cleaning.mixed_type_action: unsupported action %q", resolved.MixedTypeAction)
	}
	if !validKeepPolicy(resolved.DeduplicateKeep) {
		return fmt.Errorf("cleaning.deduplicate_keep: unsupported policy %q", resolved.DeduplicateKeep)
	}
	if !validNotScope(resolved.NotScope) {
		return fmt.Errorf("cleaning.not_scope: unsupported scope %q", resolved.NotScope)
	}
	if resolved.LockRetryCount < 0 {
		return fmt.Errorf("runtime.lock_retry_count: must not be negative")
	}
	if resolved.LockRetryInterval < 0 || resolved.LockTimeout < 0 || resolved.FailedRunRetention < 0 {
		return fmt.Errorf("runtime: durations must not be negative")
	}
	if resolved.LockRetryCount > 0 && (resolved.LockRetryInterval == 0 || resolved.LockTimeout == 0) {
		return fmt.Errorf("runtime: positive lock retries require positive interval and timeout")
	}
	if resolved.TemporaryWorkspace == "" {
		return fmt.Errorf("runtime.temporary_workspace: path is required")
	}
	if resolved.RecentWorkflows < 0 {
		return fmt.Errorf("runtime.recent_workflows: must not be negative")
	}
	aliases := make(map[string]struct{}, len(resolved.Aliases))
	for index, alias := range resolved.Aliases {
		key := strings.ToLower(alias.Name)
		if _, exists := aliases[key]; exists {
			return fmt.Errorf("aliases[%d].name: duplicate alias identity %q", index, alias.Name)
		}
		aliases[key] = struct{}{}
	}
	styles := make(map[string]struct{}, len(resolved.Styles))
	for index, style := range resolved.Styles {
		key := strings.ToLower(style.Name)
		if _, exists := styles[key]; exists {
			return fmt.Errorf("styles[%d].name: duplicate style identity %q", index, style.Name)
		}
		styles[key] = struct{}{}
	}
	if _, exists := styles[strings.ToLower(resolved.TableStyle)]; !exists {
		return fmt.Errorf("output.default_style: unknown named style %q", resolved.TableStyle)
	}
	return nil
}

func validCollision(value contract.CollisionAction) bool {
	switch value {
	case contract.CollisionBlock, contract.CollisionOverwrite, contract.CollisionAlternateName:
		return true
	default:
		return false
	}
}

func validMixedTypeAction(value contract.MixedTypeAction) bool {
	switch value {
	case contract.MixedTypeBlock, contract.MixedTypeReport, contract.MixedTypeIgnore:
		return true
	default:
		return false
	}
}

func validKeepPolicy(value contract.KeepPolicy) bool {
	switch value {
	case contract.KeepFirst, contract.KeepLast, contract.KeepError:
		return true
	default:
		return false
	}
}

func validNotScope(value contract.NotScope) bool {
	switch value {
	case contract.NotScopeSource1, contract.NotScopeSource2, contract.NotScopeBoth:
		return true
	default:
		return false
	}
}
