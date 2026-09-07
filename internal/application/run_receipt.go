package application

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"data-toolkit/internal/core"
	"data-toolkit/internal/report"
)

// RunReceipt is the compact append-only terminal record for one attempted run.
// It intentionally excludes reports, findings, exceptions, and lifecycle events.
type RunReceipt struct {
	ReceiptID                core.ReceiptID          `json:"receipt_id"`
	RequestID                core.RequestID          `json:"request_id"`
	RunID                    core.RunID              `json:"run_id"`
	NormalizedWorkflowDigest core.SHA256Digest       `json:"normalized_workflow_digest"`
	BinaryVersion            string                  `json:"binary_version"`
	CapabilitiesDigest       core.SHA256Digest       `json:"capabilities_digest"`
	EffectiveConfigDigest    core.SHA256Digest       `json:"effective_config_digest"`
	SourceSnapshotHash       *core.SHA256Digest      `json:"source_snapshot_hash,omitempty"`
	PublishedOutput          *report.PublishedOutput `json:"published_output,omitempty"`
	Status                   core.TerminalStatus     `json:"status"`
	TerminalError            *APIError               `json:"terminal_error,omitempty"`
	StartedAt                time.Time               `json:"started_at"`
	FinishedAt               time.Time               `json:"finished_at"`
}

// Validate enforces the required fields and terminal-state-dependent fields of
// the V1 compact receipt contract.
func (value RunReceipt) Validate() error {
	if err := value.ReceiptID.Validate(); err != nil {
		return fmt.Errorf("run receipt receipt ID: %w", err)
	}
	if err := value.RequestID.Validate(); err != nil {
		return fmt.Errorf("run receipt request ID: %w", err)
	}
	if err := value.RunID.Validate(); err != nil {
		return fmt.Errorf("run receipt run ID: %w", err)
	}
	if err := value.NormalizedWorkflowDigest.Validate(); err != nil {
		return fmt.Errorf("run receipt normalized workflow digest: %w", err)
	}
	if strings.TrimSpace(value.BinaryVersion) == "" {
		return errors.New("run receipt binary version is required")
	}
	if err := value.CapabilitiesDigest.Validate(); err != nil {
		return fmt.Errorf("run receipt capabilities digest: %w", err)
	}
	if err := value.EffectiveConfigDigest.Validate(); err != nil {
		return fmt.Errorf("run receipt effective config digest: %w", err)
	}
	if value.SourceSnapshotHash != nil {
		if err := value.SourceSnapshotHash.Validate(); err != nil {
			return fmt.Errorf("run receipt source snapshot hash: %w", err)
		}
	}
	if err := value.Status.Validate(); err != nil {
		return fmt.Errorf("run receipt status: %w", err)
	}
	if err := validateRunReceiptTimestamp("started_at", value.StartedAt); err != nil {
		return err
	}
	if err := validateRunReceiptTimestamp("finished_at", value.FinishedAt); err != nil {
		return err
	}

	if value.Status == core.TerminalStatusSucceeded {
		if value.PublishedOutput == nil {
			return errors.New("succeeded run receipt requires published output")
		}
		if value.TerminalError != nil {
			return errors.New("succeeded run receipt must not contain terminal error")
		}
		if err := value.PublishedOutput.Validate(); err != nil {
			return fmt.Errorf("run receipt published output: %w", err)
		}
		return nil
	}

	if value.PublishedOutput != nil {
		return errors.New("non-succeeded run receipt must not contain published output")
	}
	if value.TerminalError == nil {
		return errors.New("non-succeeded run receipt requires terminal error")
	}
	if err := value.TerminalError.Validate(); err != nil {
		return fmt.Errorf("run receipt terminal error: %w", err)
	}
	return nil
}

func validateRunReceiptTimestamp(name string, value time.Time) error {
	if value.IsZero() || value.Location() != time.UTC {
		return fmt.Errorf("run receipt %s must be a UTC RFC3339Nano timestamp", name)
	}
	return nil
}
