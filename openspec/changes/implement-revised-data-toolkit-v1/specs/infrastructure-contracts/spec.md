## ADDED Requirements

### Requirement: Canonical streaming table model
The system SHALL represent transform data only as a validated `Schema`, an
ordered `RowStream`, and typed immutable `Value` instances. A cell SHALL contain
its column identity and value without per-cell provenance, lineage, or reverse
mutation state.

#### Scenario: Distinct null and text values
- **WHEN** a source contains null, an empty string, and the literal string `-`
- **THEN** the canonical stream preserves them as three distinct values

#### Scenario: Exact numeric value
- **WHEN** an adapter reads an integer or decimal that cannot round-trip through `float64`
- **THEN** the canonical value preserves the exact numeric representation

#### Scenario: Diagnostic location without cell provenance
- **WHEN** a value produces an inspection finding or execution error
- **THEN** the finding or error identifies the source location without adding provenance to every cell

### Requirement: Complete schema metadata
The system SHALL model stable column identity, physical position, source header,
optional subheader, alias, semantic type, data type, and type authority as
schema metadata, and SHALL keep row values outside column descriptors.

#### Scenario: Duplicate source headers
- **WHEN** inspection finds duplicate source headers
- **THEN** each column receives a distinct stable column ID and remains addressable without guessing by header text

### Requirement: Strict V1 configuration
The system SHALL accept only the revised V1 configuration groups and exact
fields defined by the product brief, SHALL reject unknown fields, and SHALL
enforce UTF-8, NFC, source-hash reporting, and an effective-configuration
digest in every run receipt.

#### Scenario: Removed foundation field
- **WHEN** configuration contains `output.txt_delimiter`, `boolean.not_scope`, or `interfaces.recent_workflows`
- **THEN** validation fails before workflow planning with a field-specific error

#### Scenario: Registered semantic type reference
- **WHEN** a semantic type refers to an unregistered data type
- **THEN** semantic validation rejects the configuration

### Requirement: Immutable configuration precedence
The system SHALL resolve `embedded defaults -> user config -> CLI config patches
-> workflow overrides -> output overrides` into a new typed `EffectiveConfig`
without mutating any input layer. The normalized resolved configuration SHALL
produce the effective-configuration digest used by validation, execution,
reports, and receipts.

#### Scenario: Output override wins
- **WHEN** all five layers provide a supported output setting
- **THEN** the effective configuration uses the output override and records its effective-configuration digest in the run receipt

### Requirement: Protected CLI configuration patches
Each CLI patch SHALL be one strict `ConfigOverride{path,value}` applied in
supplied order to a new configuration value. Duplicate paths, invalid JSON
Pointers, nonexistent paths, type-changing or otherwise invalid structural
changes, and any target equal to, above, or below a protected path SHALL fail
before planning. Full schema and semantic validation SHALL run after all CLI
patches and again after later workflow and output overrides.

`cli.non_overridable_config_paths` SHALL be sorted and unique and SHALL protect
at least `/schema_version`, `/cli/non_overridable_config_paths`,
`/report/receipt_store`, and `/runtime/temporary_workspace`. A CLI patch SHALL
NOT modify the protected-path list itself.

#### Scenario: Protected descendant patch
- **WHEN** a CLI patch targets a protected path, its ancestor, or its descendant
- **THEN** configuration validation returns `invalid_config`, leaves every input layer unchanged, and does not plan the run

### Requirement: Versioned application envelopes
The system SHALL expose typed versioned request and result envelopes for
capability discovery, description, configuration validation, inspection,
workflow validation, and execution, with a request ID and structured status.

#### Scenario: Request correlation
- **WHEN** an API operation returns success or failure
- **THEN** its result contains the originating request ID and supported schema version

### Requirement: Structured deterministic errors
The system SHALL return stable error codes, component, field or location when
available, and human-readable detail, without embedding secrets or input row
values unnecessarily.

#### Scenario: Unknown operation
- **WHEN** a workflow references an unregistered operation ID or version
- **THEN** validation returns a structured capability error before source rows are consumed

### Requirement: Compact append-only run receipt
The system SHALL append one terminal receipt per attempted run containing request
ID, normalized-workflow digest, capability and binary versions,
effective-configuration digest, source snapshot hash, published-output reference,
terminal status, one structured terminal error when present, and timestamps. A
receipt SHALL NOT contain per-cell lineage, a reversible change log, the full
request or configuration, ordinary input values, findings, exceptions, or a
persistent lifecycle-event stream.

#### Scenario: Failed run receipt
- **WHEN** execution fails after a snapshot is created
- **THEN** the receipt records the snapshot hash, failure, terminal error, and absence of a successfully published output

#### Scenario: Secret redaction
- **WHEN** a request or environment contains credential material
- **THEN** no credential value is written to the receipt or ordinary report

### Requirement: Primitive identities and enums
All contract objects SHALL use snake_case JSON names, `additionalProperties:
false`, UTF-8 strings normalized to NFC, and the following fixed aliases and
enums. An omitted optional field differs from an empty required field.

| Contract type | Underlying JSON/Go type | Validation |
|---|---|---|
| `SchemaVersion` | string | Constant `v1` |
| `RequestID` | string | UUID, non-empty |
| `RunID` | string | UUID, non-empty |
| `ReceiptID` | string | Lowercase SHA-256 digest |
| `DataTypeID` | string | Lowercase dotted identifier |
| `FormatID` | string | `csv`, `json`, or `excel` in V1; Google Sheets is an agent bridge, not a runtime format |
| `OperationID` | string | Lowercase dotted identifier |
| `SourceID` | string | Request-local identifier |
| `TableID` | string | Inspection-derived stable identifier |
| `ColumnID` | string | Stable within one inspected source table |
| `RowID` | string | Stable within one snapshot |
| `SHA256Digest` | string | 64 lowercase hexadecimal characters |
| `TypeAuthority` | string enum | `declared` or `inferred` |
| `FindingSeverity` | string enum | `info`, `warning`, or `error` |
| `TerminalStatus` | string enum | `succeeded`, `failed`, `cancelled`, or `blocked` |
| `OperationOwner` | string enum | `reader_engine`, `logical_engine`, or `writer_engine` |
| `OperationExposure` | string enum | `api`, `workflow`, or `lifecycle` |
| `StreamingMode` | string enum | `streaming` or `bounded_spool` |
| `CollisionPolicy` | string enum | `block`, `overwrite`, or `alternate_name` |
| `RemoteCollisionPolicy` | string enum | `block` or `alternate_name` |
| `ExceptionAction` | string enum | `block`, `report`, or `ignore` |
| `ExceptionDetail` | string enum | `summary` or `full` |

#### Scenario: Unknown enum value
- **WHEN** a request, descriptor, or configuration contains an enum value outside its fixed catalog
- **THEN** strict validation rejects it before planning

### Requirement: Exact canonical data structures
The root module SHALL implement the following exact semantic structures. Fields
marked internal are not serialized on the JSON API.

| Structure | Field | Go type | JSON field | Required | Contract |
|---|---|---|---|:---:|---|
| `Value` | `typeID` | `DataTypeID` | internal | yes | Registered data type |
|  | `encoded` | `[]byte` | internal | yes | Copied on construction/access; empty only for an empty `string` value |
| `Cell` | `ColumnID` | `ColumnID` | `column_id` | yes | Must match schema position |
|  | `Value` | `Value` | internal | yes | Immutable typed value |
| `ColumnDescriptor` | `ID` | `ColumnID` | `id` | yes | Unique in schema |
|  | `PhysicalPosition` | `string` | `physical_position` | yes | Source position such as `A` or JSON path |
|  | `SourceHeader` | `string` | `source_header` | yes | May be empty only when inspection reports it |
|  | `Subheader` | `*string` | `subheader` | no | Second header row |
|  | `Alias` | `*string` | `alias` | no | Resolved global alias |
|  | `SemanticType` | `*string` | `semantic_type` | no | Resolved semantic type name |
|  | `DataType` | `DataTypeID` | `data_type` | yes | Registered type |
|  | `TypeAuthority` | `TypeAuthority` | `type_authority` | yes | Declared or inferred |
| `Schema` | `ID` | `TableID` | `id` | yes | Stable table identity |
|  | `SourceID` | `SourceID` | `source_id` | yes | Owning source |
|  | `SheetID` | `*string` | `sheet_id` | no | Excel only |
|  | `Columns` | `[]ColumnDescriptor` | `columns` | yes | Non-empty, ordered, unique IDs |
| `Row` | `ID` | `RowID` | `id` | yes | Stable within snapshot |
|  | `SourceID` | `SourceID` | `source_id` | yes | Must match schema |
|  | `SheetID` | `*string` | `sheet_id` | no | Must match schema when present |
|  | `Ordinal` | `uint64` | `ordinal` | yes | Zero-based source data-row order |
|  | `Cells` | `[]Cell` | internal | yes | Same length/order as schema columns |
| `Dataset` | `Schema` | `Schema` | internal | yes | Validated before stream use |
|  | `Rows` | `RowStream` | internal | yes | Non-nil and closeable |

```go
type RowStream interface {
    Next(context.Context) (Row, error) // io.EOF after the last row
    Close() error                      // idempotent
}
```

Canonical null SHALL be a registered `DataTypeID` of `null`; empty string SHALL
be `string` with zero encoded length in its type-owned representation, and the
literal hyphen SHALL be an ordinary non-empty string. Decimal values SHALL use
exact decimal canonical bytes and SHALL NOT use `float64`.

#### Scenario: Row/schema mismatch
- **WHEN** a row has a different source, sheet, cell count, column order, or unregistered value type from its schema
- **THEN** validation fails before that row reaches a logical operation

### Requirement: Exact descriptors and reporting structures
Infrastructure SHALL define the following JSON structures for discovery,
findings, terminal reports, and compact receipts.

| Structure | Field | JSON type | Required | Contract |
|---|---|---:|:---:|---|
| `DataTypeDescriptor` | `id` | `DataTypeID` | yes | Stable type ID |
|  | `version` | string | yes | `v1` |
|  | `name` | string | yes | English display name |
|  | `canonical_kind` | string | yes | `null`, `string`, `integer`, `decimal`, `boolean`, `date`, or `time` |
| `StyleDescriptor` | `name` | string | yes | Unique case-insensitively |
|  | `right_to_left` | boolean | yes | Style direction |
| `OperationDescriptor` | `id` | `OperationID` | yes | Stable catalog ID |
|  | `version` | string | yes | `v1` |
|  | `owner` | `OperationOwner` | yes | Owning engine |
|  | `exposure` | `OperationExposure` | yes | API, workflow, or lifecycle |
|  | `name` | string | yes | English name |
|  | `description` | string | yes | English behavior summary |
|  | `deterministic` | boolean | yes | Must be `true` |
|  | `streaming_mode` | `StreamingMode` | yes | Streaming contract |
|  | `input_shape` | string | yes | Named contract input |
|  | `output_shape` | string | yes | Named contract output |
|  | `parameter_schema` | object | yes | Strict Draft 2020-12 JSON Schema |
|  | `required_capabilities` | string[] | yes | Sorted; may be empty |
| `OperationRef` | `operation_id` | `OperationID` | yes | Registered ID |
|  | `version` | string | yes | Resolved version |
|  | `owner` | `OperationOwner` | yes | Registry owner |
| `SourceLocator` | `source_id` | `SourceID` | yes | Source identity |
|  | `sheet_id` | string | no | Excel worksheet |
|  | `root_path` | string | no | JSON root |
|  | `row_ordinal` | integer | no | Zero-based data row |
|  | `column_id` | `ColumnID` | no | Stable column ID |
|  | `physical_position` | string | no | Cell/path locator |
| `OutputLocator` | `output_path` | string | yes | Requested normalized output path |
|  | `row_ordinal` | integer | no | Zero-based output data-row ordinal when row-specific |
|  | `column_id` | `ColumnID` | no | Present only when a column is identified |
|  | `physical_position` | string | no | Present only when format-specific output position is available |
| `Finding` | `code` | string | yes | Stable finding code |
|  | `severity` | `FindingSeverity` | yes | Severity enum |
|  | `message` | string | yes | English detail |
|  | `locator` | `SourceLocator` | no | Precise location |
|  | `requires_decision` | boolean | yes | Whether execution must wait |
| `ErrorDetail` | `key` | string | yes | Stable key |
|  | `value` | string | yes | Redacted deterministic value |
| `ValidationResult` | `kind` | string | yes | `format`, `schema`, `exact_columns`, or `all_rows_match` |
|  | `passed` | boolean | yes | Assertion result |
|  | `code` | string | no | Failure code |
|  | `message` | string | yes | English result |
|  | `locator` | `OutputLocator` | no | First reopened-output failure |
| `ExceptionSummary` | `action` | `ExceptionAction` | yes | Effective policy |
|  | `count` | integer | yes | Non-negative |
|  | `first_locator` | `SourceLocator` | no | First exception |
|  | `report_path` | string | no | Present only for full external report |
| `PublishedOutput` | `path` | string | yes | Final local path |
|  | `format_id` | `FormatID` | yes | Output format |
|  | `sha256` | `SHA256Digest` | yes | Published bytes |
|  | `size_bytes` | integer | yes | Non-negative |
| `RunReport` | `request_id` | `RequestID` | yes | API correlation |
|  | `run_id` | `RunID` | yes | Created before source access |
|  | `status` | `TerminalStatus` | yes | Terminal state |
|  | `output` | `PublishedOutput` | no | Success only |
|  | `validations` | `ValidationResult[]` | yes | Ordered; may be empty before staging |
|  | `findings` | `Finding[]` | yes | Bounded by policy |
|  | `exception_summary` | `ExceptionSummary` | no | Policy-dependent |
|  | `first_failure` | `APIError` | no | Failure/blocked/cancelled only |
|  | `source_snapshot_hash` | `SHA256Digest` | no | Present after snapshot |
|  | `effective_config_digest` | `SHA256Digest` | yes | Resolved config |
|  | `capabilities_digest` | `SHA256Digest` | yes | Runtime registry state |
|  | `receipt_id` | `ReceiptID` | yes | Terminal receipt |
|  | `started_at` | string | yes | UTC RFC3339Nano |
|  | `finished_at` | string | yes | UTC RFC3339Nano |

`RunReceipt` SHALL contain exactly: `receipt_id`, `request_id`, `run_id`,
`normalized_workflow_digest`, `binary_version`, `capabilities_digest`,
`effective_config_digest`, optional `source_snapshot_hash`, optional
`published_output`, `status`, optional `terminal_error` as one `APIError` without
its `run_context`, `started_at`, and `finished_at`. No other field is allowed.

#### Scenario: Receipt compactness
- **WHEN** a run terminates with findings and exceptions
- **THEN** the `RunReport` may contain them but the `RunReceipt` contains only its exact compact field set

`ConfigOverride` contains exactly `path:string` and `value:JSON value`. `path`
is an existing JSON Pointer to a config value; duplicate paths are invalid.
`WorkflowTemplateDocument` contains exactly `$schema`, `schema_version`,
`variables`, and `workflow`. Its `$schema` is
`urn:data-toolkit:workflow-template:v1`; it is resolved by the CLI before strict
`WorkflowRequest` decoding. `WorkflowVariable` contains `id`, `name`, `type`,
`required`, optional `default`, and optional `description`. IDs are contiguous
positive integers in declaration order, names are unique, and types are one of
`string`, `integer`, `number`, `boolean`, `array`, `object`, or `null`. When
present, `default` SHALL match the declared type. A supplied binding wins over a
default; a referenced required variable without a supplied binding or default
SHALL fail before strict `WorkflowRequest` decoding.
`VariableRef` is exactly `{ "$var": positive integer }` and may replace one JSON
value only, never an object key or substring.

### Requirement: Exact configuration structures
The strict YAML/JSON configuration SHALL use the following field catalog. Every
object rejects additional properties.

`ConfigDocument` is the root object in the table. `EffectiveConfig` is an
immutable resolved `ConfigDocument` with the same fields and types and no extra
JSON fields; its digest is carried separately in API and report structures.

| Group | Field | Type | Required | Validation/default |
|---|---|---|:---:|---|
| root | `schema_version` | string | yes | `v1` |
|  | `display` | `DisplayConfig` | yes | — |
|  | `aliases` | `AliasConfig[]` | yes | Default `[]` |
|  | `semantic_types` | `SemanticTypeConfig[]` | yes | Default built-ins |
|  | `styles` | `StyleConfig[]` | yes | At least one |
|  | `output` | `OutputConfig` | yes | — |
|  | `cleaning` | `CleaningConfig` | yes | — |
|  | `exceptions` | `ExceptionConfig` | yes | — |
|  | `runtime` | `RuntimeConfig` | yes | — |
|  | `report` | `ReportConfig` | yes | — |
|  | `cli` | `CLIConfig` | yes | — |
|  | `google_sheets` | `GoogleSheetsConfig` | yes | Agent bridge policy only; no credentials |
| display | `null_display` | string | yes | Default `-` |
|  | `decimal_precision` | integer | yes | Default `3`, minimum `0` |
|  | `date_format` | string | yes | Default `02/01/06` Go layout |
|  | `time_format` | string | yes | Default `15:04` Go layout |
|  | `phone_format` | string | yes | Default `052-6105412` pattern |
|  | `text_encoding` | string | yes | Constant `UTF-8` |
|  | `unicode_normalization` | string | yes | Constant `NFC` |
| alias item | `name` | string | yes | Unique case-insensitively |
|  | `semantic_type` | string | yes | Existing semantic type |
| semantic type item | `name` | string | yes | Unique case-insensitively |
|  | `data_type` | `DataTypeID` | yes | Registered data type |
|  | `default_format` | string | no | Type-compatible |
| style item | `name` | string | yes | Unique case-insensitively |
|  | `header_fill` | string | no | `#RRGGBB` |
|  | `header_text` | string | no | `#RRGGBB` |
|  | `alternating_row_fill` | string | no | `#RRGGBB` |
|  | `bold_headers` | boolean | yes | — |
|  | `right_to_left` | boolean | yes | — |
| output | `directory` | string | yes | Existing/creatable local directory |
|  | `filename_template` | string | yes | Non-empty |
|  | `default_style` | string | yes | Existing style name |
|  | `collision` | `CollisionPolicy` | yes | Default `block` |
|  | `output_delimiter` | string | yes | Default `,`; one Unicode scalar |
|  | `line_ending` | string enum | yes | `lf` or `crlf`; default `crlf` |
|  | `null_rendering_policy` | string enum | yes | `display` or `round_trip`; default `round_trip` |
| cleaning | `mixed_type_action` | `ExceptionAction` | yes | Default `block` |
|  | `deduplicate_keep` | string enum | yes | `first`, `last`, or `error`; default `first` |
|  | `auto_normalize_formats` | boolean | yes | Default `true` |
|  | `trim_detected_whitespace` | boolean | yes | Default `true` |
| exceptions | `action` | `ExceptionAction` | yes | Default `block` |
|  | `detail` | `ExceptionDetail` | yes | Default `summary` |
| runtime | `maximum_memory_bytes` | integer | yes | Positive; default `536870912` |
|  | `lock_retry_count` | integer | yes | Minimum `0`; default `5` |
|  | `lock_retry_interval` | duration string | yes | Positive; default `500ms` |
|  | `lock_timeout` | duration string | yes | Positive; default `5s` |
|  | `temporary_workspace` | string | yes | Local directory; default `data-toolkit` |
|  | `failed_run_retention` | duration string | yes | Non-negative; default `24h` |
| report | `receipt_store` | string | yes | Local append-only store path |
|  | `retention` | duration string | yes | Positive |
|  | `include_effective_config_digest` | boolean | yes | Must be `true` in V1 |
|  | `include_source_hash` | boolean | yes | Must be `true` in V1 |
|  | `include_validation_details` | boolean | yes | Default `true` |
| cli | `non_overridable_config_paths` | string[] | yes | Sorted, unique JSON Pointers to existing fields; default protects schema version, this list, receipt store, and temporary workspace |
| google sheets | `destination_folder` | string | yes | Drive folder ID, URL, or `root`; default `root` |
|  | `title_template` | string | yes | Default `{output_stem}`; only `{output_stem}`, `{workflow_id}`, `{run_id}` placeholders allowed |
|  | `collision` | `RemoteCollisionPolicy` | yes | Default `block` |

#### Scenario: Complete default configuration
- **WHEN** embedded defaults are decoded
- **THEN** every required field above is present, passes the same schema and semantic validation as user configuration, and resolves without mutation

### Requirement: Memory-budget enforcement
Every run SHALL use a run-scoped resource budget covering reader buffers,
logical-operation state, writer buffers, and bounded batches. Implementations
must reserve material state before using it, release reservations promptly, and
spill deterministically when an operation declares `bounded_spool`. A fixed
batch size is forbidden. If the remaining budget cannot support a required exact
operation or spill, the run fails with `resource_limit_exceeded`; no output is
published.

#### Scenario: Unique keyed deduplication exceeds memory
- **WHEN** exact deduplication state exceeds the available budget and workspace spill cannot proceed
- **THEN** the run terminates with `resource_limit_exceeded`, finalizes once, and publishes no output
