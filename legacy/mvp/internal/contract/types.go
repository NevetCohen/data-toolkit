// Package contract defines the versioned, interface-neutral workflow and result
// documents used by every Data Toolkit entry point.
package contract

const CurrentSchemaVersion = "v1"

// RunRequest is the single application-service envelope used by CLI, TUI, and
// Codex integrations. Workflow remains the versioned executable contract.
type RunRequest struct {
	Workflow     Workflow `json:"workflow" yaml:"workflow"`
	ValidateOnly bool     `json:"validate_only,omitempty" yaml:"validate_only,omitempty"`
}

// Workflow is the complete normalized request accepted by the orchestrator.
type Workflow struct {
	SchemaVersion string            `json:"schema_version" yaml:"schema_version"`
	ID            string            `json:"id" yaml:"id"`
	Sources       []Source          `json:"sources" yaml:"sources"`
	Mappings      []FieldMapping    `json:"mappings,omitempty" yaml:"mappings,omitempty"`
	Operations    []Operation       `json:"operations" yaml:"operations"`
	Outputs       []Output          `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Queries       []Query           `json:"queries,omitempty" yaml:"queries,omitempty"`
	Overrides     WorkflowOverrides `json:"overrides,omitempty" yaml:"overrides,omitempty"`
	Interface     InterfaceMetadata `json:"interface" yaml:"interface"`
	Exception     ExceptionPolicy   `json:"exception" yaml:"exception"`
	Validations   []Validation      `json:"validations,omitempty" yaml:"validations,omitempty"`
}

type Source struct {
	ID       string           `json:"id" yaml:"id"`
	Format   DataFormat       `json:"format" yaml:"format"`
	Location string           `json:"location" yaml:"location"`
	Sheets   []SheetSelection `json:"sheets,omitempty" yaml:"sheets,omitempty"`
	Options  SourceOptions    `json:"options,omitempty" yaml:"options,omitempty"`
}

type DataFormat string

const (
	FormatJSON         DataFormat = "json"
	FormatCSV          DataFormat = "csv"
	FormatExcel        DataFormat = "excel"
	FormatTXT          DataFormat = "txt"
	FormatGoogleSheets DataFormat = "google_sheets"
)

type SheetSelection struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	Index *int   `json:"index,omitempty" yaml:"index,omitempty"`
}

type SourceOptions struct {
	Encoding       string `json:"encoding,omitempty" yaml:"encoding,omitempty"`
	Delimiter      string `json:"delimiter,omitempty" yaml:"delimiter,omitempty"`
	DisplayedValue bool   `json:"displayed_values,omitempty" yaml:"displayed_values,omitempty"`
	JSONRoot       string `json:"json_root,omitempty" yaml:"json_root,omitempty"`
}

type FieldMapping struct {
	SourceID string    `json:"source_id" yaml:"source_id"`
	Sheet    string    `json:"sheet,omitempty" yaml:"sheet,omitempty"`
	Column   ColumnRef `json:"column" yaml:"column"`
	Alias    string    `json:"alias,omitempty" yaml:"alias,omitempty"`
	JSONPath string    `json:"json_path,omitempty" yaml:"json_path,omitempty"`
	Type     string    `json:"type,omitempty" yaml:"type,omitempty"`
}

// ColumnRef addresses exactly one physical position, original header, or
// configured global alias.
type ColumnRef struct {
	Position string `json:"position,omitempty" yaml:"position,omitempty"`
	Header   string `json:"header,omitempty" yaml:"header,omitempty"`
	Alias    string `json:"alias,omitempty" yaml:"alias,omitempty"`
}

type Operation struct {
	ID          string              `json:"id" yaml:"id"`
	Kind        OperationKind       `json:"kind" yaml:"kind"`
	Inputs      []string            `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Condition   *Expression         `json:"condition,omitempty" yaml:"condition,omitempty"`
	Rename      []RenameField       `json:"rename,omitempty" yaml:"rename,omitempty"`
	Columns     []ColumnRef         `json:"columns,omitempty" yaml:"columns,omitempty"`
	Normalize   *NormalizeOptions   `json:"normalize,omitempty" yaml:"normalize,omitempty"`
	Deduplicate *DeduplicateOptions `json:"deduplicate,omitempty" yaml:"deduplicate,omitempty"`
	Sort        []SortField         `json:"sort,omitempty" yaml:"sort,omitempty"`
	Split       *SplitOptions       `json:"split,omitempty" yaml:"split,omitempty"`
	Text        *TextOptions        `json:"text,omitempty" yaml:"text,omitempty"`
	Regex       *RegexOptions       `json:"regex,omitempty" yaml:"regex,omitempty"`
	Overrides   OperationOverrides  `json:"overrides,omitempty" yaml:"overrides,omitempty"`
}

type OperationKind string

const (
	OperationFilter          OperationKind = "filter"
	OperationRename          OperationKind = "rename"
	OperationNormalize       OperationKind = "normalize"
	OperationTrim            OperationKind = "trim"
	OperationDeduplicate     OperationKind = "deduplicate"
	OperationDeleteEmptyRows OperationKind = "delete_empty_rows"
	OperationDeleteRows      OperationKind = "delete_rows"
	OperationSort            OperationKind = "sort"
	OperationSplit           OperationKind = "split"
	OperationConcatenate     OperationKind = "concatenate"
	OperationTransformText   OperationKind = "transform_text"
	OperationRegexTransform  OperationKind = "regex_transform"
)

type RenameField struct {
	Column ColumnRef `json:"column" yaml:"column"`
	To     string    `json:"to" yaml:"to"`
}

type NormalizeOptions struct {
	SemanticType string `json:"semantic_type" yaml:"semantic_type"`
	Format       string `json:"format,omitempty" yaml:"format,omitempty"`
}

type DeduplicateOptions struct {
	Keys []ColumnRef `json:"keys,omitempty" yaml:"keys,omitempty"`
	Keep KeepPolicy  `json:"keep,omitempty" yaml:"keep,omitempty"`
}

type KeepPolicy string

const (
	KeepFirst KeepPolicy = "first"
	KeepLast  KeepPolicy = "last"
	KeepError KeepPolicy = "error"
)

type SortField struct {
	Column    ColumnRef     `json:"column" yaml:"column"`
	Direction SortDirection `json:"direction,omitempty" yaml:"direction,omitempty"`
}

type SortDirection string

const (
	SortAscending  SortDirection = "asc"
	SortDescending SortDirection = "desc"
)

type SplitOptions struct {
	Branches []SplitBranch `json:"branches" yaml:"branches"`
}

type SplitBranch struct {
	ID        string      `json:"id" yaml:"id"`
	Condition *Expression `json:"condition,omitempty" yaml:"condition,omitempty"`
}

type TextOptions struct {
	Target    ColumnRef   `json:"target" yaml:"target"`
	Sources   []ColumnRef `json:"sources,omitempty" yaml:"sources,omitempty"`
	Separator string      `json:"separator,omitempty" yaml:"separator,omitempty"`
	Transform string      `json:"transform,omitempty" yaml:"transform,omitempty"`
}

type RegexOptions struct {
	Column      ColumnRef `json:"column" yaml:"column"`
	Pattern     string    `json:"pattern" yaml:"pattern"`
	Replacement string    `json:"replacement,omitempty" yaml:"replacement,omitempty"`
}

type OperationOverrides struct {
	MixedTypeAction MixedTypeAction `json:"mixed_type_action,omitempty" yaml:"mixed_type_action,omitempty"`
}

type Output struct {
	ID               string            `json:"id" yaml:"id"`
	Input            string            `json:"input" yaml:"input"`
	Format           DataFormat        `json:"format" yaml:"format"`
	Location         string            `json:"location,omitempty" yaml:"location,omitempty"`
	LocationPolicy   *LocationPolicy   `json:"location_policy,omitempty" yaml:"location_policy,omitempty"`
	Projection       []ProjectionField `json:"projection" yaml:"projection"`
	BranchProjection *BranchProjection `json:"branch_projection,omitempty" yaml:"branch_projection,omitempty"`
	Style            string            `json:"style,omitempty" yaml:"style,omitempty"`
	Validations      []Validation      `json:"validations,omitempty" yaml:"validations,omitempty"`
	Overrides        OutputOverrides   `json:"overrides,omitempty" yaml:"overrides,omitempty"`
}

type LocationPolicy struct {
	Directory        string `json:"directory,omitempty" yaml:"directory,omitempty"`
	FilenameTemplate string `json:"filename_template,omitempty" yaml:"filename_template,omitempty"`
}

type ProjectionField struct {
	Column   ColumnRef `json:"column" yaml:"column"`
	OutputAs string    `json:"output_as" yaml:"output_as"`
}

type BranchProjection struct {
	Source1 []ProjectionField `json:"source_1" yaml:"source_1"`
	Source2 []ProjectionField `json:"source_2" yaml:"source_2"`
}

type OutputOverrides struct {
	Filename  string          `json:"filename,omitempty" yaml:"filename,omitempty"`
	Directory string          `json:"directory,omitempty" yaml:"directory,omitempty"`
	Format    DataFormat      `json:"format,omitempty" yaml:"format,omitempty"`
	Style     string          `json:"style,omitempty" yaml:"style,omitempty"`
	Delimiter string          `json:"delimiter,omitempty" yaml:"delimiter,omitempty"`
	Collision CollisionAction `json:"collision,omitempty" yaml:"collision,omitempty"`
}

type Query struct {
	ID         string            `json:"id" yaml:"id"`
	Input      string            `json:"input" yaml:"input"`
	Kind       QueryKind         `json:"kind" yaml:"kind"`
	Projection []ProjectionField `json:"projection,omitempty" yaml:"projection,omitempty"`
	Shape      QueryShape        `json:"shape" yaml:"shape"`
}

type QueryKind string

const (
	QueryRows           QueryKind = "rows"
	QueryCount          QueryKind = "count"
	QueryDistinctValues QueryKind = "distinct_values"
)

type QueryShape string

const (
	QueryShapeRows   QueryShape = "rows"
	QueryShapeScalar QueryShape = "scalar"
	QueryShapeValues QueryShape = "values"
)

type Validation struct {
	ID       string         `json:"id,omitempty" yaml:"id,omitempty"`
	Kind     ValidationKind `json:"kind" yaml:"kind"`
	Target   string         `json:"target,omitempty" yaml:"target,omitempty"`
	Columns  []ColumnRef    `json:"columns,omitempty" yaml:"columns,omitempty"`
	Expected string         `json:"expected,omitempty" yaml:"expected,omitempty"`
}

type ValidationKind string

const (
	ValidationFileExists     ValidationKind = "file_exists"
	ValidationRowCount       ValidationKind = "row_count"
	ValidationRequiredFields ValidationKind = "required_fields"
	ValidationUnique         ValidationKind = "unique"
	ValidationNonEmpty       ValidationKind = "non_empty"
)

type WorkflowOverrides struct {
	NullDisplay            string          `json:"null_display,omitempty" yaml:"null_display,omitempty"`
	DecimalPrecision       *int            `json:"decimal_precision,omitempty" yaml:"decimal_precision,omitempty"`
	DateFormat             string          `json:"date_format,omitempty" yaml:"date_format,omitempty"`
	TimeFormat             string          `json:"time_format,omitempty" yaml:"time_format,omitempty"`
	PhoneFormat            string          `json:"phone_format,omitempty" yaml:"phone_format,omitempty"`
	TextEncoding           string          `json:"text_encoding,omitempty" yaml:"text_encoding,omitempty"`
	OutputDirectory        string          `json:"output_directory,omitempty" yaml:"output_directory,omitempty"`
	FilenameTemplate       string          `json:"filename_template,omitempty" yaml:"filename_template,omitempty"`
	TableStyle             string          `json:"table_style,omitempty" yaml:"table_style,omitempty"`
	TXTDelimiter           string          `json:"txt_delimiter,omitempty" yaml:"txt_delimiter,omitempty"`
	MixedTypeAction        MixedTypeAction `json:"mixed_type_action,omitempty" yaml:"mixed_type_action,omitempty"`
	DeduplicateKeep        KeepPolicy      `json:"deduplicate_keep,omitempty" yaml:"deduplicate_keep,omitempty"`
	ExceptionDetail        string          `json:"exception_detail,omitempty" yaml:"exception_detail,omitempty"`
	AutoNormalizeFormats   *bool           `json:"auto_normalize_formats,omitempty" yaml:"auto_normalize_formats,omitempty"`
	TrimDetectedWhitespace *bool           `json:"trim_detected_whitespace,omitempty" yaml:"trim_detected_whitespace,omitempty"`
	NotScope               NotScope        `json:"not_scope,omitempty" yaml:"not_scope,omitempty"`
	Collision              CollisionAction `json:"collision,omitempty" yaml:"collision,omitempty"`
	MaximumMemory          int64           `json:"maximum_memory_bytes,omitempty" yaml:"maximum_memory_bytes,omitempty"`
}

type InterfaceMetadata struct {
	Name              string `json:"name" yaml:"name"`
	ApprovalBeforeRun bool   `json:"approval_before_run" yaml:"approval_before_run"`
}

type ExceptionPolicy struct {
	Action ExceptionAction `json:"action" yaml:"action"`
	Detail string          `json:"detail,omitempty" yaml:"detail,omitempty"`
}

type ExceptionAction string

const (
	ExceptionBlock  ExceptionAction = "block"
	ExceptionReport ExceptionAction = "report"
	ExceptionIgnore ExceptionAction = "ignore"
)

type MixedTypeAction string

const (
	MixedTypeBlock  MixedTypeAction = "block"
	MixedTypeReport MixedTypeAction = "report"
	MixedTypeIgnore MixedTypeAction = "ignore"
)

type CollisionAction string

const (
	CollisionBlock         CollisionAction = "block"
	CollisionOverwrite     CollisionAction = "overwrite"
	CollisionAlternateName CollisionAction = "alternate_name"
)
