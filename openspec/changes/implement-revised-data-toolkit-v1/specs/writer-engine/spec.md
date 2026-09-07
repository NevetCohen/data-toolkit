## ADDED Requirements

### Requirement: Explicit Writer Engine registry
The Writer Engine SHALL register versioned writer-format descriptors and concrete
implementations for rendering, staging, optional styling, reopened validation,
publication, and terminal finalization. It SHALL publish only implemented writer
capabilities.

#### Scenario: Descriptor and implementation mismatch
- **WHEN** a writer descriptor declares a capability without its required implementation
- **THEN** registration fails before the capability becomes discoverable

### Requirement: Staged, reopened, validated output
`writer.write` SHALL render the final canonical stream to a staging path outside
the final target. `writer.style` is optional and Excel-only. `writer.validate`
SHALL close the staged artifact, reopen it, validate its format and declared
schema, and then evaluate requested assertions in request order. No artifact
that fails any validation may be published.

#### Scenario: Output validation fails
- **WHEN** format, schema, or a user assertion fails after staging
- **THEN** no final output is published and finalization follows the retention policy

### Requirement: Publication and terminal finalization
`writer.publish` SHALL apply local `block`, `overwrite`, or `alternate_name`
collision policy only after validation. `writer.finalize` SHALL run exactly once
for every attempted run after its workspace exists, including reader failure,
logical failure, writer failure, and cancellation. It SHALL close the registered
resources in reverse acquisition order, clean or retain workspace artifacts,
assemble the `RunReport`, append one compact receipt, and return the terminal
result to the Application API. The API serializes the result but does not own
finalization.

#### Scenario: Cancellation during streaming
- **WHEN** cancellation occurs after a run workspace exists
- **THEN** no output is published, `writer.finalize` closes resources and records one cancelled receipt

### Requirement: Exact writer request and descriptor structures
The Writer Engine SHALL use the following exact strict request and descriptor
structures. Output paths SHALL be normalized to absolute paths before request
hashing and comparison with the source path.

| Structure | Field | JSON type | Required | Contract |
|---|---|---:|:---:|---|
| `CSVOutputOptions` | `delimiter` | string | no | Overrides effective output delimiter |
|  | `line_ending` | string | no | `lf` or `crlf` |
|  | `null_rendering_policy` | string | no | `display` or `round_trip` |
| `OutputRequest` | `path` | string | yes | Must differ from normalized source path |
|  | `format_id` | `FormatID` | yes | `csv`, `json`, or `excel` |
|  | `sheet_name` | string | no | Excel output only; default `Data` |
|  | `style` | string | no | Named style; default effective style |
|  | `collision` | `CollisionPolicy` | no | Output override |
|  | `csv` | `CSVOutputOptions` | no | Allowed only when `format_id=csv` |
| `WriterFormatDescriptor` | `id` | `FormatID` | yes | Stable format ID |
|  | `version` | string | yes | `v1` |
|  | `name` | string | yes | English name |
|  | `extensions` | string[] | yes | Lowercase with leading dot |
|  | `mime_types` | string[] | yes | Sorted; may be empty |
|  | `capabilities` | string[] | yes | Non-empty subset of `write`, `style`, `validate` |
|  | `output_option_schema` | object | yes | Strict Draft 2020-12 JSON Schema |
|  | `limits` | `FormatLimits` | yes | Exact shared limits from Reader Engine |

`OutputRequest` SHALL reject every format-specific field that does not apply to
its selected format. Runtime discovery SHALL publish only a descriptor whose
declared capabilities have concrete registered implementations.

#### Scenario: Output equals source
- **WHEN** an output request normalizes to the same local path as the source
- **THEN** strict request validation rejects it before staging is created

### Requirement: Exact staged artifact and validation structures
The Writer lifecycle SHALL use the following exact internal/result structures.

| Structure | Field | Go/JSON type | Required | Contract |
|---|---|---|:---:|---|
| `StagedArtifact` | `path` | string | yes | Not the final target |
|  | `format_id` | `FormatID` | yes | Output format |
|  | `sha256` | `SHA256Digest` | yes | Closed staged bytes |
|  | `size_bytes` | int64 | yes | Non-negative |
| `OutputValidation` | `format` | `ValidationResult` | yes | Reopen and format check |
|  | `schema` | `ValidationResult` | yes | Declared schema check |
|  | `assertions` | `ValidationResult[]` | yes | User-request order |

#### Scenario: Publication input
- **WHEN** publication starts
- **THEN** it receives only a closed staged artifact whose format, schema, and requested assertions all passed

### Requirement: Writer Engine operation catalog
The Writer Engine registry SHALL contain exactly these lifecycle operations.

| ID | Version | Streaming | Input | Output |
|---|---|---|---|---|
| `writer.write` | `v1` | `streaming` | `Dataset + OutputRequest` | `StagedArtifact` |
| `writer.style` | `v1` | `bounded_spool` | `StagedArtifact` | `StagedArtifact` |
| `writer.validate` | `v1` | `streaming` | `StagedArtifact + assertions` | `OutputValidation` |
| `writer.publish` | `v1` | `streaming` | validated `StagedArtifact` | `PublishedOutput` |
| `writer.finalize` | `v1` | `streaming` | terminal run context + workspace | `RunReport` |

#### Scenario: Writer lifecycle operation in workflow
- **WHEN** a user `OperationCall` references a Writer Engine operation
- **THEN** workflow validation returns `invalid_workflow` before source access
