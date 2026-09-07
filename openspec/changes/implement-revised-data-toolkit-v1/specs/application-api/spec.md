## ADDED Requirements

### Requirement: Common JSON envelope
Every machine-facing command SHALL use the following exact JSON envelopes with
`additionalProperties: false`. JSON names are normative; timestamps use UTC
RFC3339Nano.

| Structure | Field | JSON type | Required | Contract |
|---|---|---:|:---:|---|
| `RequestEnvelope<T>` | `schema_version` | string | yes | Constant `v1` |
|  | `request_id` | string | yes | Non-empty UUID; injectable for tests |
|  | `payload` | object | yes | Command-specific request payload |
| `ResultEnvelope<T>` | `schema_version` | string | yes | Constant `v1` |
|  | `request_id` | string | yes | Must equal the request ID |
|  | `status` | string | yes | `ok` or `error` |
|  | `result` | object | conditional | Present only when `status=ok` |
|  | `error` | `APIError` | conditional | Present only when `status=error` |

#### Scenario: Exclusive result branch
- **WHEN** a command finishes
- **THEN** its envelope contains exactly one of `result` or `error` and preserves the request ID

### Requirement: Exact application API
The in-process API SHALL expose exactly these V1 methods and SHALL use the same
payload types as the JSON surface. An `APIError` is a product result; an ordinary
Go `error` is reserved for a process-level failure that prevented construction
of a valid result envelope.

```go
type API interface {
    Capabilities(context.Context, CapabilitiesRequest) (CapabilitiesResult, *APIError, error)
    DescribeOperation(context.Context, DescribeOperationRequest) (OperationDescriptor, *APIError, error)
    ValidateConfig(context.Context, ConfigValidationRequest) (ConfigValidationResult, *APIError, error)
    Inspect(context.Context, InspectionRequest) (InspectionReport, *APIError, error)
    ValidateWorkflow(context.Context, WorkflowValidationRequest) (WorkflowValidationResult, *APIError, error)
    Run(context.Context, RunRequest) (RunReport, *APIError, error)
}
```

#### Scenario: Thin interface client
- **WHEN** the CLI invokes a product operation
- **THEN** it calls one API method and does not reproduce validation, registry, planning, or adapter logic

### Requirement: Capability discovery contract
`CapabilitiesRequest` SHALL be an empty strict object. `CapabilitiesResult`
SHALL contain the exact fields below, sorted by normalized ID and version.

| Field | JSON type | Required | Contract |
|---|---:|:---:|---|
| `api_version` | string | yes | Constant `v1` |
| `binary_version` | string | yes | Build version |
| `capabilities_digest` | string | yes | Lowercase SHA-256 of normalized descriptors |
| `data_types` | `DataTypeDescriptor[]` | yes | May be empty |
| `reader_formats` | `ReaderFormatDescriptor[]` | yes | May be empty |
| `writer_formats` | `WriterFormatDescriptor[]` | yes | May be empty |
| `operations` | `OperationDescriptor[]` | yes | Reader, logical, and Writer operations |
| `styles` | `StyleDescriptor[]` | yes | Named compiled/configured styles |

#### Scenario: Runtime-only discovery
- **WHEN** registration changes at composition time
- **THEN** the discovery result and digest change without editing CLI or Codex Skill catalogs

### Requirement: Operation description contract
`DescribeOperationRequest` SHALL contain `operation_id` and optional `version`.
Omitted `version` selects the only registered version or fails if more than one
is registered. The successful result SHALL be one complete
`OperationDescriptor`, including its parameter JSON Schema.

| Field | JSON type | Required | Contract |
|---|---:|:---:|---|
| `operation_id` | string | yes | Exact registered ID |
| `version` | string | no | Semantic contract version such as `v1` |

#### Scenario: Ambiguous operation version
- **WHEN** the requested operation has multiple registered versions and no version is supplied
- **THEN** the API returns `ambiguous_capability_version`

### Requirement: Configuration validation contract
`ConfigValidationRequest` SHALL contain `config_path`; the path identifies one
local YAML or JSON configuration document. `ConfigValidationResult` SHALL not
echo the full configuration.

| Structure | Field | JSON type | Required | Contract |
|---|---|---:|:---:|---|
| `ConfigValidationRequest` | `config_path` | string | yes | Existing local file path |
| `ConfigValidationResult` | `valid` | boolean | yes | `true` on successful result |
|  | `schema_version` | string | yes | Constant `v1` |
|  | `effective_config_digest` | string | yes | SHA-256 after defaults and user layer |
|  | `warnings` | `Finding[]` | yes | Bounded; may be empty |

#### Scenario: Unknown configuration field
- **WHEN** the configuration contains a field outside the strict V1 schema
- **THEN** the API returns `invalid_config` with the exact JSON field path

### Requirement: Inspection API contract
`InspectionRequest` SHALL contain one `SourceRequest`, optional
`user_config_path`, and bounded inspection limits. Its successful result SHALL
be the exact `InspectionReport` defined by the Infrastructure and Reader Engine
specifications.

| Field | JSON type | Required | Default | Contract |
|---|---:|:---:|---:|---|
| `source` | `SourceRequest` | yes | — | One local source |
| `user_config_path` | string | no | embedded defaults | YAML or JSON config |
| `sample_rows` | integer | no | `20` | Range `0..100` |
| `maximum_findings` | integer | no | `100` | Range `1..1000` |

#### Scenario: Inspection needs a decision
- **WHEN** a sheet, delimiter, header, root path, or duplicate header remains ambiguous
- **THEN** inspection succeeds with `requires_decision=true` and structured decision items rather than guessing

### Requirement: Workflow validation API contract
`WorkflowValidationRequest` SHALL contain one exact resolved `WorkflowRequest`,
an optional user configuration path, and optional ordered `ConfigOverride[]`.
The CLI resolves a `WorkflowTemplateDocument` before constructing this request.
The result SHALL contain the normalized workflow digest and complete referenced
descriptors without consuming source rows.

| Structure | Field | JSON type | Required | Contract |
|---|---|---:|:---:|---|
| `WorkflowValidationRequest` | `workflow` | `WorkflowRequest` | yes | Strict V1 workflow |
|  | `user_config_path` | string | no | Optional YAML/JSON config |
|  | `config_overrides` | `ConfigOverride[]` | no | CLI-produced, ordered, policy-checked patches |
| `WorkflowValidationResult` | `valid` | boolean | yes | `true` on successful result |
|  | `normalized_workflow_digest` | string | yes | Lowercase SHA-256 |
|  | `effective_config_digest` | string | yes | Lowercase SHA-256 |
|  | `resolved_operations` | `OperationRef[]` | yes | Same order as calls |
|  | `google_sheets_policy` | `GoogleSheetsConfig` | yes | Resolved non-secret bridge policy for Codex handoff |
|  | `warnings` | `Finding[]` | yes | Bounded; may be empty |

#### Scenario: Validation does not read data rows
- **WHEN** a structurally valid workflow is validated
- **THEN** schemas, registries, paths, and parameters are checked without opening a canonical row stream

### Requirement: Run API contract
`RunRequest` SHALL contain the previously valid `WorkflowRequest`, optional user
configuration path, optional ordered `ConfigOverride[]`, and optional
`expected_workflow_digest`. This optionality applies only to direct in-process
API invocation; when the digest is provided it SHALL match current normalization
before the source is opened.

| Field | JSON type | Required | Contract |
|---|---:|:---:|---|
| `workflow` | `WorkflowRequest` | yes | Strict V1 workflow |
| `user_config_path` | string | no | Optional YAML/JSON config |
| `config_overrides` | `ConfigOverride[]` | no | Same effective layer used by validation |
| `expected_workflow_digest` | string | no | Prevents validate/run drift |

After a run workspace exists, every terminal execution outcome SHALL return one
`RunReport` as `ResultEnvelope.status=ok`; the report's own `status` and optional
`first_failure` distinguish success, failure, blockage, and cancellation. The
CLI SHALL still derive a non-zero exit class from a non-succeeded report.

Failures before workspace creation SHALL use `ResultEnvelope.status=error` with
one `APIError` and no `run_context`. If a run ID exists but the workspace was
not created and finalization cannot begin, the error SHALL attach compact
`run_context` with no `receipt_id`. Every failure after workspace creation SHALL
finalize and return the complete report produced by `writer.finalize` through
the exclusive result branch.

#### Scenario: Validate-run drift
- **WHEN** `expected_workflow_digest` differs from the normalized current request
- **THEN** the API returns `workflow_digest_mismatch` before source access

#### Scenario: Finalized failed run
- **WHEN** a Reader, Logical Engine, Writer, or cancellation failure occurs after workspace creation
- **THEN** the API returns the complete finalized `RunReport` with one receipt, and the CLI maps its `first_failure` to the documented non-zero exit class

### Requirement: Stable API error catalog
`APIError` SHALL use the exact fields and codes below. New error codes require a
new OpenSpec change; free-form messages are not control flow.

| Field | JSON type | Required | Contract |
|---|---:|:---:|---|
| `code` | string | yes | One catalog value below |
| `message` | string | yes | English human-readable summary |
| `component` | string | yes | `infrastructure`, `reader_engine`, `logical_engine`, `writer_engine`, `orchestrator`, `cli`, or `codex_adapter` |
| `field_path` | string | no | JSON Pointer when applicable |
| `locator` | `SourceLocator` or `OutputLocator` | no | Exactly one variant when present; precise source or output location |
| `retryable` | boolean | yes | Whether an unchanged request may be retried |
| `details` | `ErrorDetail[]` | yes | Sorted by key; may be empty |
| `run_context` | `RunFailureContext` | no | Only after a run ID exists |

`RunFailureContext` contains exactly required `run_id:RunID` and
`status:TerminalStatus`, plus optional `receipt_id:ReceiptID`, which SHALL be
present only when a receipt has already been appended. It prevents recursive
error and report structures.

Error codes are: `invalid_request`, `invalid_config`, `invalid_workflow`,
`unknown_capability`, `ambiguous_capability_version`, `inspection_ambiguous`,
`dirty_input`, `source_not_found`, `source_locked`, `source_changed`,
`workflow_digest_mismatch`, `operation_failed`, `resource_limit_exceeded`, `output_invalid`,
`output_collision`, `publish_failed`, `cancelled`, and `internal_error`.

#### Scenario: Stable programmatic failure
- **WHEN** the same invalid request is evaluated against the same registry and configuration
- **THEN** the error code, component, field path, retryability, and ordered details are identical

### Requirement: CLI command and exit mapping
The CLI SHALL map commands to API methods and exit classes exactly as follows.

| Command | API method | Request source |
|---|---|---|
| `capabilities --json` | `Capabilities` | CLI constructs empty payload |
| `describe --kind operation --id <id> [--version <v>] --json` | `DescribeOperation` | flags |
| `config validate --request <file> --json` | `ValidateConfig` | request envelope file |
| `inspect --request <file> --json` | `Inspect` | request envelope file |
| `workflow validate --file <workflow.json> [--config <file>] [--var ...] [--config-set ...] --json` | `ValidateWorkflow` | resolved workflow template |
| `run --file <workflow.json> [--config <file>] [--var ...] [--config-set ...] --expected-workflow-digest <digest> --json` | `Run` | resolved workflow template; digest mandatory for CLI run |

| Exit | Class | Error codes |
|---:|---|---|
| `0` | success | none |
| `2` | usage | flag/argument error before envelope construction |
| `3` | request decode | `invalid_request` caused by unreadable/malformed request file |
| `4` | validation | `invalid_config`, `invalid_workflow`, `workflow_digest_mismatch` |
| `5` | capability | `unknown_capability`, `ambiguous_capability_version` |
| `6` | decision required | `inspection_ambiguous`, `dirty_input` |
| `7` | source access | `source_not_found`, `source_locked`, `source_changed` |
| `8` | execution | `operation_failed`, `resource_limit_exceeded`, `internal_error` after run start |
| `9` | output validation | `output_invalid` |
| `10` | publication | `output_collision`, `publish_failed` |
| `130` | cancellation | `cancelled` |

#### Scenario: JSON stdout on non-zero exit
- **WHEN** a JSON command ends with any non-zero product exit class
- **THEN** stdout contains exactly one valid `ResultEnvelope` and diagnostics, if any, are written only to stderr

For a finalized `run`, the CLI SHALL select the exit class from
`RunReport.first_failure.code` or cancellation status while preserving the
successful result-envelope branch that contains the terminal report.
