// Package config loads, validates, and resolves non-secret Data Toolkit V1
// settings. Credentials and compiled extension registrations are not config.
package config

const CurrentVersion = "v1"

type Document struct {
	SchemaVersion string         `json:"schema_version" yaml:"schema_version"`
	Display       Display        `json:"display" yaml:"display"`
	Aliases       []Alias        `json:"aliases" yaml:"aliases"`
	SemanticTypes []SemanticType `json:"semantic_types" yaml:"semantic_types"`
	Styles        []Style        `json:"styles" yaml:"styles"`
	Output        Output         `json:"output" yaml:"output"`
	Cleaning      Cleaning       `json:"cleaning" yaml:"cleaning"`
	Exceptions    Exceptions     `json:"exceptions" yaml:"exceptions"`
	Boolean       Boolean        `json:"boolean" yaml:"boolean"`
	Runtime       Runtime        `json:"runtime" yaml:"runtime"`
	Interfaces    Interfaces     `json:"interfaces" yaml:"interfaces"`
}

type Display struct {
	NullDisplay          string `json:"null_display" yaml:"null_display"`
	DecimalPrecision     int    `json:"decimal_precision" yaml:"decimal_precision"`
	DateFormat           string `json:"date_format" yaml:"date_format"`
	TimeFormat           string `json:"time_format" yaml:"time_format"`
	PhoneFormat          string `json:"phone_format" yaml:"phone_format"`
	TextEncoding         string `json:"text_encoding" yaml:"text_encoding"`
	UnicodeNormalization string `json:"unicode_normalization" yaml:"unicode_normalization"`
}

type Alias struct {
	Name         string `json:"name" yaml:"name"`
	SemanticType string `json:"semantic_type" yaml:"semantic_type"`
}

type SemanticType struct {
	Name          string `json:"name" yaml:"name"`
	DataType      string `json:"data_type" yaml:"data_type"`
	DefaultFormat string `json:"default_format,omitempty" yaml:"default_format,omitempty"`
}

type Style struct {
	Name               string `json:"name" yaml:"name"`
	HeaderFill         string `json:"header_fill,omitempty" yaml:"header_fill,omitempty"`
	HeaderText         string `json:"header_text,omitempty" yaml:"header_text,omitempty"`
	AlternatingRowFill string `json:"alternating_row_fill,omitempty" yaml:"alternating_row_fill,omitempty"`
	BoldHeaders        bool   `json:"bold_headers" yaml:"bold_headers"`
	RightToLeft        bool   `json:"right_to_left" yaml:"right_to_left"`
}

type Output struct {
	Directory        string `json:"directory" yaml:"directory"`
	FilenameTemplate string `json:"filename_template" yaml:"filename_template"`
	DefaultStyle     string `json:"default_style" yaml:"default_style"`
	TXTDelimiter     string `json:"txt_delimiter" yaml:"txt_delimiter"`
	Collision        string `json:"collision" yaml:"collision"`
}

type Cleaning struct {
	MixedTypeAction        string `json:"mixed_type_action" yaml:"mixed_type_action"`
	DeduplicateKeep        string `json:"deduplicate_keep" yaml:"deduplicate_keep"`
	AutoNormalizeFormats   bool   `json:"auto_normalize_formats" yaml:"auto_normalize_formats"`
	TrimDetectedWhitespace bool   `json:"trim_detected_whitespace" yaml:"trim_detected_whitespace"`
}

type Exceptions struct {
	Action string `json:"action" yaml:"action"`
	Detail string `json:"detail" yaml:"detail"`
}

type Boolean struct {
	NotScope string `json:"not_scope" yaml:"not_scope"`
}

type Runtime struct {
	MaximumMemoryBytes int64  `json:"maximum_memory_bytes" yaml:"maximum_memory_bytes"`
	LockRetryCount     int    `json:"lock_retry_count" yaml:"lock_retry_count"`
	LockRetryInterval  string `json:"lock_retry_interval" yaml:"lock_retry_interval"`
	LockTimeout        string `json:"lock_timeout" yaml:"lock_timeout"`
	TemporaryWorkspace string `json:"temporary_workspace" yaml:"temporary_workspace"`
	FailedRunRetention string `json:"failed_run_retention" yaml:"failed_run_retention"`
}

type Interfaces struct {
	RecentWorkflows int `json:"recent_workflows" yaml:"recent_workflows"`
}

// Patch is a programmatic precedence layer. Pointer fields distinguish omitted
// settings from explicit zero or false values.
type Patch struct {
	Display       *DisplayPatch
	Aliases       *[]Alias
	SemanticTypes *[]SemanticType
	Styles        *[]Style
	Output        *OutputPatch
	Cleaning      *CleaningPatch
	Exceptions    *ExceptionsPatch
	Boolean       *BooleanPatch
	Runtime       *RuntimePatch
	Interfaces    *InterfacesPatch
}

type DisplayPatch struct {
	NullDisplay          *string
	DecimalPrecision     *int
	DateFormat           *string
	TimeFormat           *string
	PhoneFormat          *string
	TextEncoding         *string
	UnicodeNormalization *string
}

type OutputPatch struct {
	Directory        *string
	FilenameTemplate *string
	DefaultStyle     *string
	TXTDelimiter     *string
	Collision        *string
}

type CleaningPatch struct {
	MixedTypeAction        *string
	DeduplicateKeep        *string
	AutoNormalizeFormats   *bool
	TrimDetectedWhitespace *bool
}

type ExceptionsPatch struct {
	Action *string
	Detail *string
}

type BooleanPatch struct {
	NotScope *string
}

type RuntimePatch struct {
	MaximumMemoryBytes *int64
	LockRetryCount     *int
	LockRetryInterval  *string
	LockTimeout        *string
	TemporaryWorkspace *string
	FailedRunRetention *string
}

type InterfacesPatch struct {
	RecentWorkflows *int
}
