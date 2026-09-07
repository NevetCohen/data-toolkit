package application

import (
	"errors"
	"fmt"
	"time"

	"data-toolkit/internal/core"
	"data-toolkit/internal/report"
)

// RunReport is the complete terminal result assembled by writer.finalize.
// Required slices are non-nil so their JSON representation is always an array.
type RunReport struct {
	RequestID             core.RequestID            `json:"request_id"`
	RunID                 core.RunID                `json:"run_id"`
	Status                core.TerminalStatus       `json:"status"`
	Output                *report.PublishedOutput   `json:"output,omitempty"`
	Validations           []report.ValidationResult `json:"validations"`
	Findings              []report.Finding          `json:"findings"`
	ExceptionSummary      *report.ExceptionSummary  `json:"exception_summary,omitempty"`
	FirstFailure          *APIError                 `json:"first_failure,omitempty"`
	SourceSnapshotHash    *core.SHA256Digest        `json:"source_snapshot_hash,omitempty"`
	EffectiveConfigDigest core.SHA256Digest         `json:"effective_config_digest"`
	CapabilitiesDigest    core.SHA256Digest         `json:"capabilities_digest"`
	ReceiptID             core.ReceiptID            `json:"receipt_id"`
	StartedAt             time.Time                 `json:"started_at"`
	FinishedAt            time.Time                 `json:"finished_at"`
}

// Validate enforces the exact required fields and terminal-state-dependent
// fields of the V1 RunReport contract.
func (value RunReport) Validate() error {
	if err := value.RequestID.Validate(); err != nil {
		return fmt.Errorf("run report request ID: %w", err)
	}
	if err := value.RunID.Validate(); err != nil {
		return fmt.Errorf("run report run ID: %w", err)
	}
	if err := value.Status.Validate(); err != nil {
		return fmt.Errorf("run report status: %w", err)
	}
	if value.Validations == nil {
		return errors.New("run report validations are required")
	}
	for index, validation := range value.Validations {
		if err := validation.Validate(); err != nil {
			return fmt.Errorf("run report validation %d: %w", index, err)
		}
	}
	if value.Findings == nil {
		return errors.New("run report findings are required")
	}
	if value.ExceptionSummary != nil {
		if err := value.ExceptionSummary.Validate(); err != nil {
			return fmt.Errorf("run report exception summary: %w", err)
		}
	}
	if value.SourceSnapshotHash != nil {
		if err := value.SourceSnapshotHash.Validate(); err != nil {
			return fmt.Errorf("run report source snapshot hash: %w", err)
		}
	}
	if err := value.EffectiveConfigDigest.Validate(); err != nil {
		return fmt.Errorf("run report effective config digest: %w", err)
	}
	if err := value.CapabilitiesDigest.Validate(); err != nil {
		return fmt.Errorf("run report capabilities digest: %w", err)
	}
	if err := value.ReceiptID.Validate(); err != nil {
		return fmt.Errorf("run report receipt ID: %w", err)
	}
	if err := validateRunReportTimestamp("started_at", value.StartedAt); err != nil {
		return err
	}
	if err := validateRunReportTimestamp("finished_at", value.FinishedAt); err != nil {
		return err
	}

	if value.Status == core.TerminalStatusSucceeded {
		if value.Output == nil {
			return errors.New("succeeded run report requires output")
		}
		if value.FirstFailure != nil {
			return errors.New("succeeded run report must not contain first failure")
		}
		if err := value.Output.Validate(); err != nil {
			return fmt.Errorf("run report output: %w", err)
		}
		return nil
	}

	if value.Output != nil {
		return errors.New("non-succeeded run report must not contain output")
	}
	if value.FirstFailure == nil {
		return errors.New("non-succeeded run report requires first failure")
	}
	if err := value.FirstFailure.Validate(); err != nil {
		return fmt.Errorf("run report first failure: %w", err)
	}
	return nil
}

func validateRunReportTimestamp(name string, value time.Time) error {
	if value.IsZero() || value.Location() != time.UTC {
		return fmt.Errorf("run report %s must be a UTC RFC3339Nano timestamp", name)
	}
	return nil
}
