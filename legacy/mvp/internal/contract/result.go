package contract

import "time"

type RunStatus string

const (
	RunStatusStarted   RunStatus = "started"
	RunStatusSucceeded RunStatus = "succeeded"
	RunStatusFailed    RunStatus = "failed"
)

type StartedResult struct {
	SchemaVersion string    `json:"schema_version" yaml:"schema_version"`
	RunID         string    `json:"run_id" yaml:"run_id"`
	WorkflowID    string    `json:"workflow_id" yaml:"workflow_id"`
	Status        RunStatus `json:"status" yaml:"status"`
	StartedAt     time.Time `json:"started_at" yaml:"started_at"`
}

type FinalResult struct {
	SchemaVersion    string             `json:"schema_version" yaml:"schema_version"`
	RunID            string             `json:"run_id" yaml:"run_id"`
	WorkflowID       string             `json:"workflow_id" yaml:"workflow_id"`
	Status           RunStatus          `json:"status" yaml:"status"`
	StartedAt        time.Time          `json:"started_at" yaml:"started_at"`
	FinishedAt       time.Time          `json:"finished_at" yaml:"finished_at"`
	Outputs          []ProducedOutput   `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	QueryAnswers     []QueryAnswer      `json:"query_answers,omitempty" yaml:"query_answers,omitempty"`
	ResolvedSettings map[string]any     `json:"resolved_settings" yaml:"resolved_settings"`
	Validations      []ValidationResult `json:"validations,omitempty" yaml:"validations,omitempty"`
	Warnings         []Diagnostic       `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	Exceptions       []ExceptionRecord  `json:"exceptions,omitempty" yaml:"exceptions,omitempty"`
	Errors           []Diagnostic       `json:"errors,omitempty" yaml:"errors,omitempty"`
}

type ProducedOutput struct {
	ID       string     `json:"id" yaml:"id"`
	Format   DataFormat `json:"format" yaml:"format"`
	Location string     `json:"location" yaml:"location"`
	Rows     int64      `json:"rows" yaml:"rows"`
}

type QueryAnswer struct {
	ID    string     `json:"id" yaml:"id"`
	Shape QueryShape `json:"shape" yaml:"shape"`
	Value any        `json:"value" yaml:"value"`
}

type ValidationResult struct {
	ID      string `json:"id" yaml:"id"`
	Passed  bool   `json:"passed" yaml:"passed"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

type Diagnostic struct {
	Code    string         `json:"code" yaml:"code"`
	Path    string         `json:"path,omitempty" yaml:"path,omitempty"`
	Message string         `json:"message" yaml:"message"`
	Details map[string]any `json:"details,omitempty" yaml:"details,omitempty"`
}

type ExceptionRecord struct {
	Reference      string         `json:"reference" yaml:"reference"`
	Sequence       uint64         `json:"sequence" yaml:"sequence"`
	SourceOrdinal  uint64         `json:"source_ordinal" yaml:"source_ordinal"`
	SourceID       string         `json:"source_id,omitempty" yaml:"source_id,omitempty"`
	SheetID        string         `json:"sheet_id,omitempty" yaml:"sheet_id,omitempty"`
	RowOrdinal     uint64         `json:"row_ordinal" yaml:"row_ordinal"`
	RowID          string         `json:"row_id,omitempty" yaml:"row_id,omitempty"`
	ColumnOrdinal  int            `json:"column_ordinal" yaml:"column_ordinal"`
	ColumnID       string         `json:"column_id,omitempty" yaml:"column_id,omitempty"`
	ObservedClass  string         `json:"observed_class,omitempty" yaml:"observed_class,omitempty"`
	ValueReference string         `json:"value_reference,omitempty" yaml:"value_reference,omitempty"`
	Reason         string         `json:"reason,omitempty" yaml:"reason,omitempty"`
	Code           string         `json:"code" yaml:"code"`
	Details        map[string]any `json:"details,omitempty" yaml:"details,omitempty"`
}
