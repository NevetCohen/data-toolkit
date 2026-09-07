## ADDED Requirements

### Requirement: CSV vertical-slice gate
Before activating additional formats, the system SHALL pass an end-to-end CSV
workflow containing a constant filter, at least one further logical operation,
one reopened and schema-validated CSV output, immutable-source verification,
deterministic-output verification, cancellation handling, and compact
run-receipt verification. Inspection is verified independently as a discovery
operation.

#### Scenario: Vertical slice success
- **WHEN** the versioned vertical-slice fixture is executed twice with equivalent output destinations
- **THEN** normalized outputs and explicit assertion results are deterministic,
and both receipts identify the same source snapshot hash and workflow digest

### Requirement: Excel city-filter acceptance
The system SHALL create a new Excel file from the approved `all_final.xlsx`
source by filtering column `ישוב` to the constant `בת ים`, without changing the
source, and SHALL validate the selected sheet, headers, row predicate, output
style, and run receipt.

#### Scenario: City-filter workflow
- **WHEN** the approved local acceptance workflow runs
- **THEN** every output data row has `ישוב = בת ים`, the output is a valid new workbook, and the source hash is unchanged

### Requirement: JSON field-extraction acceptance
The system SHALL extract exactly `שם פרטי`, `שם מלא`, `entityID`, `TZ`, `ישוב`,
and `מייל` from the approved JSON source into a new validated CSV output while
preserving order, Unicode, null distinctions, and exact scalar values.

#### Scenario: JSON-to-CSV workflow
- **WHEN** the approved local extraction workflow runs
- **THEN** the output header and data contain exactly the six fields in declared order and the source hash is unchanged

### Requirement: Cleanup-script scoped coverage acceptance
V1 SHALL remain narrow and SHALL NOT claim full replacement or equivalence with `Invoke-RawMemberCleanup.Streaming.py`.
The sanitized cleanup manifest SHALL preserve exactly four behaviors expressible through the fixed ten logical operations as V1 coverage evidence.
It SHALL record the other fourteen behaviors as `outside_v1` for a future
OpenSpec change that this change MUST NOT create.

#### Scenario: Cleanup V1 scoped coverage
- **WHEN** the four mapped cleanup fixtures are processed through registered V1 operations
- **THEN** their recorded outputs and operation-defined failure behavior are met without importing or invoking the legacy script at runtime, while all fourteen `outside_v1` entries remain excluded from V1 acceptance

### Requirement: Codex reliability suite
The project SHALL provide a versioned suite of at least 20 JSON, CSV, and Excel
tasks covering Hebrew, null, exact numbers, dirty input, nested constant filters,
formatting, clarification, rejection, output validation, and compact run receipts.

#### Scenario: Luna 5.6 target
- **WHEN** GPT-5.6 Luna at medium effort runs the suite
- **THEN** all safety cases pass, at least 90 percent pass fully on the first attempt, and all pass within one structured-error correction

### Requirement: Baseline-derived performance thresholds
The project SHALL record reproducible time and peak-memory baselines on an
approximately 40 MB JSON fixture and a CSV fixture of up to 500,000 rows before
setting numeric thresholds, then SHALL enforce the approved thresholds in
repeatable performance checks.

#### Scenario: Performance regression
- **WHEN** a measured workflow exceeds an approved threshold under the recorded environment and method
- **THEN** the performance gate fails with the measured value and threshold

### Requirement: Wide-fixture memory and spill gate
The project SHALL include a non-sensitive wide fixture with 141 through 146
columns, 33 through 35 percent nulls, a text-heavy distribution, and hundreds
of MB of data or an equivalent row count. A row-oriented baseline SHALL be
recorded first. Columnar batches of 4K, 16K, or 64K rows are permitted only
when measured allocations or garbage collection identify a bottleneck; they are
not a public stream contract. The public model remains `Schema + RowStream +
Value`.

Every baseline and permitted prototype SHALL record wall time, CPU time, peak
RSS, Go heap, total allocations, garbage-collection count/pause, and spill
bytes. It SHALL also verify row count and order, null distinctions, exact
integer/decimal values, normalized output hash, and the first failure locator
on the corresponding error fixture. A prototype SHALL not be accepted merely
because it is faster if any fidelity result differs from the row-oriented
baseline.

Every run component SHALL reserve memory from the run budget, spill before
exhaustion where its operation supports it, and return
`resource_limit_exceeded` in the execution exit class when it cannot proceed.
No output may be published after that error.

#### Scenario: Bounded sort under pressure
- **WHEN** stable `row.sort` or any mode of `row.deduplicate` exceeds its in-memory reservation
- **THEN** it uses bounded spool or spill, honors cancellation and cleanup, or fails with `resource_limit_exceeded` without publication

#### Scenario: Batch prototype fidelity
- **WHEN** an internal 4K, 16K, or 64K batch prototype is benchmarked
- **THEN** its semantic fidelity and first-error location match the row-oriented baseline before its performance result may be considered

### Requirement: Workflow template and CLI fixture gate
Versioned fixtures SHALL cover contiguous variable IDs, multiple variables,
defaults, invalid declared types, missing references, deterministic substitution
and identical digest. They SHALL also cover an allowed config patch, a protected
path, a denylist mutation attempt, precedence, and validation after every patch.

#### Scenario: Protected configuration override
- **WHEN** `--config-set` targets a protected path, its ancestor, or its descendant
- **THEN** validation rejects it before planning and leaves the effective configuration unchanged

### Requirement: Reader, Writer, and logical lifecycle gate
Lifecycle scenarios SHALL cover successful execution, Reader failure, Logical
Engine failure, Writer failure, and cancellation. Each scenario after workspace
creation SHALL observe `writer.finalize`, prescribed cleanup or retention, and
exactly one receipt. Logical fixtures SHALL cover multi-key sort, null ordering,
NFC Unicode ordering, stability, many unique deduplication keys, spill,
cancellation, and memory-budget compliance.

#### Scenario: Writer failure finalization
- **WHEN** staging, reopen validation, an assertion, or publication fails
- **THEN** no partial output is reported and `writer.finalize` emits one terminal report and receipt after cleanup policy is applied

### Requirement: Google Sheets agent-bridge gate
The Codex fixture suite SHALL cover unavailable connector, successful source
export, successful native destination import, destination-folder move, remote
collision, title and style policy, upload failure, local cleanup, and the
prohibition on deleting a Drive file. Google Sheets is supported only through
the Codex XLSX bridge and is not a runtime `FormatID`.

#### Scenario: Missing Google connector
- **WHEN** a Google Sheets bridge run cannot discover the required Drive connector operation
- **THEN** it stops before data processing and offers the documented manual XLSX fallback

### Requirement: Extension and architecture acceptance
External tests SHALL add a data type, logical operation, and file format through
descriptor, implementation, constructor, validation, and explicit registration
without editing the orchestrator or another engine. The root module SHALL reject
imports of `legacy/mvp` and forbidden outward dependencies.

#### Scenario: Test-only extension
- **WHEN** a valid test extension is registered
- **THEN** discovery and dispatch expose it while duplicate or invalid registration is rejected

### Requirement: Repository quality gates
Completion SHALL require formatting, focused contract tests, `go test ./...`,
`go vet ./...`, `go build ./cmd/data-toolkit`, strict OpenSpec validation, stale
scope/config searches, and documentation alignment.

#### Scenario: Final implementation gate
- **WHEN** all implementation tasks are marked complete
- **THEN** every required command passes and no active document advertises TXT, TUI, queries, Google Sheets as a native runtime format, or scaffold-only behavior as supported V1 functionality
