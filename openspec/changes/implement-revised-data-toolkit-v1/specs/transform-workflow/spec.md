## ADDED Requirements

### Requirement: Single-source single-output request
`WorkflowRequest` SHALL contain one `SourceRequest`, an ordered list of
registered `OperationCall` values, one `OutputRequest`, supported overrides, an
exception policy, and explicit output assertions. V1 SHALL reject multiple sources,
multiple outputs, queries, and unsupported fields.

#### Scenario: Second source rejected
- **WHEN** a request contains more than one source
- **THEN** schema validation fails before inspection or planning

### Requirement: Registry-based workflow validation
The orchestrator SHALL validate envelope version, identities, source and output
formats, operation versions and parameters, capability requirements, schema
references, configuration overrides, and V1 scope before execution.

#### Scenario: Invented capability
- **WHEN** a request references a format or operation absent from runtime registries
- **THEN** validation returns a structured capability error before source rows are read

### Requirement: Minimal explicit output assertions
V1 SHALL support exactly two user-declared assertion kinds: `exact_columns`,
which declares the exact ordered stable column IDs expected in the output, and
`all_rows_match`, which declares that every output row satisfies one valid
constant-only filter expression. The Writer Engine SHALL always reopen the staged
output and verify its format and declared schema; that mandatory safety check is
not a user-declared assertion. Unknown assertion kinds SHALL fail workflow
validation before source rows are read.

#### Scenario: Row predicate assertion fails
- **WHEN** a staged output contains a row that fails its declared `all_rows_match` expression
- **THEN** output validation fails with an `OutputLocator` for the requested output path, zero-based row ordinal when row-specific, and optional column ID or physical position when available; no final output is published

### Requirement: Ordered thin orchestration
The orchestrator SHALL preserve declared logical-operation order, reject
non-workflow operation exposure in user calls, and surround the declared calls
with the fixed registered Reader and Writer Engine lifecycle. It SHALL NOT implement
operation semantics, infer missing logical calls, search for plans, or solve constraints.

#### Scenario: Fixed engine ownership
- **WHEN** a valid workflow is planned
- **THEN** its logical calls retain request order and its fixed reader and writer lifecycle steps are selected from their registered descriptors without appearing as user calls

### Requirement: Safe transform lifecycle
Execution SHALL follow validated request and configuration resolution; shared
read lock, immutable source snapshot and source hash; canonical read with
schema and dirty-input validation; ordered transformations; staged rendering;
optional style; close and reopen; mandatory format/schema validation; explicit
output assertions; atomic publication; and `writer.finalize`. `writer.finalize`
SHALL run after workspace creation on every terminal path, close resources in
reverse acquisition order, apply artifact cleanup or retention policy, assemble
the `RunReport`, and append exactly one receipt. No successful publication is
permitted before validation.

#### Scenario: Successful run
- **WHEN** every step and validation succeeds
- **THEN** exactly one new output is published and the final report and receipt identify it

#### Scenario: Mid-run failure
- **WHEN** any operation fails after staging begins
- **THEN** no partial final output is reported and cleanup or retention follows effective policy

### Requirement: Cancellation and transient progress
The orchestrator SHALL propagate cancellation and MAY emit transient progress to
an in-process observer. It SHALL NOT persist a lifecycle-event stream or return
one as part of the terminal run result.

#### Scenario: Cancellation during streaming
- **WHEN** the request context is cancelled while rows are being consumed
- **THEN** stream processing stops, resources close, no output is published, and the terminal receipt records cancellation

### Requirement: Complete final report
Every run SHALL return output status, mandatory and explicit assertion results,
bounded findings or an exception summary when the effective policy requests
them, a first failure location when available, structured errors,
effective-configuration digest, source hash, capability versions, and run
receipt identity. Findings and exceptions SHALL NOT be persisted in the receipt.

#### Scenario: Report after blocked dirty input
- **WHEN** stream-time input validation finds mixed types and effective policy is `block`
- **THEN** the report identifies the blocking source location and contains no published output

### Requirement: Exact Transform Workflow structures
The transform contract SHALL use these exact strict structures. `operations`
contains only descriptors with `owner=logical_engine` and
`exposure=workflow`.

| Structure | Field | JSON type | Required | Default/contract |
|---|---|---:|:---:|---|
| `WorkflowRequest` | `schema_version` | string | yes | Constant `v1` |
|  | `id` | string | yes | Non-empty request-local workflow ID |
|  | `source` | `SourceRequest` | yes | Exactly one |
|  | `operations` | `OperationCall[]` | yes | Ordered; at least one |
|  | `output` | `OutputRequest` | yes | Exactly one |
|  | `overrides` | `WorkflowOverrides` | no | Supported policy fields only |
|  | `exceptions` | `ExceptionPolicy` | no | Effective config default |
|  | `validations` | `ValidationRequest[]` | yes | May be empty; mandatory reopen checks still apply |
| `OperationCall` | `id` | string | yes | Unique step ID within workflow |
|  | `operation_id` | `OperationID` | yes | Exact registered logical operation |
|  | `version` | string | yes | `v1` |
|  | `parameters` | object | yes | Must match descriptor JSON Schema |
| `WorkflowOverrides` | `display` | `DisplayOverrides` | no | Task-specific rendering preferences |
|  | `cleaning` | `CleaningOverrides` | no | Task-specific cleaning policy |
|  | `runtime` | `RuntimeOverrides` | no | Task-specific memory budget only |
| `ExceptionPolicy` | `action` | `ExceptionAction` | yes | `block`, `report`, or `ignore` |
|  | `detail` | `ExceptionDetail` | yes | `summary` or `full` |
|  | `maximum_details` | integer | yes | `0..10000`; default `1000` |

`WorkflowOverrides` SHALL NOT contain source paths, output paths, operation
registrations, style definitions, secrets, or report-retention policy.

Every override field is optional; an empty override object is invalid and SHALL
be omitted instead.

| Structure | Allowed optional fields |
|---|---|
| `DisplayOverrides` | `null_display:string`, `decimal_precision:integer`, `date_format:string`, `time_format:string`, `phone_format:string` |
| `CleaningOverrides` | `mixed_type_action:ExceptionAction`, `deduplicate_keep:"first"|"last"|"error"`, `auto_normalize_formats:boolean`, `trim_detected_whitespace:boolean` |
| `RuntimeOverrides` | `maximum_memory_bytes:integer` |

#### Scenario: Strict operation call
- **WHEN** an operation call omits its version, contains an unknown parameter, or references a lifecycle operation
- **THEN** workflow validation returns `invalid_workflow` before source access

### Requirement: Exact validation request structures
V1 user assertions SHALL use `ValidationRequest`, a tagged union with exactly
these variants.

| Structure | Fields | Contract |
|---|---|---|
| `ExactColumnsValidation` | `kind:"exact_columns"`, `column_ids:ColumnID[]` | Non-empty exact ordered output IDs |
| `AllRowsMatchValidation` | `kind:"all_rows_match"`, `expression:ExpressionNode` | Valid constant-only expression over output schema |

Each assertion object SHALL reject fields belonging to the other variant.
Mandatory staged-output `format` and `schema` validations are not request
objects and always run before these assertions.

#### Scenario: Unknown assertion field
- **WHEN** an assertion contains an unknown kind or a field from another variant
- **THEN** workflow validation rejects it before source access

### Requirement: Exact deterministic execution plan
The normalized plan SHALL contain these ordered steps and no inferred logical
operation: `reader.snapshot`, `reader.read`, every request operation in order,
`writer.write`, optional `writer.style` for Excel style output, `writer.validate`,
`writer.publish`, and `writer.finalize`. `writer.finalize` runs after every
terminal path; `writer.publish` runs only after all validations pass.

#### Scenario: Plan reproducibility
- **WHEN** the same normalized workflow, descriptors, and effective configuration are planned twice
- **THEN** step IDs, operation references, parameter digests, and order are byte-identical
