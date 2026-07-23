// Package config loads, validates, and resolves non-secret Data Toolkit
// settings. Credentials are deliberately outside these types.
package config

import (
	"time"

	"data-toolkit/internal/contract"
)

const CurrentSchemaVersion = "v1"

// Config is a partial user configuration. Pointer fields distinguish an
// omitted setting from an explicit false, zero, or empty value.
type Config struct {
	SchemaVersion string         `json:"schema_version" yaml:"schema_version"`
	Display       DisplayConfig  `json:"display,omitempty" yaml:"display,omitempty"`
	Aliases       []ColumnAlias  `json:"aliases,omitempty" yaml:"aliases,omitempty"`
	Styles        []TableStyle   `json:"styles,omitempty" yaml:"styles,omitempty"`
	Output        OutputConfig   `json:"output,omitempty" yaml:"output,omitempty"`
	Cleaning      CleaningConfig `json:"cleaning,omitempty" yaml:"cleaning,omitempty"`
	Runtime       RuntimeConfig  `json:"runtime,omitempty" yaml:"runtime,omitempty"`
}

type DisplayConfig struct {
	NullDisplay      *string `json:"null_display,omitempty" yaml:"null_display,omitempty"`
	DecimalPrecision *int    `json:"decimal_precision,omitempty" yaml:"decimal_precision,omitempty"`
	DateFormat       *string `json:"date_format,omitempty" yaml:"date_format,omitempty"`
	TimeFormat       *string `json:"time_format,omitempty" yaml:"time_format,omitempty"`
	PhoneFormat      *string `json:"phone_format,omitempty" yaml:"phone_format,omitempty"`
	TextEncoding     *string `json:"text_encoding,omitempty" yaml:"text_encoding,omitempty"`
}

type ColumnAlias struct {
	Name         string `json:"name" yaml:"name"`
	SemanticType string `json:"semantic_type" yaml:"semantic_type"`
}

// TableStyle is a named adapter-neutral style. Format adapters translate the
// declared colors and emphasis to their native representation.
type TableStyle struct {
	Name               string `json:"name" yaml:"name"`
	HeaderFill         string `json:"header_fill,omitempty" yaml:"header_fill,omitempty"`
	HeaderText         string `json:"header_text,omitempty" yaml:"header_text,omitempty"`
	AlternatingRowFill string `json:"alternating_row_fill,omitempty" yaml:"alternating_row_fill,omitempty"`
	BoldHeaders        *bool  `json:"bold_headers,omitempty" yaml:"bold_headers,omitempty"`
}

type OutputConfig struct {
	Directory        *string                   `json:"directory,omitempty" yaml:"directory,omitempty"`
	FilenameTemplate *string                   `json:"filename_template,omitempty" yaml:"filename_template,omitempty"`
	DefaultStyle     *string                   `json:"default_style,omitempty" yaml:"default_style,omitempty"`
	TXTDelimiter     *string                   `json:"txt_delimiter,omitempty" yaml:"txt_delimiter,omitempty"`
	Collision        *contract.CollisionAction `json:"collision,omitempty" yaml:"collision,omitempty"`
}

type CleaningConfig struct {
	MixedTypeAction        *contract.MixedTypeAction `json:"mixed_type_action,omitempty" yaml:"mixed_type_action,omitempty"`
	DeduplicateKeep        *contract.KeepPolicy      `json:"deduplicate_keep,omitempty" yaml:"deduplicate_keep,omitempty"`
	ExceptionDetail        *string                   `json:"exception_detail,omitempty" yaml:"exception_detail,omitempty"`
	AutoNormalizeFormats   *bool                     `json:"auto_normalize_formats,omitempty" yaml:"auto_normalize_formats,omitempty"`
	TrimDetectedWhitespace *bool                     `json:"trim_detected_whitespace,omitempty" yaml:"trim_detected_whitespace,omitempty"`
	NotScope               *contract.NotScope        `json:"not_scope,omitempty" yaml:"not_scope,omitempty"`
}

type RuntimeConfig struct {
	MaximumMemoryBytes *int64  `json:"maximum_memory_bytes,omitempty" yaml:"maximum_memory_bytes,omitempty"`
	LockRetryCount     *int    `json:"lock_retry_count,omitempty" yaml:"lock_retry_count,omitempty"`
	LockRetryInterval  *string `json:"lock_retry_interval,omitempty" yaml:"lock_retry_interval,omitempty"`
	LockTimeout        *string `json:"lock_timeout,omitempty" yaml:"lock_timeout,omitempty"`
	TemporaryWorkspace *string `json:"temporary_workspace,omitempty" yaml:"temporary_workspace,omitempty"`
	FailedRunRetention *string `json:"failed_run_retention,omitempty" yaml:"failed_run_retention,omitempty"`
	RecentWorkflows    *int    `json:"recent_workflows,omitempty" yaml:"recent_workflows,omitempty"`
}

// Resolved contains concrete effective values after all precedence layers.
type Resolved struct {
	SchemaVersion          string
	NullDisplay            string
	DecimalPrecision       int
	DateFormat             string
	TimeFormat             string
	PhoneFormat            string
	TextEncoding           string
	Aliases                []ColumnAlias
	Styles                 []TableStyle
	OutputDirectory        string
	FilenameTemplate       string
	TableStyle             string
	TXTDelimiter           string
	Collision              contract.CollisionAction
	MixedTypeAction        contract.MixedTypeAction
	DeduplicateKeep        contract.KeepPolicy
	ExceptionDetail        string
	AutoNormalizeFormats   bool
	TrimDetectedWhitespace bool
	NotScope               contract.NotScope
	MaximumMemoryBytes     int64
	LockRetryCount         int
	LockRetryInterval      time.Duration
	LockTimeout            time.Duration
	TemporaryWorkspace     string
	FailedRunRetention     time.Duration
	RecentWorkflows        int
	OutputFormat           contract.DataFormat
}
