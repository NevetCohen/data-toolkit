## ADDED Requirements

### Requirement: Explicit Reader Engine registry
The Reader Engine SHALL register versioned reader-format descriptors and concrete
implementations for source inspection and canonical reading. It SHALL publish
only reader capabilities that are actually implemented.

#### Scenario: Descriptor and implementation mismatch
- **WHEN** a reader descriptor declares inspection or reading without its required implementation
- **THEN** registration fails before the capability becomes discoverable

### Requirement: Deterministic inspection
`reader.inspect` SHALL inspect a source without changing it or creating a data
output. It SHALL return a common `InspectionReport` plus one format-specific
details object containing bounded samples, stable column IDs, types, counts,
findings, and required user decisions. Inspection is outside Transform Workflow;
a run SHALL NOT make an additional full-source inspection pass.

#### Scenario: Ambiguous source structure
- **WHEN** delimiter, header, sheet, root path, type, or cleanliness cannot be selected deterministically
- **THEN** inspection returns a blocking decision instead of guessing

### Requirement: Immutable source snapshot and canonical read
After the orchestrator validates and plans a run, `reader.snapshot` SHALL acquire
shared read access under the effective lock policy, create one immutable workspace
snapshot, and calculate its SHA-256 while copying. `reader.read` SHALL consume
only that completed snapshot, decode directly to `Schema + RowStream + Value`, and
validate schema and dirty values before a row reaches the Logical Engine.

#### Scenario: Source remains unchanged
- **WHEN** a run succeeds, fails, or is cancelled after snapshot creation
- **THEN** the source has only been opened for shared read access and the report hash identifies the exact processed bytes

#### Scenario: Blocking dirty value
- **WHEN** a value cannot be decoded as the selected schema type and the effective policy is `block`
- **THEN** reading stops at the first precise source locator and no Writer Engine publication occurs

### Requirement: Deterministic source-format contracts
CSV reader behavior SHALL be UTF-8/NFC with closed-set delimiter detection and
lossless null distinction. JSON reading SHALL require an explicit or unambiguous
scalar root path and SHALL not flatten or explode. Excel reading SHALL default to
the first sheet, honor an explicit sheet, and read cached/displayed formula values
without launching Excel. These rules are reader responsibilities.

#### Scenario: Unsupported nested JSON source
- **WHEN** a JSON source requires implicit flattening or exploding to become tabular
- **THEN** Reader Engine returns a capability error before any Logical Engine row is emitted

### Requirement: Exact reader request and descriptor structures
The Reader Engine SHALL use the following exact strict structures. Paths are
local paths and SHALL be normalized to absolute paths before request hashing.

| Structure | Field | JSON type | Required | Contract |
|---|---|---:|:---:|---|
| `SourceRequest` | `id` | `SourceID` | yes | Request-local identity |
|  | `path` | string | yes | Existing local file; read-only |
|  | `format_id` | `FormatID` | no | Otherwise infer from registered extensions; ambiguity fails |
|  | `sheet` | string | no | Excel only; default first sheet |
|  | `root_path` | string | no | JSON Pointer; required when root is ambiguous |
|  | `header_row` | integer | no | One-based; default inspection result |
|  | `delimiter` | string | no | CSV only; one Unicode scalar |
| `ReaderFormatDescriptor` | `id` | `FormatID` | yes | Stable format ID |
|  | `version` | string | yes | `v1` |
|  | `name` | string | yes | English name |
|  | `extensions` | string[] | yes | Lowercase with leading dot |
|  | `mime_types` | string[] | yes | Sorted; may be empty |
|  | `capabilities` | string[] | yes | Non-empty subset of `inspect`, `read` |
|  | `source_option_schema` | object | yes | Strict Draft 2020-12 JSON Schema |
|  | `limits` | `FormatLimits` | yes | Explicit V1 restrictions |
| `FormatLimits` | `maximum_sources` | integer | yes | `1` in V1 |
|  | `maximum_outputs` | integer | yes | `1` in V1 |
|  | `maximum_sheets_per_run` | integer | yes | `1` in V1 |
|  | `supports_nested_values` | boolean | yes | `false` in the V1 output table model |

`SourceRequest` SHALL reject every format-specific field that does not apply to
its resolved format. Runtime discovery SHALL publish only a descriptor whose
declared capabilities have concrete registered implementations.

#### Scenario: Non-local source path
- **WHEN** a source request contains a remote URL or a mutable write path
- **THEN** strict request validation rejects it before snapshot acquisition

### Requirement: Exact inspection result structures
Inspection SHALL return the shared fields plus exactly one format-specific
details object selected by `format_id`.

| Structure | Field | JSON type | Required | Contract |
|---|---|---:|:---:|---|
| `InspectionReport` | `source_id` | `SourceID` | yes | Matches request |
|  | `format_id` | `FormatID` | yes | Resolved format |
|  | `descriptor_version` | string | yes | Reader provider version |
|  | `source_size_bytes` | integer | yes | Non-negative |
|  | `source_sha256` | `SHA256Digest` | yes | Inspected bytes |
|  | `tables` | `InspectedTable[]` | yes | Non-empty on usable source |
|  | `findings` | `Finding[]` | yes | Bounded and ordered |
|  | `decisions` | `InspectionDecision[]` | yes | May be empty |
|  | `requires_decision` | boolean | yes | True iff a blocking decision exists |
|  | `csv` | `CSVInspectionDetails` | conditional | CSV only |
|  | `json` | `JSONInspectionDetails` | conditional | JSON only |
|  | `excel` | `ExcelInspectionDetails` | conditional | Excel only |
| `InspectedTable` | `schema` | `Schema` | yes | Stable IDs and inferred types |
|  | `row_count` | integer | no | Omitted when bounded inspection cannot know it |
|  | `sample_rows` | `InspectionSampleRow[]` | yes | At most request limit |
|  | `null_counts` | `ColumnCount[]` | yes | Counts observed or scanned |
|  | `empty_counts` | `ColumnCount[]` | yes | Counts observed or scanned |
| `InspectionSampleRow` | `ordinal` | integer | yes | Source data-row ordinal |
|  | `values` | `InspectionSampleValue[]` | yes | Display-safe bounded sample |
| `InspectionSampleValue` | `column_id` | `ColumnID` | yes | Stable column |
|  | `type_id` | `DataTypeID` | yes | Inferred or declared type |
|  | `display` | string | yes | Redacted or truncated to 256 Unicode scalars |
| `ColumnCount` | `column_id` | `ColumnID` | yes | Stable column |
|  | `count` | integer | yes | Non-negative |
| `InspectionDecision` | `code` | string | yes | `select_sheet`, `select_delimiter`, `select_header_row`, `select_root_path`, `select_column_id`, `select_type`, or `resolve_dirty_input` |
|  | `message` | string | yes | English question context |
|  | `options` | string[] | yes | Sorted; non-empty |
|  | `field_path` | string | yes | Request JSON Pointer to fill |

| Format details | Fields and exact types |
|---|---|
| `CSVInspectionDetails` | `delimiter:string`, `delimiter_candidates:string[]`, `has_header:boolean`, `header_row:integer`, `encoding:"UTF-8"`, `line_ending:"lf"|"crlf"|"mixed"`, `multiline_fields:boolean` |
| `JSONInspectionDetails` | `root_path:string`, `root_candidates:string[]`, `root_kind:"array"`, `scalar_paths:string[]`, `nested_paths:string[]` |
| `ExcelInspectionDetails` | `sheets:string[]`, `selected_sheet:string`, `header_row:integer`, `formula_cells:integer`, `missing_formula_cache_cells:integer` |

#### Scenario: Format detail exclusivity
- **WHEN** an inspection report is serialized
- **THEN** exactly one of `csv`, `json`, or `excel` is present and matches `format_id`

### Requirement: Exact immutable snapshot structure
The Reader lifecycle SHALL produce the following exact internal structure.

| Structure | Field | Go type | Required | Contract |
|---|---|---|:---:|---|
| `Snapshot` | `source_id` | `SourceID` | yes | Request source |
|  | `path` | string | yes | Immutable workspace file |
|  | `sha256` | `SHA256Digest` | yes | Hash of completed copied bytes |
|  | `size_bytes` | int64 | yes | Non-negative |

#### Scenario: Snapshot before row access
- **WHEN** `reader.read` starts
- **THEN** it receives only a completed closed snapshot whose size and SHA-256 were verified

### Requirement: Reader Engine operation catalog
The Reader Engine registry SHALL contain exactly these operations. Lifecycle
operations are selected by the orchestrator and SHALL NOT appear in user calls.

| ID | Version | Exposure | Streaming | Input | Output |
|---|---|---|---|---|---|
| `reader.inspect` | `v1` | `api` | `bounded_spool` | `InspectionRequest` | `InspectionReport` |
| `reader.snapshot` | `v1` | `lifecycle` | `streaming` | `SourceRequest` | `Snapshot` |
| `reader.read` | `v1` | `lifecycle` | `streaming` | `Snapshot + Schema` | `Dataset` |

#### Scenario: Lifecycle operation in workflow
- **WHEN** a user `OperationCall` references a Reader Engine operation
- **THEN** workflow validation returns `invalid_workflow` before source access
