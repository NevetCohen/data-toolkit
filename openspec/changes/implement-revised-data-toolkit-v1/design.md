## Context

Data Toolkit V1 is a deterministic local transformation system. The root module
already contains foundation code for canonical values, registries,
configuration, orchestration, an application service, and CLI scaffolding, but
that code is not the finished V1 product. Its public data model is `Schema +
RowStream + Value`; it must not modify sources and must make publication safe
and auditable. This change freezes the implementation contract before Go work
begins.

The implementation must preserve external sources, keep `legacy/mvp` outside
the root dependency graph, use one Go binary, and remain deterministic at the
target scale. Approved production datasets and the revised Obsidian brief are
read-only; only sanitized representative fixtures may enter the repository.

## Goals

- Give one owner to each execution phase and make the execution plan byte-stable.
- Support strict file-backed workflow templates, repeatable runtime values, and
  safe configuration patches.
- Keep local processing, resource use, cleanup, reporting, and receipts bounded
  and deterministic.
- Allow Codex to bridge Google Sheets with the installed Google Drive connector
  without extending the Go runtime across a credentials or network boundary.
- Deliver active CSV, JSON, and Excel Reader/Writer paths through one typed
  canonical stream and a narrow ten-operation logical catalog.
- Establish focused contract tests at each boundary before broadening the first
  end-to-end vertical slice.

## Non-goals

- Implementing the Go engines, CLI, connector calls, or batch stream.
- Google Sheets as a `FormatID`, direct network adapter, remote overwrite,
  multi-source joins, queries, TXT, or TUI support.
- Interpolating variables inside strings, JSON keys, or field names.
- Dynamic runtime plugins, executable configuration, a general planner, or
  importing `legacy/mvp` as current contract code.
- Per-cell lineage, rollback machinery, persistent lifecycle event logs, or a
  full request/configuration copy in the compact receipt.

## Architecture

| Subsystem | Sole responsibility |
|---|---|
| Infrastructure | typed immutable values, schemas, registries, errors, config, resources, receipts |
| Reader Engine | inspect, shared read lock, immutable snapshot/hash, decode and stream-time input validation |
| Logical Engine | registered user operations in declared order |
| Writer Engine | staged rendering, style, reopen validation, assertions, atomic publish, finalization |
| Orchestrator | request validation, configuration resolution, deterministic planning, and phase dispatch |
| CLI | JSON file loading, typed substitution, config patches, JSON envelopes and exit classes |
| Codex Adapter | discovery, template construction, run handoff, recovery, and Google Sheets XLSX bridge |

Inspection is a Reader operation before workflow execution. A planned run is:

`validate -> resolve config -> plan -> reader.snapshot -> reader.read -> logical
operations -> writer.write -> writer.style? -> writer.validate -> writer.publish
-> writer.finalize`.

After workspace creation, `writer.finalize` always runs. It releases resources
in reverse acquisition order, applies cleanup or retention policy, creates the
terminal `RunReport`, and appends exactly one receipt. The application API only
returns that envelope.

## Key decisions

### 1. Split Reader and Writer capability ownership

The old combined file lifecycle is split into `reader.inspect`,
`reader.snapshot`, `reader.read` and `writer.write`, `writer.style`,
`writer.validate`, `writer.publish`, `writer.finalize`. `CapabilitiesResult`
therefore exposes `reader_formats` and `writer_formats`; owners are
`reader_engine`, `logical_engine`, and `writer_engine`.

### 2. Reader validates before the logical engine

Reader acquires a shared-read lock, builds an immutable snapshot and hash, then
decodes directly to schema, stream, and values. Schema and dirty-input checks
occur before a value is emitted to a logical operation. Snapshot mismatch is a
source failure, not an implicit retry.

### 3. Writer owns the terminal safety boundary

Writer renders to staging, applies optional Excel style, closes and reopens the
artifact, runs mandatory format/schema checks and user assertions, then atomically
publishes according to collision policy. Any failure and cancellation reaches
`writer.finalize`; no later operation appends another receipt.

### 4. Strict workflow templates

`WorkflowTemplateDocument` has `$schema`, `schema_version`, `variables`, and
`workflow`. Variable IDs are positive contiguous integers beginning at one;
names are unique descriptive identifiers; each declaration includes type,
required, optional default, and optional description. A reference is exactly
`{"$var": id}` in a value position. Substitution does not operate in a string,
field name, or JSON key. CLI validates substitutions before strict
`WorkflowRequest` validation and digest calculation.

### 5. Configuration patches have a narrow, verified boundary

The CLI accepts repeatable `--config-set <json-pointer>=<json-value>` as
`ConfigOverride{path,value}`. Duplicate, missing, structural-invalid, protected,
ancestor, and descendant paths fail. Precedence is defaults, user config, CLI
patches, workflow overrides, and output overrides. Full validation follows each
resolution and the result contributes to the effective-config digest.

`cli.non_overridable_config_paths` protects at least `/schema_version`, the
denylist itself, `/report/receipt_store`, and `/runtime/temporary_workspace`.

### 6. Google Sheets is an agent bridge, not a runtime adapter

The sixth project Skill checks the actual Drive connector first. A source Sheet
is metadata-verified, exported to XLSX, processed by the local chain using only
the returned absolute workspace path, and the temporary XLSX is removed in a
finally path. A destination local XLSX is validated first. Before import, the
Skill performs an exact paginated title/MIME/parent search; `block` stops and
`alternate_name` selects the first free deterministic numeric suffix. The Skill
then imports natively, verifies and optionally moves after parent readback,
performs bounded structural/display readback, and locally deletes the XLSX.
Drive files are never deleted. Over-10-MB export and missing connector cases are
explicit capability gaps with manual XLSX fallbacks. `google_sheets` config
supplies folder, title template (default `{output_stem}`), and `block` or
`alternate_name` collision policy.

### 7. Bounded logical operations and memory

`row.deduplicate` supports `first`, `last`, and `error` through bounded spool;
disk-backed state is allowed only when needed by the budget. `row.sort@v1`
sorts multiple keys with typed deterministic comparison, explicit null order,
NFC Unicode code-point string order, and input-order stability for equal keys.
It uses external merge/bounded spool and obeys cancellation/cleanup. Components
reserve from the run budget and spill before exhaustion; an unrecoverable limit
is `resource_limit_exceeded` in the execution exit class, with no publication.

The public stream contract does not expose a fixed batch. A row-oriented wide
baseline is measured first. Any internal 4K, 16K, or 64K prototype is permitted
only after allocations or GC identify a bottleneck and only if row order, nulls,
exact numbers, output hash, and first-error location remain identical.

### 8. Deterministic format contracts remain separate from logical operations

CSV is UTF-8/NFC with closed-set delimiter inspection and lossless null rules.
JSON requires an explicit or unambiguous scalar root and never implicitly
flattens or explodes. Excel defaults to the first sheet, reads cached/displayed
formula values without launching Excel, and applies only a named final style.
Reader and Writer descriptors expose only implemented format capabilities;
Logical Engine never branches on format.

### 9. Finalized runs return complete reports

The common JSON envelope remains exclusive: either `result` or `error`.
Pre-workspace validation/setup failures use the error branch. After workspace
creation, `writer.finalize` produces the terminal `RunReport`, and the API
returns it through the result branch even when the report status is failed,
blocked, or cancelled. The CLI derives the non-zero exit class from the report's
terminal failure while preserving the complete report for the interface.

### 10. Configuration changes are atomic and non-secret

The strict schema, Go types, embedded defaults, CLI patches, and semantic
validation change together. The effective precedence is defaults, user config,
CLI patches, workflow overrides, and output overrides. Protected JSON Pointers,
their ancestors, and their descendants cannot be patched. Registries, source
selection, credentials, prompts, and inspection results are not configuration.

### 11. CLI and Codex consume runtime truth

The executable exposes the fixed command families in the Application API spec.
`capabilities` and `describe` serialize live registries. The six Skills define
procedure and handoffs, not a duplicated operation/format catalog. Ambiguity
causes a focused question; a missing capability causes an explicit stop rather
than a bypass script.

### 12. Deliver one gated vertical slice before broadening

The first slice uses one CSV source, a constant filter, one further logical
operation, one reopened/schema-validated CSV output, immutable-source checks,
deterministic bytes, cancellation, bounded memory, and one compact receipt.
Only after that gate passes are JSON, Excel, the practical workflows, Google
bridge fixtures, and the Codex reliability suite broadened.

### 13. Treat cleanup-script evidence as scoped coverage

V1 does not replace or claim equivalence with
`Invoke-RawMemberCleanup.Streaming.py`. The static, sanitized cleanup manifest
preserves exactly four behaviors that can be expressed through the fixed ten
logical operations as coverage evidence. Its other fourteen behaviors remain
`outside_v1` because they exceed the fixed catalog or the single-source,
single-new-output workflow boundary. They are input for a future OpenSpec
change that this change MUST NOT create.

## Normative contract map

The specifications, not implementation tasks, are authoritative for every V1
shape and identifier. An implementer SHALL stop for an OpenSpec correction
rather than invent a missing field, operation, error, or Skill contract.

| Contract family | Normative specification | Fixed contents |
|---|---|---|
| Primitive IDs, canonical data, configuration, reports, receipts, templates | `specs/infrastructure-contracts/spec.md` | Exact aliases, enums, structures, fields, defaults, patch rules, and memory budget |
| Application API | `specs/application-api/spec.md` | Envelopes, six methods, payloads, errors, commands, exits, and finalized-run behavior |
| Reader Engine | `specs/reader-engine/spec.md` | Source request, Reader descriptor, inspection structures, snapshot, and three operations |
| Writer Engine | `specs/writer-engine/spec.md` | Output request/options, Writer descriptor, staged/validation structures, and five operations |
| Workflow | `specs/transform-workflow/spec.md` | Workflow, operation calls, overrides, exception policy, assertions, and plan |
| Logical Engine | `specs/logical-operations/spec.md` | Descriptor, expressions, ten operations, parameters, and failure reasons |
| CLI | `specs/cli-json-surface/spec.md` and `specs/application-api/spec.md` | File-backed commands, substitution, patches, stdout/stderr, and exits |
| Codex Adapter | `specs/codex-adapter/spec.md` | Six Skills, inputs, outputs, handoffs, Google bridge, and stop conditions |
| Acceptance | `specs/v1-acceptance/spec.md` | Vertical slice, practical workflows including scoped cleanup coverage, fidelity/performance, Codex, and quality gates |

### Data structure index

| Owner | Structures |
|---|---|
| Canonical core | `Value`, `Cell`, `ColumnDescriptor`, `Schema`, `Row`, `RowStream`, `Dataset` |
| Descriptor/report infrastructure | `DataTypeDescriptor`, `StyleDescriptor`, `OperationDescriptor`, `OperationRef`, `SourceLocator`, `OutputLocator`, `Finding`, `ErrorDetail`, `ValidationResult`, `ExceptionSummary`, `PublishedOutput`, `RunReport`, `RunReceipt` |
| Configuration and templates | `ConfigDocument`, `DisplayConfig`, `AliasConfig`, `SemanticTypeConfig`, `StyleConfig`, `OutputConfig`, `CleaningConfig`, `ExceptionConfig`, `RuntimeConfig`, `ReportConfig`, `CLIConfig`, `GoogleSheetsConfig`, `EffectiveConfig`, `ConfigOverride`, `WorkflowTemplateDocument`, `WorkflowVariable`, `VariableRef` |
| Application API | `RequestEnvelope<T>`, `ResultEnvelope<T>`, `APIError`, `RunFailureContext`, `CapabilitiesRequest`, `CapabilitiesResult`, `DescribeOperationRequest`, `ConfigValidationRequest`, `ConfigValidationResult`, `InspectionRequest`, `WorkflowValidationRequest`, `WorkflowValidationResult`, `RunRequest` |
| Reader Engine | `SourceRequest`, `ReaderFormatDescriptor`, `FormatLimits`, `InspectionReport`, `InspectedTable`, `InspectionSampleRow`, `InspectionSampleValue`, `ColumnCount`, `InspectionDecision`, `CSVInspectionDetails`, `JSONInspectionDetails`, `ExcelInspectionDetails`, `Snapshot` |
| Writer Engine | `CSVOutputOptions`, `OutputRequest`, `WriterFormatDescriptor`, `StagedArtifact`, `OutputValidation` |
| Workflow | `WorkflowRequest`, `OperationCall`, `WorkflowOverrides`, `DisplayOverrides`, `CleaningOverrides`, `RuntimeOverrides`, `ExceptionPolicy`, `ValidationRequest`, `ExactColumnsValidation`, `AllRowsMatchValidation` |
| Logical expression and parameters | `TypedLiteral`, `ExpressionNode`, `LogicalGroup`, `LogicalNot`, `ConstantComparison`, `ProjectParameters`, `RenameParameters`, `RenameItem`, `FilterParameters`, `TrimParameters`, `NormalizeParameters`, `NormalizeTarget`, `ReplaceParameters`, `RegexReplaceParameters`, `DeduplicateParameters`, `SortParameters`, `SortKey` |

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| template values change semantics | typed value-only references, strict validation, stable digests |
| config patch bypasses operational safety | pointer denylist plus ancestor/descendant protection and full revalidation |
| large sort/dedup exceeds RAM | reservations, spill, resource-limit error, wide benchmark fixture |
| connector behavior differs by installation | callable-tool probe, readbacks, no network workaround, manual fallback |
| failure leaks workspace or duplicates receipt | Writer-owned finalization on every terminal path and lifecycle fixtures |
| exact Reader/Writer structures drift after the split | keep complete field tables in their owning specs and compare structs/schemas mechanically |
| batch optimization changes semantics | row-oriented fidelity baseline before performance comparison |
| private acceptance data leaks into fixtures | derive sanitized cases and keep approved source data read-only |
| cleanup script exceeds the V1 catalog or workflow boundary | retain four mapped behaviors only as coverage evidence, record fourteen as `outside_v1`, and do not create their future change here |

## Migration

No runtime migration is performed by this documentation change. Future
implementation proceeds in this order:

1. Freeze traceability and pass strict OpenSpec plus current Go baselines.
2. Align canonical values, API/report contracts, and strict configuration.
3. Split combined foundation code into Reader and Writer registries/packages.
4. Implement the CSV vertical slice and pass its safety/fidelity gate.
5. Add the remaining logical operations one at a time with descriptor tests.
6. Add JSON and Excel Reader/Writer providers independently.
7. Implement all six Skills, workflow instructions, and Google bridge fixtures.
8. Pass the two practical workflows and the scoped four-behavior cleanup coverage, then the versioned Codex suite.
9. Record performance baselines, set thresholds, and run repository-wide gates.

Rollback uses Git reversion of implementation commits. Failed local artifacts
are removed or retained only through `writer.finalize`; sources remain unchanged.

## Open questions

None. Baseline-derived numeric thresholds and the resolved cleanup scope are
implementation evidence with explicit preparatory tasks, not unresolved product
decisions: four mapped behaviors are V1 coverage evidence, while fourteen are
`outside_v1` for a future change that this change MUST NOT create.
