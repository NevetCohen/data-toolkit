// Package core defines stable identities shared by canonical V1 data packages.
package core

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// SchemaVersion identifies the version of a serialized V1 contract.
type SchemaVersion string

// RequestID correlates an API request.
type RequestID string

// RunID identifies one execution attempt.
type RunID string

// ReceiptID identifies a terminal receipt.
type ReceiptID string

// DataTypeID identifies a registered canonical data type.
type DataTypeID string

// FormatID identifies a V1 runtime format.
type FormatID string

// OperationID identifies a registered operation.
type OperationID string

// SourceID identifies a source within one request.
type SourceID string

// SheetID identifies the selected sheet of a source table.
type SheetID string

// TableID identifies an inspected source table.
type TableID string

// ColumnID identifies a column within one inspected source table.
type ColumnID string

// RowID identifies a row within one snapshot.
type RowID string

// SHA256Digest is a lowercase hexadecimal SHA-256 digest.
type SHA256Digest string

const (
	V1 SchemaVersion = "v1"

	CSVFormat   FormatID = "csv"
	JSONFormat  FormatID = "json"
	ExcelFormat FormatID = "excel"
)

// TypeAuthority identifies whether a column type was declared or inferred.
type TypeAuthority string

const (
	TypeAuthorityDeclared TypeAuthority = "declared"
	TypeAuthorityInferred TypeAuthority = "inferred"
)

// FindingSeverity classifies the significance of a finding.
type FindingSeverity string

const (
	FindingSeverityInfo    FindingSeverity = "info"
	FindingSeverityWarning FindingSeverity = "warning"
	FindingSeverityError   FindingSeverity = "error"
)

// TerminalStatus identifies the terminal state of a run.
type TerminalStatus string

const (
	TerminalStatusSucceeded TerminalStatus = "succeeded"
	TerminalStatusFailed    TerminalStatus = "failed"
	TerminalStatusCancelled TerminalStatus = "cancelled"
	TerminalStatusBlocked   TerminalStatus = "blocked"
)

// OperationOwner identifies the engine that owns an operation.
type OperationOwner string

const (
	OperationOwnerReaderEngine  OperationOwner = "reader_engine"
	OperationOwnerLogicalEngine OperationOwner = "logical_engine"
	OperationOwnerWriterEngine  OperationOwner = "writer_engine"
)

// OperationExposure identifies how an operation can be invoked.
type OperationExposure string

const (
	OperationExposureAPI       OperationExposure = "api"
	OperationExposureWorkflow  OperationExposure = "workflow"
	OperationExposureLifecycle OperationExposure = "lifecycle"
)

// StreamingMode identifies the resource contract for an operation.
type StreamingMode string

const (
	StreamingModeStreaming    StreamingMode = "streaming"
	StreamingModeBoundedSpool StreamingMode = "bounded_spool"
)

// CollisionPolicy identifies the local output collision behavior.
type CollisionPolicy string

const (
	CollisionPolicyBlock         CollisionPolicy = "block"
	CollisionPolicyOverwrite     CollisionPolicy = "overwrite"
	CollisionPolicyAlternateName CollisionPolicy = "alternate_name"
)

// RemoteCollisionPolicy identifies the Google Sheets bridge collision behavior.
type RemoteCollisionPolicy string

const (
	RemoteCollisionPolicyBlock         RemoteCollisionPolicy = "block"
	RemoteCollisionPolicyAlternateName RemoteCollisionPolicy = "alternate_name"
)

// ExceptionAction identifies how a configured exception is handled.
type ExceptionAction string

const (
	ExceptionActionBlock  ExceptionAction = "block"
	ExceptionActionReport ExceptionAction = "report"
	ExceptionActionIgnore ExceptionAction = "ignore"
)

// ExceptionDetail identifies the requested exception reporting detail.
type ExceptionDetail string

const (
	ExceptionDetailSummary ExceptionDetail = "summary"
	ExceptionDetailFull    ExceptionDetail = "full"
)

func (version SchemaVersion) Validate() error {
	if version != V1 {
		return fmt.Errorf("schema version %q is invalid; want %q", version, V1)
	}
	return nil
}

func (id RequestID) Validate() error { return validateUUID("request ID", string(id)) }

func (id RunID) Validate() error { return validateUUID("run ID", string(id)) }

func (id ReceiptID) Validate() error { return validateDigest("receipt ID", string(id)) }

func (id DataTypeID) Validate() error {
	return validateLowercaseDottedIdentifier("data type ID", string(id))
}

func (id FormatID) Validate() error {
	switch id {
	case CSVFormat, JSONFormat, ExcelFormat:
		return nil
	default:
		return fmt.Errorf("format ID %q is not supported in V1", id)
	}
}

func (id OperationID) Validate() error {
	return validateLowercaseDottedIdentifier("operation ID", string(id))
}

func (id SourceID) Validate() error { return validateStableIdentifier("source ID", string(id)) }

func (id SheetID) Validate() error { return validateStableIdentifier("sheet ID", string(id)) }

func (id TableID) Validate() error { return validateStableIdentifier("table ID", string(id)) }

func (id ColumnID) Validate() error { return validateStableIdentifier("column ID", string(id)) }

func (id RowID) Validate() error { return validateStableIdentifier("row ID", string(id)) }

func (digest SHA256Digest) Validate() error {
	return validateDigest("SHA-256 digest", string(digest))
}

func (authority TypeAuthority) Validate() error {
	return validateEnum("type authority", string(authority), TypeAuthorityDeclared, TypeAuthorityInferred)
}

func (severity FindingSeverity) Validate() error {
	return validateEnum("finding severity", string(severity), FindingSeverityInfo, FindingSeverityWarning, FindingSeverityError)
}

func (status TerminalStatus) Validate() error {
	return validateEnum("terminal status", string(status), TerminalStatusSucceeded, TerminalStatusFailed, TerminalStatusCancelled, TerminalStatusBlocked)
}

func (owner OperationOwner) Validate() error {
	return validateEnum("operation owner", string(owner), OperationOwnerReaderEngine, OperationOwnerLogicalEngine, OperationOwnerWriterEngine)
}

func (exposure OperationExposure) Validate() error {
	return validateEnum("operation exposure", string(exposure), OperationExposureAPI, OperationExposureWorkflow, OperationExposureLifecycle)
}

func (mode StreamingMode) Validate() error {
	return validateEnum("streaming mode", string(mode), StreamingModeStreaming, StreamingModeBoundedSpool)
}

func (policy CollisionPolicy) Validate() error {
	return validateEnum("collision policy", string(policy), CollisionPolicyBlock, CollisionPolicyOverwrite, CollisionPolicyAlternateName)
}

func (policy RemoteCollisionPolicy) Validate() error {
	return validateEnum("remote collision policy", string(policy), RemoteCollisionPolicyBlock, RemoteCollisionPolicyAlternateName)
}

func (action ExceptionAction) Validate() error {
	return validateEnum("exception action", string(action), ExceptionActionBlock, ExceptionActionReport, ExceptionActionIgnore)
}

func (detail ExceptionDetail) Validate() error {
	return validateEnum("exception detail", string(detail), ExceptionDetailSummary, ExceptionDetailFull)
}

func validateUUID(name, value string) error {
	if !isValidString(value) || len(value) != 36 {
		return fmt.Errorf("%s must be a non-empty UUID", name)
	}
	for index, character := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			if character != '-' {
				return fmt.Errorf("%s must be a UUID", name)
			}
			continue
		}
		if !isHex(character) {
			return fmt.Errorf("%s must be a UUID", name)
		}
	}
	return nil
}

func validateDigest(name, value string) error {
	if !isValidString(value) || len(value) != 64 {
		return fmt.Errorf("%s must be 64 lowercase hexadecimal characters", name)
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return fmt.Errorf("%s must be 64 lowercase hexadecimal characters", name)
		}
	}
	return nil
}

func validateLowercaseDottedIdentifier(name, value string) error {
	if !isValidString(value) {
		return fmt.Errorf("%s must be a lowercase dotted identifier", name)
	}
	for _, segment := range strings.Split(value, ".") {
		if segment == "" {
			return fmt.Errorf("%s must be a lowercase dotted identifier", name)
		}
		for _, character := range segment {
			if !((character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '_' || character == '-') {
				return fmt.Errorf("%s must be a lowercase dotted identifier", name)
			}
		}
	}
	return nil
}

func validateStableIdentifier(name, value string) error {
	if !isValidString(value) || strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required and must be UTF-8 NFC", name)
	}
	return nil
}

func validateEnum[T ~string](name, value string, allowed ...T) error {
	for _, candidate := range allowed {
		if value == string(candidate) {
			return nil
		}
	}
	return fmt.Errorf("%s %q is invalid", name, value)
}

func isValidString(value string) bool {
	return utf8.ValidString(value) && norm.NFC.IsNormalString(value)
}

func isHex(character rune) bool {
	return (character >= '0' && character <= '9') ||
		(character >= 'a' && character <= 'f') ||
		(character >= 'A' && character <= 'F')
}
