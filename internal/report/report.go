// Package report defines deterministic application events and validation
// findings shared across the V1 foundation.
package report

import (
	"data-toolkit/internal/core"
	"data-toolkit/internal/core/locator"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Event struct {
	Sequence  uint64 `json:"sequence"`
	Kind      string `json:"kind"`
	Component string `json:"component,omitempty"`
	Message   string `json:"message,omitempty"`
}

type Finding struct {
	Code             string                 `json:"code"`
	Severity         core.FindingSeverity   `json:"severity"`
	Message          string                 `json:"message"`
	Locator          *locator.SourceLocator `json:"locator,omitempty"`
	RequiresDecision bool                   `json:"requires_decision"`
}

// ValidationKind identifies one of the fixed V1 reopened-output validations.
type ValidationKind string

const (
	ValidationKindFormat       ValidationKind = "format"
	ValidationKindSchema       ValidationKind = "schema"
	ValidationKindExactColumns ValidationKind = "exact_columns"
	ValidationKindAllRowsMatch ValidationKind = "all_rows_match"
)

// ValidationResult reports one reopened-output validation result.
// A locator is meaningful only for a failed validation and identifies the
// first precise location in the reopened output when one is available.
type ValidationResult struct {
	Kind    ValidationKind         `json:"kind"`
	Passed  bool                   `json:"passed"`
	Code    string                 `json:"code,omitempty"`
	Message string                 `json:"message"`
	Locator *locator.OutputLocator `json:"locator,omitempty"`
}

// ExceptionSummary reports the bounded exception information retained in a
// terminal run report. The first locator and external report path are omitted
// when the effective reporting policy does not provide them.
type ExceptionSummary struct {
	Action       core.ExceptionAction   `json:"action"`
	Count        uint64                 `json:"count"`
	FirstLocator *locator.SourceLocator `json:"first_locator,omitempty"`
	ReportPath   *string                `json:"report_path,omitempty"`
}

// PublishedOutput identifies the validated bytes atomically published to the
// final local output path.
type PublishedOutput struct {
	Path      string            `json:"path"`
	FormatID  core.FormatID     `json:"format_id"`
	SHA256    core.SHA256Digest `json:"sha256"`
	SizeBytes int64             `json:"size_bytes"`
}

// Validate checks the exact V1 published-output contract.
func (output PublishedOutput) Validate() error {
	if strings.TrimSpace(output.Path) == "" {
		return errors.New("published output path is required")
	}
	if err := output.FormatID.Validate(); err != nil {
		return fmt.Errorf("published output format: %w", err)
	}
	if err := output.SHA256.Validate(); err != nil {
		return fmt.Errorf("published output hash: %w", err)
	}
	if output.SizeBytes < 0 {
		return errors.New("published output size must be non-negative")
	}
	return nil
}

// Validate checks the fixed V1 exception summary contract.
func (summary ExceptionSummary) Validate() error {
	if err := summary.Action.Validate(); err != nil {
		return fmt.Errorf("exception summary action: %w", err)
	}
	return nil
}

// Validate checks the fixed V1 validation contract.
func (result ValidationResult) Validate() error {
	switch result.Kind {
	case ValidationKindFormat, ValidationKindSchema, ValidationKindExactColumns, ValidationKindAllRowsMatch:
	default:
		return fmt.Errorf("validation kind %q is invalid", result.Kind)
	}
	if result.Message == "" {
		return errors.New("validation message is required")
	}
	if result.Passed && result.Locator != nil {
		return errors.New("validation locator is only allowed for a failed validation")
	}
	if result.Locator != nil && result.Locator.OutputPath == "" {
		return errors.New("validation locator output path is required")
	}
	return nil
}

// ErrorDetail is one stable, display-safe detail attached to an API error.
type ErrorDetail struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

const redactedErrorDetailValue = "[redacted]"

// NewErrorDetails returns deterministically ordered details with sensitive
// values redacted before they can reach an error, report, or receipt.
func NewErrorDetails(values map[string]string) []ErrorDetail {
	details := make([]ErrorDetail, 0, len(values))
	for key, value := range values {
		details = append(details, ErrorDetail{
			Key:   key,
			Value: redactErrorDetailValue(key, value),
		})
	}

	sort.Slice(details, func(left, right int) bool {
		return details[left].Key < details[right].Key
	})
	return details
}

func redactErrorDetailValue(key, value string) string {
	for _, part := range strings.FieldsFunc(strings.ToLower(key), func(r rune) bool {
		return r < 'a' || r > 'z'
	}) {
		switch part {
		case "authorization", "credential", "credentials", "cookie", "password", "secret", "token":
			return redactedErrorDetailValue
		}
	}
	return value
}
